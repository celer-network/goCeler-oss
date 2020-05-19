// Copyright 2018-2020 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/hashlist"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type settledPayInfo struct {
	req       *rpc.SettledPayment
	pay       *entity.ConditionalPay
	note      *any.Any
	igcid     ctype.CidType
	igstate   int
	egstate   int
	routeLoop bool
	delegated bool
}

func (h *CelerMsgHandler) HandlePaySettleRequest(frame *common.MsgFrame) error {
	log.Debugln("Received payment settle request from", frame.PeerAddr.Hex())
	request := frame.Message.GetPaymentSettleRequest()

	if len(request.GetSettledPays()) == 0 {
		return errors.New("empty SettledPays list")
	}

	payInfos, err := h.paySettleRequestInbound(request, frame.PeerAddr, frame.LogEntry)
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
	request *rpc.PaymentSettleRequest, peerFrom ctype.Addr, logEntry *pem.PayEventMessage) ([]*settledPayInfo, error) {

	// Sign the state in advance, verify request later
	mySig, err := h.signer.SignEthMessage(request.GetStateOnlyPeerFromSig().GetSimplexState())
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	// Copy signed simplex state to avoid modifying request.
	recvdState := *request.GetStateOnlyPeerFromSig()
	recvdState.SigOfPeerTo = mySig

	var recvdSimplex entity.SimplexPaymentChannel
	err = proto.Unmarshal(recvdState.SimplexState, &recvdSimplex)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal simplex err %w", err) // corrupted peer
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
		if errors.Is(requestErr, common.ErrInvalidSeqNum) {
			errMsg.Code = rpc.ErrCode_INVALID_SEQ_NUM
		}
		_, stateCosigned, _, err2 := h.dal.GetPeerSimplex(cid)
		if err2 != nil {
			logEntry.Error = append(logEntry.Error, fmt.Sprintf("GetPeerSimplex on requestErr err %s", err2))
		}
		response = &rpc.PaymentSettleResponse{
			StateCosigned: stateCosigned,
			Error:         errMsg,
		}
	} else {
		response = &rpc.PaymentSettleResponse{
			StateCosigned: &recvdState,
		}
		reason := payInfos[0].req.GetReason()
		for _, pi := range payInfos {
			payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
			logEntry.PayId = ctype.PayID2Hex(payID)
			logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
			if bytes.Compare(pi.pay.GetDest(), h.nodeConfig.GetOnChainAddr().Bytes()) == 0 {
				// only trigger receiving done callback if I'm recipient of the pay.
				h.tokenCallbackLock.RLock()
				if h.onReceivingToken != nil {
					go h.onReceivingToken.HandleReceivingDone(payID, pi.pay, pi.note, reason)
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
	peerFrom ctype.Addr,
	recvdState *rpc.SignedSimplexState,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) ([]*settledPayInfo, error) {

	// Check if settle reason is valid
	logEntry.SettleReason = request.GetSettledPays()[0].GetReason()
	err := checkPaySettleReason(request)
	if err != nil {
		return nil, err
	}

	// Verify signature
	sig := recvdState.SigOfPeerFrom
	if !utils.SigIsValid(peerFrom, recvdState.SimplexState, sig) {
		return nil, common.ErrInvalidSig // corrupted peer
	}

	var payInfos []*settledPayInfo
	err = h.dal.Transactional(
		h.processPaySettleRequestTx, request, cid, recvdState, recvdSimplex, &payInfos, logEntry)
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
	recvdState := args[2].(*rpc.SignedSimplexState)
	recvdSimplex := args[3].(*entity.SimplexPaymentChannel)
	retPayInfos := args[4].(*[]*settledPayInfo)
	logEntry := args[5].(*pem.PayEventMessage)
	*retPayInfos = nil

	chanState, storedSimplex, found, err := tx.GetChanStateAndPeerSimplex(cid)
	if err != nil {
		return fmt.Errorf("GetChanForRecvPayRequest err %w", err)
	}
	if !found {
		return common.ErrChannelNotFound
	}
	logEntry.SeqNums.Stored = storedSimplex.GetSeqNum()
	err = fsm.OnChannelUpdate(cid, chanState)
	if err != nil {
		return fmt.Errorf("OnChannelUpdate err %w", err)
	}

	// verify peerFrom
	sPeer, rPeer := storedSimplex.GetPeerFrom(), recvdSimplex.GetPeerFrom()
	if !bytes.Equal(sPeer, rPeer) {
		// corrupted peer
		return fmt.Errorf("%w stored %x recvd %x", common.ErrInvalidChannelPeerFrom, sPeer, rPeer)
	}

	// verify sequence number
	if !validRecvdSeqNum(storedSimplex.SeqNum, recvdSimplex.SeqNum, request.GetBaseSeq()) {
		return common.ErrInvalidSeqNum // packet loss
	}

	// verify pending pay list
	newPayIDs, removedPayIDs, err := hashlist.SymmetricDifference(
		recvdSimplex.PendingPayIds.PayIds, storedSimplex.PendingPayIds.PayIds)
	if err != nil || len(newPayIDs) > 0 || len(removedPayIDs) == 0 {
		log.Errorln("sym diff (pending):", err,
			utils.PrintPayIdList(recvdSimplex.PendingPayIds),
			utils.PrintPayIdList(storedSimplex.PendingPayIds))
		return common.ErrInvalidPendingPays // corrupted peer
	}

	var requestedPayIDs [][]byte
	var payInfos []*settledPayInfo
	for _, settledPay := range request.GetSettledPays() {
		requestedPayIDs = append(requestedPayIDs, settledPay.GetSettledPayId())
		pi := &settledPayInfo{
			req: settledPay,
		}
		payInfos = append(payInfos, pi)
	}
	diffAB, diffBA, err := hashlist.SymmetricDifference(requestedPayIDs, removedPayIDs)
	if err != nil || len(diffAB) > 0 || len(diffBA) > 0 {
		log.Errorln("sym diff (req):", err,
			utils.PrintByteArrays(requestedPayIDs), utils.PrintByteArrays(removedPayIDs))
		return common.ErrInvalidPendingPays // corrupted peer
	}

	// get resolved pays
	resolvedAmt := new(big.Int).SetUint64(0)
	for _, pi := range payInfos {
		payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
		pi.pay, pi.note, pi.igcid, pi.igstate, pi.egstate, found, err = tx.GetPayForRecvSettleReq(payID)
		if err != nil {
			return fmt.Errorf("GetPayForRecvSettleReq %x err %w", payID, err)
		}
		if !found {
			return fmt.Errorf("GetPayForRecvSettleReq %x %w", payID, common.ErrPayNotFound) // db error
		}
		amt := new(big.Int).SetBytes(pi.pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
		resolvedAmt = resolvedAmt.Add(resolvedAmt, amt)

		err = tx.DeleteSecretByPayID(payID)
		if err != nil {
			log.Errorln("DeleteSecretByPayID err", err, payID.Hex())
		}
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

	// check pay delegation
	for _, pi := range payInfos {
		payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
		status, found, err2 := tx.GetDelegatedPayStatus(payID)
		if err2 != nil {
			return fmt.Errorf("GetDelegatedPayStatus %x err %w", payID, err)
		}
		if found && status == structs.DelegatedPayStatus_RECVING {
			if paid {
				err = tx.UpdateDelegatedPayStatus(payID, structs.DelegatedPayStatus_RECVD)
				if err != nil {
					return fmt.Errorf("UpdateDelegatedPayStatus %x err %w", payID, err)
				}
				pi.delegated = true
			} else {
				err = tx.DeleteDelegatedPay(payID)
				if err != nil {
					return fmt.Errorf("DeleteDelegatedPay %x err %w", payID, err)
				}
			}
		}
	}

	// verify total pending amount
	storedPendingAmt := new(big.Int).SetBytes(storedSimplex.TotalPendingAmount)
	recvdPendingAmt := new(big.Int).SetBytes(recvdSimplex.TotalPendingAmount)
	if new(big.Int).Add(recvdPendingAmt, resolvedAmt).Cmp(storedPendingAmt) != 0 {
		log.Errorln(common.ErrInvalidPendingAmt, storedPendingAmt, resolvedAmt, recvdPendingAmt)
		return common.ErrInvalidPendingAmt // corrupted peer
	}

	// verify last pay resolve deadline
	sDeadline := storedSimplex.GetLastPayResolveDeadline()
	rDeadline := recvdSimplex.GetLastPayResolveDeadline()
	if sDeadline != rDeadline {
		log.Errorln(common.ErrInvalidLastPayDeadline, rDeadline, sDeadline)
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
		for _, pi := range payInfos {
			h.checkPayRouteLoop(cid, pi)
			if pi.routeLoop {
				continue
			}
			if pi.igstate == enums.PayState_INGRESS_REJECTED {
				continue
			}
			// reject settle expired requests on one of the following cases:
			// 1. pay not expired
			// 2. pay already paid to downstream, i.e., has egress PayCoSignedPaid state
			// 3. pay has no egress state and has resolved onchain with non-zero amount
			// 4. pay has egress state other than CoSignedCanceled, and has resolved onchain with non-zero amount

			payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
			// verify pay already expired
			if curblkNum < pi.pay.GetResolveDeadline()+config.PayRecvTimeoutSafeMargin {
				log.Errorln(common.ErrInvalidSettleReason,
					"deadline", pi.pay.GetResolveDeadline(), curblkNum)
				return common.ErrInvalidSettleReason // should not happen if peer follows the same protocol
			}

			if pi.egstate == enums.PayState_COSIGNED_PAID {
				return fmt.Errorf("%w %x", common.ErrEgressPayPaid, payID) // corrupted peer
			}

			// verify pay canceled downstream or not resolved on chain
			// TODO: pay dest can query db about the onchain resolve history instead of onchain view
			if pi.egstate != enums.PayState_COSIGNED_CANCELED {
				amt, _, err2 := h.disputer.GetCondPayInfoFromRegistry(payID)
				if err2 != nil {
					return fmt.Errorf("GetCondPayInfoFromRegistry %x err %w", payID, err2)
				}
				if amt.Cmp(ctype.ZeroBigInt) == 1 { // pay is onchain resolved
					return fmt.Errorf("%w %x %s", common.ErrPayOnChainResolved, payID, amt)
				}
			}
		}

	case rpc.PaymentSettleReason_PAY_REJECTED:
		pi := payInfos[0]
		payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
		if pi.igstate != enums.PayState_INGRESS_REJECTED {
			return fmt.Errorf("invalid status for rejected pay %x", payID)
		}

	case rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		pi := payInfos[0]
		payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
		h.checkPayRouteLoop(cid, pi)
		if !pi.routeLoop {
			if pi.igstate != enums.PayState_INGRESS_REJECTED {
				return fmt.Errorf("invalid status for unreachable pay %x", payID)
			}
		}
	}

	// payment state machine
	for _, pi := range payInfos {
		if !pi.routeLoop {
			payID := ctype.Bytes2PayID(pi.req.GetSettledPayId())
			if paid {
				err = fsm.OnPayIngressCoSignedPaid(tx, payID, pi.igstate)
			} else {
				err = fsm.OnPayIngressCoSignedCanceled(tx, payID, pi.igstate)
			}
			if err != nil {
				return fmt.Errorf("pay %x fsm err %w, paid %t", payID, err, paid)
			}
		}
	}

	err = tx.UpdateChanForRecvRequest(cid, recvdState)
	if err != nil {
		return fmt.Errorf("UpdateChanForRecvRequest err %w", err) // rare db error
	}

	*retPayInfos = payInfos

	return nil
}

func (h *CelerMsgHandler) checkPayRouteLoop(cid ctype.CidType, payInfo *settledPayInfo) {
	if payInfo.igcid != cid {
		payInfo.routeLoop = true
	} else if ctype.Bytes2Addr(payInfo.pay.GetSrc()) == h.nodeConfig.GetOnChainAddr() {
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
		payID := ctype.Bytes2PayID(payInfos[0].req.GetSettledPayId())
		delegated := payInfos[0].delegated
		if bytes.Compare(pay.GetDest(), h.nodeConfig.GetOnChainAddr().Bytes()) != 0 && !delegated {
			// forward PaidMax settle request to downstream
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
			_, peer, err := h.routeForwarder.LookupEgressChannelOnPay(payID)
			if err != nil {
				return fmt.Errorf("LookupEgressChannelOnPay err %w", err)
			}

			isLocalPeer, err := h.serverForwarder(peer, true, celerMsg)
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
