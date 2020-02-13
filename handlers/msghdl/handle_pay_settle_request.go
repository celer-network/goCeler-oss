// Copyright 2018-2019 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/celer-network/goCeler-oss/utils/hashlist"
	"github.com/golang/protobuf/proto"
)

type settledPayInfo struct {
	req       *rpc.SettledPayment
	pay       *entity.ConditionalPay
	routeLoop bool
}

func (h *CelerMsgHandler) HandlePaySettleRequest(frame *common.MsgFrame) error {
	peerFrom := ctype.Addr2Hex(frame.PeerAddr)
	log.Debugln("Received payment settle request from", peerFrom)
	request := frame.Message.GetPaymentSettleRequest()

	if len(request.GetSettledPays()) == 0 {
		return errors.New("empty SettledPays list")
	}

	payInfos, err := h.paySettleRequestInbound(request, peerFrom, frame.LogEntry)
	if err != nil {
		return err
	}

	err = h.paySettleRequestOutbound(payInfos, frame.LogEntry)
	if err != nil {
		return err
	}

	return nil
}

func (h *CelerMsgHandler) paySettleRequestInbound(
	request *rpc.PaymentSettleRequest, peerFrom string, logEntry *pem.PayEventMessage) ([]*settledPayInfo, error) {

	// Sign the state in advance, verify request later
	mySig, _ := h.crypto.Sign(request.GetStateOnlyPeerFromSig().GetSimplexState())
	// Copy signed simplex state to avoid modifying request.
	recvdState := *request.GetStateOnlyPeerFromSig()
	recvdState.SigOfPeerTo = mySig

	var recvdSimplex entity.SimplexPaymentChannel
	err := proto.Unmarshal(recvdState.SimplexState, &recvdSimplex)
	if err != nil {
		log.Error(err)
		return nil, err // corrupted peer
	}
	cid := ctype.Bytes2Cid(recvdSimplex.GetChannelId())
	logEntry.FromCid = ctype.Cid2Hex(cid)
	logEntry.SeqNums.In = recvdSimplex.GetSeqNum()
	logEntry.SeqNums.InBase = request.GetBaseSeq()

	payInfos, requestErr := h.processPaySettleRequest(
		request, cid, peerFrom, &recvdState, &recvdSimplex, logEntry)

	var response *rpc.PaymentSettleResponse
	if requestErr != nil {
		errMsg := &rpc.Error{
			Seq:    recvdSimplex.GetSeqNum(),
			Reason: requestErr.Error(),
		}
		if requestErr == common.ErrInvalidSeqNum {
			invalidSeqNum.Add(1)
			errMsg.Code = rpc.ErrCode_INVALID_SEQ_NUM
		}
		_, stateCosigned, err2 := h.dal.GetSimplexPaymentChannel(cid, peerFrom)
		if err2 != nil {
			log.Error(err2)
		}
		response = &rpc.PaymentSettleResponse{
			StateCosigned: stateCosigned,
			Error:         errMsg,
		}
	} else {
		response = &rpc.PaymentSettleResponse{
			StateCosigned: &recvdState,
		}
		myAddr := ctype.Hex2Bytes(h.nodeConfig.GetOnChainAddr())
		reason := payInfos[0].req.GetReason()
		for _, payInfo := range payInfos {
			payID := ctype.Bytes2PayID(payInfo.req.GetSettledPayId())
			logEntry.PayId = ctype.PayID2Hex(payID)
			logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
			pay := payInfo.pay
			note, err2 := h.dal.GetPayNote(payID)
			if err2 != nil {
				log.Traceln(err2)
			}
			if bytes.Compare(pay.GetDest(), myAddr) == 0 {
				// only trigger receiving done callback if I'm recipient of the pay.
				h.tokenCallbackLock.RLock()
				if h.onReceivingToken != nil {
					go h.onReceivingToken.HandleReceivingDone(payID, pay, note, reason)
				}
				h.tokenCallbackLock.RUnlock()
			}
		}
	}

	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_PaymentSettleResponse{
			PaymentSettleResponse: response,
		},
	}
	err = h.streamWriter.WriteCelerMsg(peerFrom, celerMsg)
	if err != nil {
		return nil, err
	}

	return payInfos, requestErr
}

