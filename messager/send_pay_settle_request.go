// Copyright 2018-2020 Celer Network

package messager

import (
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils/hashlist"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

func (m *Messager) SendOnePaySettleRequest(
	pay *entity.ConditionalPay,
	payAmt *big.Int,
	reason rpc.PaymentSettleReason,
	logEntry *pem.PayEventMessage) error {
	pays := []*entity.ConditionalPay{pay}
	payAmts := []*big.Int{payAmt}
	_, err := m.SendPaysSettleRequest(pays, payAmts, reason, logEntry)
	return err
}

func (m *Messager) SendPaysSettleRequest(
	pays []*entity.ConditionalPay,
	payAmts []*big.Int, // total amount for all setted payments
	reason rpc.PaymentSettleReason,
	logEntry *pem.PayEventMessage) ([]*entity.ConditionalPay, error) {
	if len(pays) == 0 {
		return nil, fmt.Errorf("Empty settle pay list")
	}
	if len(pays) != len(payAmts) {
		return nil, fmt.Errorf("pay and amt list length not match")
	}
	for i := 1; i < len(pays); i++ {
		logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(ctype.Pay2PayID(pays[i])))
	}

	var seqnum uint64
	var celerMsg *rpc.CelerMsg
	var skippedPays []*entity.ConditionalPay
	var cid ctype.CidType
	var peerTo ctype.Addr
	err := m.dal.Transactional(m.runPaySettleTx, pays, payAmts, reason, &seqnum, &celerMsg, &skippedPays, &cid, &peerTo)
	if err == common.ErrPayNoEgress && len(pays) == 1 && reason == rpc.PaymentSettleReason_PAY_PAID_MAX {
		return nil, m.sendCrossNetPaySettleRequest(pays[0], payAmts[0], logEntry)
	}
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.MsgTo = ctype.Addr2Hex(peerTo)
	if err != nil {
		return skippedPays, err
	}
	logEntry.SeqNums.Out = seqnum
	logEntry.SeqNums.OutBase = celerMsg.GetPaymentSettleRequest().GetBaseSeq()
	log.Debugln("Send payment settle request to", peerTo.Hex(), "reason", reason)
	return skippedPays, m.msgQueue.AddMsg(peerTo, cid, seqnum, celerMsg)
}

