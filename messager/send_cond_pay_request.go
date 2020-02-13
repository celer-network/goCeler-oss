// Copyright 2018-2019 Celer Network

package messager

import (
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/rtconfig"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/celer-network/goCeler-oss/utils/hashlist"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

func (m *Messager) SendCondPayRequest(pay *entity.ConditionalPay, note *any.Any, logEntry *pem.PayEventMessage) error {
	payBytes, err := proto.Marshal(pay)
	if err != nil {
		return err
	}

	token := pay.GetTransferFunc().GetMaxTransfer().GetToken()
	cid, peer, err := m.channelRouter.LookupNextChannelOnToken(
		ctype.Bytes2Hex(pay.GetDest()), utils.GetTokenAddrStr(token))
	if err != nil {
		log.Error(err)
		return err
	}
	if cid == ctype.ZeroCid {
		log.Errorln(common.ErrRouteNotFound, "dst", ctype.Bytes2Hex(pay.GetDest()), "src", ctype.Bytes2Hex(pay.GetSrc()))
		return common.ErrRouteNotFound
	}

	directPay := m.IsDirectPay(pay, peer)
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: &rpc.CondPayRequest{
				CondPay:   payBytes,
				Note:      note,
				DirectPay: directPay,
			},
		},
	}

	logEntry.MsgTo = peer
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.DirectPay = directPay

	if isLocalPeer, errf := m.serverForwarder(peer, celerMsg); errf != nil {
		log.Errorln(errf)
		err = errf
	} else if isLocalPeer {
		err = m.sendCondPayRequest(pay, note, cid, peer, logEntry)
	}
	return err
}

func (m *Messager) ForwardCondPayRequestMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	var pay entity.ConditionalPay
	err := proto.Unmarshal(msg.GetCondPayRequest().GetCondPay(), &pay)
	if err != nil {
		log.Error(err)
		return err
	}
	logEntry.PayId = ctype.PayID2Hex(ctype.Pay2PayID(&pay))
	logEntry.Dst = ctype.Bytes2Hex(pay.GetDest())
	cid, peer, err := m.channelRouter.LookupNextChannelOnToken(
		ctype.Bytes2Hex(pay.Dest), utils.GetTokenAddrStr(pay.TransferFunc.MaxTransfer.Token))
	if err != nil {
		log.Error(err)
		return err
	}

	directPay := m.IsDirectPay(&pay, peer)
	logEntry.MsgTo = peer
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.DirectPay = directPay

	err = m.sendCondPayRequest(&pay, msg.GetCondPayRequest().GetNote(), cid, peer, logEntry)
	return err
}

func (m *Messager) sendCondPayRequest(
	pay *entity.ConditionalPay, note *any.Any, cidNextHop ctype.CidType, peerTo string, logEntry *pem.PayEventMessage) error {

	payID := ctype.Pay2PayID(pay)
	directPay := m.IsDirectPay(pay, peerTo)
	log.Debugf("Send pay request %x, src %x, dst %x, direct %t", payID, pay.GetSrc(), pay.GetDest(), directPay)
	payBytes, err := proto.Marshal(pay)
	if err != nil {
		return err
	}

	// verify pay destination
	if ctype.Bytes2Addr(pay.GetDest()) == ctype.Hex2Addr(m.nodeConfig.GetOnChainAddr()) {
		log.Errorln(common.ErrInvalidPayDst, utils.PrintConditionalPay(pay))
		return common.ErrInvalidPayDst
	}

	// verify payment deadline is within limit
	if pay.GetResolveDeadline() >
		m.monitorService.GetCurrentBlockNumber().Uint64()+rtconfig.GetMaxPaymentTimeout() {
		log.Errorln(common.ErrInvalidPayDeadline,
			pay.GetResolveDeadline(), m.monitorService.GetCurrentBlockNumber().Uint64())
		return common.ErrInvalidPayDeadline
	}

	var seqnum uint64
	var celerMsg *rpc.CelerMsg

	err = m.dal.Transactional(m.runCondPayTx, cidNextHop, payID, pay, payBytes, note, directPay, &seqnum, &celerMsg)
	if err != nil {
		return err
	}
	err = m.msgQueue.AddMsg(peerTo, cidNextHop, seqnum, celerMsg)
	if err != nil {
		// this can only happen when peer got disconnected after sendCondPayRequest() is called
		// not return AddMsg, as db has been updated and rolling back is complicated
		// the msg will be sent out when the peer reconnected, though the msg itself might be outdated.
		log.Warnln(err, cidNextHop.Hex())
	}
	logEntry.SeqNums.Out = seqnum
	logEntry.SeqNums.OutBase = celerMsg.GetCondPayRequest().GetBaseSeq()

	return nil
}

