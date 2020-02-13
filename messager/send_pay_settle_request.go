// Copyright 2018-2019 Celer Network

package messager

import (
	"errors"
	"math/big"

	"github.com/celer-network/goCeler-oss/common"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils/hashlist"
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
		return nil, errors.New("Empty settle pay list")
	}
	if len(pays) != len(payAmts) {
		return nil, errors.New("pay and amt list length not match")
	}

	cid, _, _, err := m.dal.GetPayEgressState(ctype.Pay2PayID(pays[0]))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	logEntry.ToCid = ctype.Cid2Hex(cid)
	for i := 1; i < len(pays); i++ {
		logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(ctype.Pay2PayID(pays[i])))
	}
	peerTo, err := m.dal.GetPeer(cid)
	if err != nil {
		log.Errorln(err, cid.Hex())
		return nil, err
	}
	logEntry.MsgTo = peerTo
	log.Debugln("Send payment settle request to", peerTo, "reason", reason)

	var seqnum uint64
	var celerMsg *rpc.CelerMsg
	var skippedPays []*entity.ConditionalPay
	err = m.dal.Transactional(m.runPaySettleTx, cid, pays, payAmts, reason, &seqnum, &celerMsg, &skippedPays)
	if err != nil {
		return skippedPays, err
	}
	logEntry.SeqNums.Out = seqnum
	logEntry.SeqNums.OutBase = celerMsg.GetPaymentSettleRequest().GetBaseSeq()

	return skippedPays, m.msgQueue.AddMsg(peerTo, cid, seqnum, celerMsg)
}

func (m *Messager) runPaySettleTx(tx *storage.DALTx, args ...interface{}) error {
	cidNextHop := args[0].(ctype.CidType)
	pays := args[1].([]*entity.ConditionalPay)
	payAmts := args[2].([]*big.Int)
	reason := args[3].(rpc.PaymentSettleReason)
	retSeqNum := args[4].(*uint64)
	retCelerMsg := args[5].(**rpc.CelerMsg)
	retSkippedPays := args[6].(*[]*entity.ConditionalPay)

	// channel state machine
	err := fsm.OnPscUpdateSimplex(tx, cidNextHop)
	if err != nil {
		log.Error(err)
		return err
	}

	workingSimplex, seqNums, err :=
		ledgerview.GetBaseSimplexChannel(tx, cidNextHop, m.nodeConfig.GetOnChainAddr())
	if err != nil {
		log.Errorln(err, cidNextHop.Hex())
		return err
	}

	baseSeq := workingSimplex.SeqNum
	workingSimplex.SeqNum = seqNums.LastUsed + 1
	seqNums.LastUsed = workingSimplex.SeqNum
	seqNums.Base = seqNums.LastUsed

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
			log.Errorln("delete pay hash failed:", err, payID.Hex(), "paid:", paid)
			if !paid {
				skippedPays = append(skippedPays, pay)
				continue
			}
			return err
		}
		amt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
		totalPendingAmt = totalPendingAmt.Sub(totalPendingAmt, amt)
		// payment state machine
		var cid ctype.CidType
		if paid {
			cid, _, err = fsm.OnPayEgressOneSigPaid(tx, payID)
		} else {
			cid, _, err = fsm.OnPayEgressOneSigCanceled(tx, payID)
		}
		if err != nil {
			log.Errorln(err, "paid:", paid)
			if !paid {
				skippedPays = append(skippedPays, pay)
				continue
			}
			return err
		}
		if cid != cidNextHop {
			return errors.New("cannot batch requests for different egress cids")
		}
		settledPay := &rpc.SettledPayment{
			SettledPayId: payID[:],
			Reason:       reason,
			Amount:       payAmts[i].Bytes(),
		}
		settledPays = append(settledPays, settledPay)
	}
	if len(settledPays) == 0 {
		return errors.New("invalid payment settle request")
	}
	workingSimplex.TotalPendingAmount = totalPendingAmt.Bytes()

	var workingSimplexState rpc.SignedSimplexState
	workingSimplexState.SimplexState, _ = proto.Marshal(workingSimplex)
	workingSimplexState.SigOfPeerFrom, _ = m.signer.Sign(workingSimplexState.SimplexState)

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

	err = tx.PutChannelSeqNums(cidNextHop, seqNums)
	if err != nil {
		log.Error(err)
		return err
	}
	return tx.PutChannelMessage(cidNextHop, *retSeqNum, celerMsg)
}

func (m *Messager) ForwardPaySettleRequestMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	request := msg.GetPaymentSettleRequest()
	var payID ctype.PayIDType
	if len(request.GetSettledPays()) > 0 {
		if len(request.SettledPays) > 1 {
			return errors.New("batched pay settle request forwarding not supported yet")
		}
		payID = ctype.Bytes2PayID(request.SettledPays[0].SettledPayId)
		if request.SettledPays[0].GetReason() != rpc.PaymentSettleReason_PAY_PAID_MAX {
			return errors.New("can only forward max paid settle request")
		}
	} else {
		return errors.New("empty settled pays in paymentSettleRequest")
	}
	pay, _, err := m.dal.GetConditionalPay(payID)
	logEntry.PayId = ctype.PayID2Hex(ctype.Pay2PayID(pay))
	if err != nil {
		return err
	}
	amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	return m.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, logEntry)
}