func (h *CelerMsgHandler) processPaySettleRequest(
	request *rpc.PaymentSettleRequest,
	cid ctype.CidType,
	peerFrom string,
	recvdState *rpc.SignedSimplexState,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) ([]*settledPayInfo, error) {

	// Check if settle reason is valid
	err := checkPaySettleReason(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Verify signature
	sig := recvdState.SigOfPeerFrom
	if !h.crypto.SigIsValid(peerFrom, recvdState.SimplexState, sig) {
		log.Errorln(common.ErrInvalidSig, peerFrom)
		return nil, common.ErrInvalidSig // corrupted peer
	}

	var payInfos []*settledPayInfo
	err = h.dal.Transactional(
		h.processPaySettleRequestTx, request, cid, peerFrom, recvdState, recvdSimplex, &payInfos, logEntry)
	if err != nil {
		return nil, err
	}

	return payInfos, nil
}

func checkPaySettleReason(request *rpc.PaymentSettleRequest) error {
	reason := request.GetSettledPays()[0].GetReason()
	switch reason {
	case rpc.PaymentSettleReason_PAY_PAID_MAX, rpc.PaymentSettleReason_PAY_REJECTED, rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		if len(request.GetSettledPays()) > 1 {
			log.Error("batched pay settle request reason not supported")
			return common.ErrInvalidSettleReason
		}
	case rpc.PaymentSettleReason_PAY_EXPIRED, rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN:
		for _, settledPay := range request.GetSettledPays() {
			if settledPay.GetReason() != reason {
				log.Error("batched pay settle request with different reasons not supported")
				return common.ErrInvalidSettleReason
			}
		}
	default:
		return common.ErrInvalidSettleReason
	}
	return nil
}

func (h *CelerMsgHandler) processPaySettleRequestTx(tx *storage.DALTx, args ...interface{}) error {
	request := args[0].(*rpc.PaymentSettleRequest)
	cid := args[1].(ctype.CidType)
	peerFrom := args[2].(string)
	recvdState := args[3].(*rpc.SignedSimplexState)
	recvdSimplex := args[4].(*entity.SimplexPaymentChannel)
	retPayInfos := args[5].(*[]*settledPayInfo)
	logEntry := args[6].(*pem.PayEventMessage)
	*retPayInfos = nil

	// channel state machine
	err := fsm.OnPscUpdateSimplex(tx, cid)
	if err != nil {
		log.Error(err)
		return err // channel closed or in dispute
	}

	storedSimplex, _, err := tx.GetSimplexPaymentChannel(cid, peerFrom)
	if err != nil {
		log.Error(err)
		return err // db error
	}
	logEntry.SeqNums.Stored = storedSimplex.GetSeqNum()

	// verify peerFrom
	if !bytes.Equal(storedSimplex.GetPeerFrom(), recvdSimplex.GetPeerFrom()) {
		log.Errorln(common.ErrInvalidChannelPeerFrom,
			ctype.Bytes2Hex(storedSimplex.GetPeerFrom()),
			ctype.Bytes2Hex(recvdSimplex.GetPeerFrom()))
		return common.ErrInvalidChannelPeerFrom // corrupted peer
	}

	// verify sequence number
	if !validRecvdSeqNum(storedSimplex.SeqNum, recvdSimplex.SeqNum, request.GetBaseSeq()) {
		log.Errorln(common.ErrInvalidSeqNum,
			"current", storedSimplex.SeqNum,
			"received", recvdSimplex.SeqNum,
			"base", request.GetBaseSeq())
		return common.ErrInvalidSeqNum // packet loss
	}

	// verify pending pay list
	newPayIDs, removedPayIDs, err := hashlist.SymmetricDifference(
		recvdSimplex.PendingPayIds.PayIds, storedSimplex.PendingPayIds.PayIds)
	if err != nil {
		log.Errorln("sym diff (pending):", err,
			utils.PrintPayIdList(recvdSimplex.PendingPayIds),
			utils.PrintPayIdList(storedSimplex.PendingPayIds))
		return common.ErrInvalidPendingPays // corrupted peer
	}
	if len(newPayIDs) > 0 || len(removedPayIDs) == 0 {
		log.Errorln(common.ErrInvalidPendingPays,
			utils.PrintPayIdList(recvdSimplex.PendingPayIds),
			utils.PrintPayIdList(storedSimplex.PendingPayIds))
		return common.ErrInvalidPendingPays // corrupted peer
	}

	var requestedPayIDs [][]byte
	var payInfos []*settledPayInfo
	for _, settledPay := range request.GetSettledPays() {
		requestedPayIDs = append(requestedPayIDs, settledPay.GetSettledPayId())
		payInfo := &settledPayInfo{
			req: settledPay,
		}
		payInfos = append(payInfos, payInfo)
	}
	diffAB, diffBA, err := hashlist.SymmetricDifference(requestedPayIDs, removedPayIDs)
	if err != nil {
		log.Errorln("sym diff (req):", err,
			utils.PrintByteArrays(requestedPayIDs), utils.PrintByteArrays(removedPayIDs))
		return common.ErrInvalidPendingPays // corrupted peer
	}
	if len(diffAB) > 0 || len(diffBA) > 0 {
		log.Errorln(common.ErrInvalidPendingPays,
			utils.PrintByteArrays(requestedPayIDs), utils.PrintByteArrays(removedPayIDs))
		return common.ErrInvalidPendingPays // corrupted peer
	}

	// get resolved pays
	resolvedAmt := new(big.Int).SetUint64(0)
	for _, payInfo := range payInfos {
		payInfo.pay, _, err = tx.GetConditionalPay(ctype.Bytes2PayID(payInfo.req.GetSettledPayId()))
		if err != nil {
			log.Errorf("Can't find pay %x to resolve: %s", payInfo.req.GetSettledPayId(), err)
			return common.ErrPayNotFound // db error
		}
		amt := new(big.Int).SetBytes(payInfo.pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
		resolvedAmt = resolvedAmt.Add(resolvedAmt, amt)
	}

	// verify unconditional transfer
	oldSendAmt := new(big.Int).SetBytes(storedSimplex.TransferToPeer.Receiver.Amt)
	newSendAmt := new(big.Int).SetBytes(recvdSimplex.TransferToPeer.Receiver.Amt)
	deltaAmt := new(big.Int).Sub(newSendAmt, oldSendAmt)
	paid := (deltaAmt.Cmp(resolvedAmt) == 0)
	if !paid && deltaAmt.Cmp(new(big.Int).SetUint64(0)) != 0 {
		// partial payment is not supported
		log.Errorln(common.ErrInvalidTransferAmt, oldSendAmt, newSendAmt, resolvedAmt)
		return common.ErrInvalidTransferAmt // should not happen if peer follows the same protocol
	}

	// verify total pending amount
	storedPendingAmt := new(big.Int).SetBytes(storedSimplex.TotalPendingAmount)
	recvdPendingAmt := new(big.Int).SetBytes(recvdSimplex.TotalPendingAmount)
	if new(big.Int).Add(recvdPendingAmt, resolvedAmt).Cmp(storedPendingAmt) != 0 {
		log.Errorln(common.ErrInvalidPendingAmt, storedPendingAmt, resolvedAmt, recvdPendingAmt)
		return common.ErrInvalidPendingAmt // corrupted peer
	}

	// verify last pay resolve deadline
	if storedSimplex.LastPayResolveDeadline != recvdSimplex.LastPayResolveDeadline {
		log.Errorln(common.ErrInvalidLastPayDeadline,
			recvdSimplex.LastPayResolveDeadline,
			storedSimplex.LastPayResolveDeadline)
		return common.ErrInvalidLastPayDeadline // corrupted peer
	}

	// verify settle reasons
	reason := request.GetSettledPays()[0].GetReason()
	switch reason {
	case rpc.PaymentSettleReason_PAY_PAID_MAX, rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN:
		if !paid {
			log.Errorln(common.ErrInvalidSettleReason, reason)
			return common.ErrInvalidSettleReason // should not happen if peer follows the same protocol
		}

	case rpc.PaymentSettleReason_PAY_EXPIRED:
		curblkNum := h.monitorService.GetCurrentBlockNumber().Uint64()
		for _, payInfo := range payInfos {
			h.checkPayRouteLoop(tx, cid, payInfo)
			if payInfo.routeLoop {
				continue
			}
			payID := ctype.Bytes2PayID(payInfo.req.GetSettledPayId())
			pay := payInfo.pay

			// check pay status first. If payment status is already INGRESS_REJECTED, continue
			// If it is not, go on other checks
			_, status, _, err2 := tx.GetPayIngressState(payID)
			if err2 != nil {
				log.Error(err2)
				return err2
			}
			if status == fsm.PayIngressRejected {
				continue
			}

			// reject settle expired requests on one of the following cases:
			// 1. pay not expired
			// 2. pay already paid to downstream, i.e., has egress PayCoSignedPaid state
			// 3. pay has no egress state and has resolved onchain with non-zero amount
			// 4. pay has egress state other than CoSignedCanceled, and has resolved onchain with non-zero amount

			// verify pay already expired
			if curblkNum < pay.GetResolveDeadline()+config.PayRecvTimeoutSafeMargin {
				log.Errorln(common.ErrInvalidSettleReason,
					"deadline", pay.GetResolveDeadline(), curblkNum)
				return common.ErrInvalidSettleReason // should not happen if peer follows the same protocol
			}
			hasEgressState, err2 := tx.HasPayEgressState(payID)
			if err2 != nil {
				log.Error(err2)
				return err2 // db error
			}
			var egressState string
			if hasEgressState {
				_, egressState, _, err2 = tx.GetPayEgressState(payID)
				if err2 != nil {
					log.Error(err2)
					return err2 // db error
				}
				// verify pay not already paid to downstream
				if egressState == fsm.PayCoSignedPaid {
					log.Errorln(common.ErrEgressPayPaid, payID.Hex())
					return common.ErrEgressPayPaid // corrupted peer
				}
			}
			// verify pay canceled downstream or not resolved on chain
			if !hasEgressState || egressState != fsm.PayCoSignedCanceled {
				amt, _, err2 := h.disputer.GetCondPayInfoFromRegistry(payID)
				if err2 != nil {
					log.Errorln("Get info from PayRegistry error", payID.Hex(), err2)
					return err2
				}
				if amt.Cmp(ctype.ZeroBigInt) == 1 { // pay is onchain resolved
					log.Errorln(common.ErrPayOnChainResolved, payID.Hex(), amt)
					return common.ErrPayOnChainResolved
				}
			}
		}

	case rpc.PaymentSettleReason_PAY_REJECTED, rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		payInfo := payInfos[0]
		h.checkPayRouteLoop(tx, cid, payInfo)
		if !payInfo.routeLoop {
			payID := ctype.Bytes2PayID(payInfos[0].req.GetSettledPayId())
			// first check ingress state, it should be on INGRESS_REJECTED
			_, status, _, err2 := tx.GetPayIngressState(payID)
			if err2 != nil {
				log.Error(err2)
				return err2
			}
			if status != fsm.PayIngressRejected {
				err2 = fmt.Errorf("wrong payment status when receiving pay settle reason of rejected or dest unreachable: %s", payID.Hex())
				log.Error(err2)
				return err2
			}
		}
	}

	// payment state machine
	for _, payInfo := range payInfos {
		if !payInfo.routeLoop {
			payID := ctype.Bytes2PayID(payInfo.req.GetSettledPayId())
			if paid {
				_, _, err = fsm.OnPayIngressCoSignedPaid(tx, payID)
			} else {
				_, _, err = fsm.OnPayIngressCoSignedCanceled(tx, payID)
			}
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	err = tx.PutSimplexState(cid, peerFrom, recvdState)
	if err != nil {
		log.Error(err)
		return err // db error
	}

	*retPayInfos = payInfos

	return nil
}

func (h *CelerMsgHandler) checkPayRouteLoop(tx *storage.DALTx, cid ctype.CidType, payInfo *settledPayInfo) {
	payID := ctype.Bytes2PayID(payInfo.req.GetSettledPayId())
	igCid, _, _, err := tx.GetPayIngressState(payID)
	if err == nil && igCid != cid {
		payInfo.routeLoop = true
	} else if ctype.Bytes2Addr(payInfo.pay.GetSrc()) == ctype.Hex2Addr(h.nodeConfig.GetOnChainAddr()) {
		payInfo.routeLoop = true
	}
}

func (h *CelerMsgHandler) paySettleRequestOutbound(payInfos []*settledPayInfo, logEntry *pem.PayEventMessage) error {
	if len(payInfos) == 0 {
		return nil
	}
	reason := payInfos[0].req.GetReason()
	if reason == rpc.PaymentSettleReason_PAY_PAID_MAX {
		pay := payInfos[0].pay
		myAddr := ctype.Hex2Bytes(h.nodeConfig.GetOnChainAddr())
		if bytes.Compare(pay.GetDest(), myAddr) != 0 {
			// foward PaidMax settle request to downstream
			payID := ctype.Bytes2PayID(payInfos[0].req.GetSettledPayId())
			settledPay := &rpc.SettledPayment{
				SettledPayId: payID[:],
				Reason:       rpc.PaymentSettleReason_PAY_PAID_MAX,
			}
			request := &rpc.PaymentSettleRequest{}
			request.SettledPays = append(request.SettledPays, settledPay)
			celerMsg := &rpc.CelerMsg{
				Message: &rpc.CelerMsg_PaymentSettleRequest{
					PaymentSettleRequest: request,
				},
			}
			_, peer, err := h.channelRouter.LookupEgressChannelOnPay(payID)
			if err != nil {
				return err
			}

			isLocalPeer, err := h.serverForwarder(peer, celerMsg)
			// Only when the peer is connecting to other OSPs and the msg is sent successfully from other OSPs,
			// the msg will not be sent from local. All other situations would triger the msg to be sent from local.
			// This ensures that PAID_MAX settle request would be eventually forwarded.
			if !(isLocalPeer == false && err == nil) {
				amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
				return h.messager.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, logEntry)
			}
		}
	}
	return nil
}
