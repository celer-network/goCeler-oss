// Copyright 2018-2020 Celer Network

package messager

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/delegate"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/hashlist"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func (m *Messager) SendCondPayRequest(payBytes []byte, note *any.Any, xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) error {
	pay, cid, peer, celerMsg, directPay, err := m.getPayNextHopAndCelerMsg(payBytes, note, xnet, logEntry)
	if err != nil {
		return err
	}

	isLocalPeer, err := m.serverForwarder(peer, true, celerMsg)
	if !isLocalPeer {
		if err == nil {
			return nil // handled by another server
		} else if !directPay {
			return err // must retry when peer reconnects
		}
	}
	// It's either meant to a local peer or it's a failed forwarding
	// of a direct-pay.  In both cases handle it locally which puts
	// the message in the queue for delivery (now or later).
	return m.sendCondPayRequest(payBytes, pay, note, cid, peer, xnet, logEntry)
}

func (m *Messager) ForwardCondPayRequest(
	payBytes []byte, note *any.Any, delegable bool, xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) (ctype.Addr, error) {
	pay, cid, peer, celerMsg, _, err := m.getPayNextHopAndCelerMsg(payBytes, note, xnet, logEntry)
	if err != nil {
		return peer, err
	}
	// if delegable, do not retry on multiserver forward failure
	isLocalPeer, err := m.serverForwarder(peer, !delegable, celerMsg)
	if err != nil {
		return peer, err
	}
	if isLocalPeer {
		return peer, m.sendCondPayRequest(payBytes, pay, note, cid, peer, xnet, logEntry)
	}

	return peer, nil
}

func (m *Messager) ForwardCondPayRequestMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	payBytes := msg.GetCondPayRequest().GetCondPay()
	xnet := msg.GetCondPayRequest().GetCrossNet()

	pay, cid, peer, _, err := m.getPayNextHop(payBytes, xnet, logEntry)
	if err != nil {
		return err
	}

	logEntry.PayId = ctype.PayID2Hex(ctype.Pay2PayID(pay))
	logEntry.Dst = ctype.Bytes2Hex(pay.GetDest())

	return m.sendCondPayRequest(payBytes, pay, msg.GetCondPayRequest().GetNote(), cid, peer, xnet, logEntry)
}

func (m *Messager) getPayNextHop(payBytes []byte, xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) (
	*entity.ConditionalPay, ctype.CidType, ctype.Addr, bool, error) {

	var pay entity.ConditionalPay
	err := proto.Unmarshal(payBytes, &pay)
	if err != nil {
		return nil, ctype.ZeroCid, ctype.ZeroAddr, false, err
	}
	dst := ctype.Bytes2Addr(pay.GetDest())
	token := pay.GetTransferFunc().GetMaxTransfer().GetToken()
	var cid ctype.CidType
	var peer ctype.Addr

	if xnet.GetDstNetId() != 0 {
		logEntry.Xnet.SrcNetId = xnet.GetSrcNetId()
		logEntry.Xnet.DstNetId = xnet.GetDstNetId()
		myNetId, err2 := m.dal.GetNetId()
		if err2 != nil {
			return nil, ctype.ZeroCid, ctype.ZeroAddr, false, fmt.Errorf("GetNetId err: %w", err2)
		}
		if myNetId != xnet.GetDstNetId() { // destination is in a remote network
			if myNetId == xnet.GetBridgeNetId() && m.nodeConfig.GetOnChainAddr() != ctype.Bytes2Addr(xnet.GetBridgeAddr()) {
				// local bridge known but not me, route to the local bridge
				dst = ctype.Bytes2Addr(xnet.GetBridgeAddr())
			} else {
				bridgeAddr, bridgeNetId, err2 := m.getBridgeRouting(xnet.GetDstNetId())
				if err2 != nil {
					return nil, ctype.ZeroCid, ctype.ZeroAddr, false, err2
				}
				if bridgeNetId == myNetId {
					// route to the local bridge
					dst = bridgeAddr
				} else {
					// forward to the remote bridge
					peer = bridgeAddr
					xnet.Crossing = true
				}
				xnet.BridgeAddr = bridgeAddr.Bytes()
				xnet.BridgeNetId = bridgeNetId
			}
			logEntry.Xnet.ToBridgeAddr = ctype.Bytes2Hex(xnet.GetBridgeAddr())
			logEntry.Xnet.ToBridgeNetId = xnet.GetBridgeNetId()
		}
	}

	if !xnet.GetCrossing() {
		cid, peer, err = m.routeForwarder.LookupNextChannelOnToken(dst, utils.GetTokenAddr(token))
		if err != nil {
			return nil, ctype.ZeroCid, ctype.ZeroAddr, false, err
		}
	}

	directPay := m.IsDirectPay(&pay, peer, xnet.GetDstNetId())
	logEntry.MsgTo = ctype.Addr2Hex(peer)
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.DirectPay = directPay

	return &pay, cid, peer, directPay, nil
}