func (m *Messager) runPaySettleTx(tx *storage.DALTx, args ...interface{}) error {
	pays := args[0].([]*entity.ConditionalPay)
	payAmts := args[1].([]*big.Int)
	reason := args[2].(rpc.PaymentSettleReason)
	retSeqNum := args[3].(*uint64)
	retCelerMsg := args[4].(**rpc.CelerMsg)
	retSkippedPays := args[5].(*[]*entity.ConditionalPay)
	retCid := args[6].(*ctype.CidType)
	retPeerTo := args[7].(*ctype.Addr)

	cid, egstate, found, err := tx.GetPayEgress(ctype.Pay2PayID(pays[0]))
	if err != nil {
		return fmt.Errorf("GetPayEgress err %w", err)
	}
	if !found {
		return fmt.Errorf("GetPayEgress err %w", common.ErrPayNotFound)
	}
	if cid == ctype.ZeroCid {
		return common.ErrPayNoEgress
	}
	*retCid = cid

	peer, chanState, baseSeq, lastUsedSeq, lastAckedSeq, selfSimplex, found, err := tx.GetChanForSendPaySettleRequest(cid)
	if err != nil {
		return fmt.Errorf("GetChanForSendPaySettleRequest err %w", err)
	}
	if !found {
		return fmt.Errorf("GetChanForSendPaySettleRequest err %w", common.ErrChannelNotFound)
	}
	*retPeerTo = peer

	err = fsm.OnChannelUpdate(cid, chanState)
	if err != nil {
		return fmt.Errorf("OnChannelUpdate err %w", err)
	}

	workingSimplex, err := ledgerview.GetBaseSimplex(tx, cid, selfSimplex, baseSeq, lastAckedSeq)
	if err != nil {
		return fmt.Errorf("GetBaseSimplex err %w", err)
	}

	baseSeq = workingSimplex.SeqNum
	workingSimplex.SeqNum = lastUsedSeq + 1
	lastUsedSeq = workingSimplex.SeqNum

	sendAmt := new(big.Int).SetBytes(workingSimplex.TransferToPeer.Receiver.Amt)
	payAmt := new(big.Int).SetUint64(0)
	for _, amt := range payAmts {
		payAmt = payAmt.Add(payAmt, amt)
	}
	workingSimplex.TransferToPeer.Receiver.Amt = sendAmt.Add(sendAmt, payAmt).Bytes()
	totalPendingAmt := new(big.Int).SetBytes(workingSimplex.TotalPendingAmount)

	var settledPays []*rpc.SettledPayment
	var skippedPays []*entity.ConditionalPay
	paid := !(payAmt.Cmp(new(big.Int).SetUint64(0)) == 0)
	for i := 0; i < len(pays); i++ {
		pay := pays[i]
		payID := ctype.Pay2PayID(pay)
		// TODO: inefficient O(N^2) time complexity, should delete all payIds in one pass
		workingSimplex.PendingPayIds.PayIds, err =
			hashlist.DeleteHash(workingSimplex.PendingPayIds.PayIds, payID[:])
		if err != nil {
			if !paid {
				log.Warnln("delete pay hash failed:", err, payID.Hex(), "paid:", paid)
				skippedPays = append(skippedPays, pay)
				continue
			}
			return fmt.Errorf("hashlist DeleteHash %x err %w", payID, err)
		}
		amt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
		totalPendingAmt = totalPendingAmt.Sub(totalPendingAmt, amt)
		// payment state machine
		if paid {
			err = fsm.OnPayEgressOneSigPaid(tx, payID, egstate)
		} else {
			err = fsm.OnPayEgressOneSigCanceled(tx, payID, egstate)
		}
		if err != nil {
			if !paid {
				log.Warnln(err, "paid:", paid)
				skippedPays = append(skippedPays, pay)
				continue
			}
			return fmt.Errorf("pay %x fsm err %w, paid %t", payID, err, paid)
		}
		settledPay := &rpc.SettledPayment{
			SettledPayId: payID[:],
			Reason:       reason,
			Amount:       payAmts[i].Bytes(),
		}
		settledPays = append(settledPays, settledPay)
	}
	if len(settledPays) == 0 {
		return fmt.Errorf("invalid payment settle request")
	}
	workingSimplex.TotalPendingAmount = totalPendingAmt.Bytes()

	var workingSimplexState rpc.SignedSimplexState
	workingSimplexState.SimplexState, err = proto.Marshal(workingSimplex)
	if err != nil {
		return fmt.Errorf("marshal simplex state err %w", err)
	}
	workingSimplexState.SigOfPeerFrom, err = m.signer.SignEthMessage(workingSimplexState.SimplexState)
	if err != nil {
		return fmt.Errorf("sign simplex state err %w", err)
	}

	request := &rpc.PaymentSettleRequest{
		StateOnlyPeerFromSig: &workingSimplexState,
		SettledPays:          settledPays,
		BaseSeq:              baseSeq,
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_PaymentSettleRequest{
			PaymentSettleRequest: request,
		},
	}
	*retSeqNum = workingSimplex.SeqNum
	*retCelerMsg = celerMsg
	*retSkippedPays = skippedPays

	err = tx.InsertChanMessage(cid, *retSeqNum, celerMsg)
	if err != nil {
		return fmt.Errorf("InsertChanMessage err %w", err)
	}

	err = tx.UpdateChanForSendRequest(cid, lastUsedSeq, lastUsedSeq)
	if err != nil {
		return fmt.Errorf("UpdateChanForSendRequest err %w", err)
	}
	return nil
}

func (m *Messager) ForwardPaySettleRequestMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	request := msg.GetPaymentSettleRequest()
	var payID ctype.PayIDType
	if len(request.GetSettledPays()) > 0 {
		if len(request.SettledPays) > 1 {
			return fmt.Errorf("batched pay settle request forwarding not supported yet")
		}
		payID = ctype.Bytes2PayID(request.SettledPays[0].SettledPayId)
		if request.SettledPays[0].GetReason() != rpc.PaymentSettleReason_PAY_PAID_MAX {
			return fmt.Errorf("can only forward max paid settle request")
		}
	} else {
		return fmt.Errorf("empty settled pays in paymentSettleRequest")
	}
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.SettleReason = request.SettledPays[0].GetReason()
	pay, _, found, err := m.dal.GetPayment(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return fmt.Errorf("GetPayment err %w", common.ErrPayNotFound)
	}
	amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	return m.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, logEntry)
}

func (m *Messager) sendCrossNetPaySettleRequest(
	pay *entity.ConditionalPay,
	payAmt *big.Int,
	logEntry *pem.PayEventMessage) error {
	payID := ctype.Pay2PayID(pay)
	originalPayID, state, bridgeAddr, found, err := m.dal.GetCrossNetInfoByPayID(payID)
	if err != nil {
		return fmt.Errorf("GetCrossNetInfo err %w", err)
	}
	if !found {
		return common.ErrPayNoEgress
	}
	if state != enums.CrossNetPay_EGRESS {
		return fmt.Errorf("invalid cross net pay state %d", state)
	}
	settledPay := &rpc.SettledPayment{
		SettledPayId:  payID.Bytes(),
		Reason:        rpc.PaymentSettleReason_PAY_PAID_MAX,
		OriginalPayId: originalPayID.Bytes(),
	}
	request := &rpc.PaymentSettleRequest{}
	request.SettledPays = append(request.SettledPays, settledPay)
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_PaymentSettleRequest{
			PaymentSettleRequest: request,
		},
	}

	return m.streamWriter.WriteCelerMsg(bridgeAddr, celerMsg)
}