func (m *Messager) runCondPayTx(tx *storage.DALTx, args ...interface{}) error {
	cidNextHop := args[0].(ctype.CidType)
	payID := args[1].(ctype.PayIDType)
	condPay := args[2].(*entity.ConditionalPay)
	payBytes := args[3].([]byte)
	note := args[4].(*any.Any)
	directPay := args[5].(bool)
	retSeqNum := args[6].(*uint64)
	retCelerMgr := args[7].(**rpc.CelerMsg)

	// channel state machine
	err := fsm.OnPscUpdateSimplex(tx, cidNextHop)
	if err != nil {
		log.Error(err)
		return err
	}

	// payment state machine
	if directPay {
		_, _, err = fsm.OnPayEgressDirectOneSigPaid(tx, payID, cidNextHop)
	} else {
		_, _, err = fsm.OnPayEgressOneSigPending(tx, payID, cidNextHop)
	}
	if err != nil {
		log.Error(err)
		return err
	}

	blkNum := m.monitorService.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalanceTx(tx, cidNextHop, m.nodeConfig.GetOnChainAddr(), blkNum)
	if err != nil {
		log.Errorln(err, "unabled to find balance for cidNextHop", cidNextHop.Hex())
		return err
	}
	sendAmt := new(big.Int).SetBytes(condPay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	if sendAmt.Cmp(balance.MyFree) == 1 {
		// No enough sending capacity to send the new pay.
		log.Errorln("Not enough balance to send on", cidNextHop.Hex(), "need:", sendAmt, "have:", balance.MyFree)
		return common.ErrNoEnoughBalance
	}

	workingSimplex, seqNums, err :=
		ledgerview.GetBaseSimplexChannel(tx, cidNextHop, m.nodeConfig.GetOnChainAddr())
	if err != nil {
		log.Errorln("GetSimplexPaymentChannel", err, cidNextHop.Hex())
		return err
	}

	baseSeq := workingSimplex.SeqNum
	workingSimplex.SeqNum = seqNums.LastUsed + 1
	seqNums.LastUsed = workingSimplex.SeqNum
	seqNums.Base = seqNums.LastUsed

	if directPay {
		amt := new(big.Int).SetBytes(workingSimplex.TransferToPeer.Receiver.Amt)
		workingSimplex.TransferToPeer.Receiver.Amt = amt.Add(amt, sendAmt).Bytes()
	} else {
		if hashlist.Exist(workingSimplex.PendingPayIds.PayIds, payID[:]) {
			log.Errorln(common.ErrPayAlreadyPending, payID.Hex())
			return common.ErrPayAlreadyPending
		}
		workingSimplex.PendingPayIds.PayIds = append(workingSimplex.PendingPayIds.PayIds, payID[:])

		// verify number of pending pays is within limit
		if len(workingSimplex.PendingPayIds.PayIds) > int(rtconfig.GetMaxNumPendingPays()) {
			log.Errorln(common.ErrTooManyPendingPays, len(workingSimplex.PendingPayIds.PayIds))
			return common.ErrTooManyPendingPays
		}

		totalPendingAmt := new(big.Int).SetBytes(workingSimplex.TotalPendingAmount)
		workingSimplex.TotalPendingAmount = totalPendingAmt.Add(totalPendingAmt, sendAmt).Bytes()

		if condPay.GetResolveDeadline() > workingSimplex.GetLastPayResolveDeadline() {
			workingSimplex.LastPayResolveDeadline = condPay.ResolveDeadline
		}
	}

	var workingSimplexState rpc.SignedSimplexState
	workingSimplexState.SimplexState, _ = proto.Marshal(workingSimplex)
	workingSimplexState.SigOfPeerFrom, _ = m.signer.Sign(workingSimplexState.SimplexState)

	request := &rpc.CondPayRequest{
		CondPay:              payBytes,
		StateOnlyPeerFromSig: &workingSimplexState,
		Note:                 note,
		BaseSeq:              baseSeq,
		DirectPay:            directPay,
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: request,
		},
	}
	*retSeqNum = workingSimplex.SeqNum
	*retCelerMgr = celerMsg

	err = tx.PutChannelSeqNums(cidNextHop, seqNums)
	if err != nil {
		log.Error(err)
		return err
	}
	return tx.PutChannelMessage(cidNextHop, *retSeqNum, celerMsg)
}