func (m *Messager) getPayNextHopAndCelerMsg(payBytes []byte, note *any.Any, xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) (
	*entity.ConditionalPay, ctype.CidType, ctype.Addr, *rpc.CelerMsg, bool, error) {
	pay, cid, peer, directPay, err := m.getPayNextHop(payBytes, xnet, logEntry)
	if err != nil {
		return nil, ctype.ZeroCid, ctype.ZeroAddr, nil, false, err
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: &rpc.CondPayRequest{
				CondPay:   payBytes,
				Note:      note,
				DirectPay: directPay,
				CrossNet:  xnet,
			},
		},
	}
	return pay, cid, peer, celerMsg, directPay, nil
}

func (m *Messager) sendCondPayRequest(
	payBytes []byte, pay *entity.ConditionalPay, note *any.Any,
	cid ctype.CidType, peerTo ctype.Addr,
	xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) error {

	payID := ctype.Pay2PayID(pay)
	if xnet.GetCrossing() {
		return m.sendCrossNetPay(payID, payBytes, pay, note, peerTo, xnet, logEntry)
	}
	directPay := m.IsDirectPay(pay, peerTo, xnet.GetDstNetId())
	log.Debugf("Send pay request %x, src %x, dst %x, direct %t", payID, pay.GetSrc(), pay.GetDest(), directPay)

	// verify pay destination
	if ctype.Bytes2Addr(pay.GetDest()) == m.nodeConfig.GetOnChainAddr() {
		return common.ErrInvalidPayDst
	}

	// verify payment deadline is within limit
	blknum := m.monitorService.GetCurrentBlockNumber().Uint64()
	if pay.GetResolveDeadline() > blknum+rtconfig.GetMaxPaymentTimeout() {
		return fmt.Errorf("%w, deadline %d current %d", common.ErrInvalidPayDeadline, pay.GetResolveDeadline(), blknum)
	}

	var seqnum uint64
	var celerMsg *rpc.CelerMsg
	err := m.dal.Transactional(m.runCondPayTx, cid, payID, pay, payBytes, note, directPay, xnet, &seqnum, &celerMsg)
	if err != nil {
		return err
	}
	err = m.msgQueue.AddMsg(peerTo, cid, seqnum, celerMsg)
	if err != nil {
		// This can only happen when peer got disconnected after sendCondPayRequest() is called.
		// We do not return AddMsg error, as db has been updated and rolling back is complicated.
		// The msg will be sent out when the peer reconnected, though the pay itself might be expired.
		log.Warnln(err, cid.Hex())
	}
	logEntry.SeqNums.Out = seqnum
	logEntry.SeqNums.OutBase = celerMsg.GetCondPayRequest().GetBaseSeq()

	// if I am the src of a cross network payment
	if xnet.GetDstNetId() != 0 && xnet.GetSrcNetId() == xnet.GetBridgeNetId() &&
		ctype.Bytes2Addr(pay.GetSrc()) == m.nodeConfig.GetOnChainAddr() {
		return m.dal.InsertCrossNetPay(
			payID, payID, nil, enums.CrossNetPay_SRC, xnet.GetSrcNetId(), xnet.GetDstNetId(),
			ctype.Bytes2Addr(xnet.GetBridgeAddr()), xnet.GetBridgeNetId())
	}
	return nil
}

func (m *Messager) runCondPayTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	payID := args[1].(ctype.PayIDType)
	pay := args[2].(*entity.ConditionalPay)
	payBytes := args[3].([]byte)
	note := args[4].(*any.Any)
	directPay := args[5].(bool)
	xnet := args[6].(*rpc.CrossNetPay)
	retSeqNum := args[7].(*uint64)
	retCelerMgr := args[8].(**rpc.CelerMsg)

	peer, chanState, onChainBalance, baseSeq, lastUsedSeq, lastAckedSeq,
		selfSimplex, peerSimplex, found, err := tx.GetChanForSendCondPayRequest(cid)
	if err != nil {
		return fmt.Errorf("GetChanForSendCondPayRequest err %w", err)
	}
	if !found {
		return common.ErrChannelNotFound
	}
	err = fsm.OnChannelUpdate(cid, chanState)
	if err != nil {
		return fmt.Errorf("OnChannelUpdate err %w", err)
	}

	workingSimplex, err := ledgerview.GetBaseSimplex(tx, cid, selfSimplex, baseSeq, lastAckedSeq)
	if err != nil {
		return fmt.Errorf("GetBaseSimplex err %w", err)
	}

	blkNum := m.monitorService.GetCurrentBlockNumber().Uint64()
	balance := ledgerview.ComputeBalance(
		workingSimplex, peerSimplex, onChainBalance, m.nodeConfig.GetOnChainAddr(), peer, blkNum)
	sendAmt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	// OSP refill if free balance is below threshold
	if m.isOSP && chanState == enums.ChanState_OPENED {
		tokenAddr := utils.GetTokenAddrStr(pay.TransferFunc.MaxTransfer.Token)
		refillThreshold := rtconfig.GetRefillThreshold(tokenAddr)
		newMyFree := new(big.Int).Sub(balance.MyFree, sendAmt)
		if refillThreshold.Cmp(newMyFree) == 1 {
			warnMsg := fmt.Sprintf("cid %x balance %s below refill threshold %s", cid, newMyFree, refillThreshold)
			refillAmount, maxWait := rtconfig.GetRefillAmountAndMaxWait(tokenAddr)
			depositID, err2 := m.depositProcessor.RequestRefillTx(tx, cid, refillAmount, maxWait)
			if err2 == nil {
				log.Warnln(warnMsg, "triggered by pay", ctype.PayID2Hex(payID), "refill", refillAmount, "job ID:", depositID)
			} else if errors.Is(err2, common.ErrPendingRefill) {
				log.Warnln(warnMsg, "triggered by pay", ctype.PayID2Hex(payID), "refill pending")
			} else {
				return fmt.Errorf("refill err %w", err2)
			}
		}
	}
	if sendAmt.Cmp(balance.MyFree) == 1 {
		// No enough sending capacity to send the new pay.
		return fmt.Errorf("%w, need %s free %s", common.ErrNoEnoughBalance, sendAmt.String(), balance.MyFree.String())
	}

	baseSeq = workingSimplex.SeqNum
	workingSimplex.SeqNum = lastUsedSeq + 1
	lastUsedSeq = workingSimplex.SeqNum

	if directPay {
		amt := new(big.Int).SetBytes(workingSimplex.TransferToPeer.Receiver.Amt)
		workingSimplex.TransferToPeer.Receiver.Amt = amt.Add(amt, sendAmt).Bytes()
	} else {
		if hashlist.Exist(workingSimplex.PendingPayIds.PayIds, payID[:]) {
			return common.ErrPayAlreadyPending
		}
		workingSimplex.PendingPayIds.PayIds = append(workingSimplex.PendingPayIds.PayIds, payID[:])

		// verify number of pending pays is within limit
		if len(workingSimplex.PendingPayIds.PayIds) > int(rtconfig.GetMaxNumPendingPays()) {
			return fmt.Errorf("%w: %d", common.ErrTooManyPendingPays, len(workingSimplex.PendingPayIds.PayIds))
		}

		totalPendingAmt := new(big.Int).SetBytes(workingSimplex.TotalPendingAmount)
		workingSimplex.TotalPendingAmount = totalPendingAmt.Add(totalPendingAmt, sendAmt).Bytes()

		if pay.GetResolveDeadline() > workingSimplex.GetLastPayResolveDeadline() {
			workingSimplex.LastPayResolveDeadline = pay.ResolveDeadline
		}
	}

	var workingSimplexState rpc.SignedSimplexState
	workingSimplexState.SimplexState, err = proto.Marshal(workingSimplex)
	if err != nil {
		return fmt.Errorf("marshal simplex state err %w", err)
	}
	workingSimplexState.SigOfPeerFrom, err = m.signer.SignEthMessage(workingSimplexState.SimplexState)
	if err != nil {
		return fmt.Errorf("sign simplex state err %w", err)
	}

	request := &rpc.CondPayRequest{
		CondPay:              payBytes,
		StateOnlyPeerFromSig: &workingSimplexState,
		Note:                 note,
		BaseSeq:              baseSeq,
		DirectPay:            directPay,
		CrossNet:             xnet,
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: request,
		},
	}
	*retSeqNum = workingSimplex.SeqNum
	*retCelerMgr = celerMsg

	err = tx.InsertChanMessage(cid, *retSeqNum, celerMsg)
	if err != nil {
		return fmt.Errorf("InsertChanMessage err %w", err)
	}

	err = tx.UpdateChanForSendRequest(cid, lastUsedSeq, lastUsedSeq)
	if err != nil {
		return fmt.Errorf("UpdateChanForSendRequest err %w", err)
	}

	err = m.updateDelegatedPay(tx, payID, pay, note)
	if err != nil {
		return fmt.Errorf("updateDelegatedPay err %w", err)
	}

	found, egstate, err := fsm.OnCondPayRequestSent(tx, payID, cid, directPay)
	if err != nil {
		return fmt.Errorf("OnCondPayRequestSent err %w", err)
	}
	if !found {
		err = tx.InsertPayment(payID, payBytes, pay, note, ctype.ZeroCid, enums.PayState_NULL, cid, egstate)
		if err != nil {
			return fmt.Errorf("InsertPayment err %w", err)
		}
	}

	return nil
}

func (m *Messager) updateDelegatedPay(tx *storage.DALTx, payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any) error {
	dnote := &delegate.PayOriginNote{}
	if ptypes.Is(note, dnote) && ctype.Bytes2Addr(pay.GetSrc()) == m.nodeConfig.GetOnChainAddr() {
		err := ptypes.UnmarshalAny(note, dnote)
		if err != nil {
			return fmt.Errorf("UnmarshalAny err %w", err)
		}
		if !dnote.GetIsRefund() {
			// TODO: make this API take an array of payIDs and update all in a single SQL statement (batch)
			for _, op := range dnote.GetOriginalPays() {
				pid := ctype.Bytes2PayID(op.GetPayId())
				err = tx.UpdateSendingDelegatedPay(pid, payID)
				if err != nil {
					return fmt.Errorf("sending delegated pay error %x: %w", pid, err)
				}
			}
		}
	}
	return nil
}

func (m *Messager) sendCrossNetPay(
	payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any,
	peerTo ctype.Addr, xnet *rpc.CrossNetPay, logEntry *pem.PayEventMessage) error {
	if !xnet.GetCrossing() {
		return fmt.Errorf("not crossing net payment")
	}
	logEntry.Xnet.State = pem.CrossNetPayState_XNET_EGRESS
	originalPayId := ctype.PayBytes2PayID(xnet.GetOriginalPay())
	if payID != originalPayId {
		logEntry.Xnet.OriginalPayId = ctype.PayID2Hex(originalPayId)
	}

	_, found, err := m.dal.GetPayIngressChannel(payID)
	if err != nil {
		return fmt.Errorf("GetPayIngressChannel err: %w", err)
	}
	if !found {
		return fmt.Errorf("no ingress for crossing net payment")
	}

	err = m.dal.InsertCrossNetPay(
		payID, originalPayId, xnet.GetOriginalPay(), enums.CrossNetPay_EGRESS,
		xnet.GetSrcNetId(), xnet.GetDstNetId(), ctype.Bytes2Addr(xnet.GetBridgeAddr()), xnet.GetBridgeNetId())
	if err != nil {
		return fmt.Errorf("InsertCrossNetPay err: %w", err)
	}

	xnet.Timeout = pay.ResolveDeadline - m.monitorService.GetCurrentBlockNumber().Uint64()
	request := &rpc.CondPayRequest{
		CondPay:  payBytes,
		Note:     note,
		CrossNet: xnet,
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: request,
		},
	}

	return m.streamWriter.WriteCelerMsg(peerTo, celerMsg)
}

func (m *Messager) getBridgeRouting(destNetId uint64) (ctype.Addr, uint64, error) {
	bridgeAddr, bridgeNetId, found, err := m.dal.GetBridgeRouting(destNetId)
	if err != nil {
		return ctype.ZeroAddr, 0, fmt.Errorf("GetBridgeRouting err: %w", err)
	}
	if !found {
		return ctype.ZeroAddr, 0, fmt.Errorf("BridgeRouting not found")
	}
	return bridgeAddr, bridgeNetId, nil
}
