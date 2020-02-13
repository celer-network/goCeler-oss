// Copyright 2018-2019 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"math/big"
	"time"

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
	"github.com/zserge/metric"
)

var invalidSeqNum metric.Metric
var payRouteLoop metric.Metric
var payDstNoRoute metric.Metric

func init() {
	invalidSeqNum = metric.NewCounter("5d1h", "24h10m", "15m10s")
	expvar.Publish("invalid-seq-num", invalidSeqNum)
	payRouteLoop = metric.NewCounter("5d1h", "24h10m", "15m10s")
	expvar.Publish("pay-route-loop", payRouteLoop)
	payDstNoRoute = metric.NewCounter("5d1h", "24h10m", "15m10s")
	expvar.Publish("pay-dst-no-route", payDstNoRoute)
}

func (h *CelerMsgHandler) HandleCondPayRequest(frame *common.MsgFrame) error {
	if frame.Message.GetCondPayRequest() == nil {
		return common.ErrInvalidMsgType
	}
	err := h.condPayRequestInbound(frame)
	if err != nil {
		return err
	}
	err = h.condPayRequestOutbound(frame)
	if err != nil {
		return err
	}
	return nil
}

func (h *CelerMsgHandler) condPayRequestInbound(frame *common.MsgFrame) error {
	peerFrom := ctype.Addr2Hex(frame.PeerAddr)
	request := frame.Message.GetCondPayRequest()
	var pay entity.ConditionalPay
	err := proto.Unmarshal(request.GetCondPay(), &pay)
	if err != nil {
		log.Error(err)
		return err
	}
	payID := ctype.Pay2PayID(&pay)

	logEntry := frame.LogEntry
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.Token = utils.PrintTokenInfo(pay.GetTransferFunc().GetMaxTransfer().GetToken())
	logEntry.Src = ctype.Bytes2Hex(pay.GetSrc())
	logEntry.Dst = ctype.Bytes2Hex(pay.GetDest())

	// Sign the state in advance, verify request later
	mySig, _ := h.crypto.Sign(request.GetStateOnlyPeerFromSig().GetSimplexState())
	// Copy signed simplex state to avoid modifying request.
	recvdState := *request.GetStateOnlyPeerFromSig()
	recvdState.SigOfPeerTo = mySig

	var recvdSimplex entity.SimplexPaymentChannel
	err = proto.Unmarshal(recvdState.SimplexState, &recvdSimplex)
	if err != nil {
		log.Error(common.ErrSimplexParse)
		return common.ErrSimplexParse // corrupted peer
	}
	cid := ctype.Bytes2Cid(recvdSimplex.GetChannelId())
	directPay := request.GetDirectPay()
	logEntry.DirectPay = directPay
	log.Debugf("Receive pay request, payID %x, cid %x, peerFrom %s, direct %t", payID, cid, peerFrom, directPay)

	logEntry.FromCid = ctype.Cid2Hex(cid)
	logEntry.SeqNums.In = recvdSimplex.GetSeqNum()
	logEntry.SeqNums.InBase = request.GetBaseSeq()

	requestErr := h.processCondPayRequest(
		request, cid, peerFrom, payID, &pay, &recvdState, &recvdSimplex, logEntry)

	var response *rpc.CondPayResponse
	if requestErr != nil {
		errMsg := &rpc.Error{
			Seq:    recvdSimplex.GetSeqNum(),
			Reason: requestErr.Error(),
		}
		if requestErr == common.ErrInvalidSeqNum {
			invalidSeqNum.Add(1)
			errMsg.Code = rpc.ErrCode_INVALID_SEQ_NUM
		}
		if requestErr == common.ErrPayRouteLoop {
			payRouteLoop.Add(1)
			errMsg.Code = rpc.ErrCode_PAY_ROUTE_LOOP
			response = &rpc.CondPayResponse{
				StateCosigned: &recvdState,
				Error:         errMsg,
			}
		} else {
			_, stateCosigned, err2 := h.dal.GetSimplexPaymentChannel(cid, peerFrom)
			if err2 != nil {
				logEntry.Error = append(logEntry.Error, err2.Error())
				log.Error(err2)
			}
			response = &rpc.CondPayResponse{
				StateCosigned: stateCosigned,
				Error:         errMsg,
			}
		}
	} else {
		response = &rpc.CondPayResponse{
			StateCosigned: &recvdState,
		}

		if directPay {
			log.Trace("direct Pay received: ", payID)
			note := request.GetNote()
			reason := rpc.PaymentSettleReason_PAY_PAID_MAX
			h.tokenCallbackLock.RLock()
			if h.onReceivingToken != nil {
				go h.onReceivingToken.HandleReceivingDone(payID, &pay, note, reason)
			}
			h.tokenCallbackLock.RUnlock()
		}
	}

	log.Tracef("Replying (direct %t): %s", directPay, response.String())
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayResponse{
			CondPayResponse: response,
		},
	}
	err = h.streamWriter.WriteCelerMsg(peerFrom, celerMsg)
	if err != nil {
		if requestErr != nil {
			logEntry.Error = append(logEntry.Error, err.Error())
			return requestErr
		}
		return err
	}
	return requestErr
}

func (h *CelerMsgHandler) processCondPayRequest(
	request *rpc.CondPayRequest,
	cid ctype.CidType,
	peerFrom string,
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	recvdState *rpc.SignedSimplexState,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {

	// verify channel ID
	storedCid := h.dal.GetCidByPeerAndToken(
		ctype.Hex2Bytes(peerFrom), pay.GetTransferFunc().GetMaxTransfer().GetToken())
	if storedCid != cid {
		log.Errorln(common.ErrInvalidChannelID, storedCid.Hex(), cid.Hex())
		return common.ErrInvalidChannelID
	}

	// verify signature
	sig := recvdState.SigOfPeerFrom
	if !h.crypto.SigIsValid(peerFrom, recvdState.SimplexState, sig) {
		log.Errorln(common.ErrInvalidSig, peerFrom)
		return common.ErrInvalidSig // corrupted peer
	}

	// verify pay source
	if ctype.Bytes2Addr(pay.GetSrc()) == ctype.Hex2Addr(h.nodeConfig.GetOnChainAddr()) {
		log.Errorln(common.ErrInvalidPaySrc, utils.PrintConditionalPay(pay))
		if seqErr := h.checkSeqNum(request, cid, peerFrom, recvdSimplex, logEntry); seqErr != nil {
			return seqErr
		}
		return common.ErrInvalidPaySrc // pay src is myself
	}

	// verify payment deadline is within limit
	if pay.GetResolveDeadline() > h.monitorService.GetCurrentBlockNumber().Uint64()+rtconfig.GetMaxPaymentTimeout() {
		log.Errorln(common.ErrInvalidPayDeadline,
			pay.GetResolveDeadline(), h.monitorService.GetCurrentBlockNumber().Uint64())
		if seqErr := h.checkSeqNum(request, cid, peerFrom, recvdSimplex, logEntry); seqErr != nil {
			return seqErr
		}
		return common.ErrInvalidPayDeadline // should not happen if peer has the same config
	}
	routeLoop, _ := h.dal.HasPayEgressState(payID)
	err := h.dal.Transactional(
		h.processCondPayRequestTx, request, cid, peerFrom, payID, pay, recvdState, recvdSimplex, logEntry, routeLoop)
	if err != nil {
		return err
	}
	if routeLoop {
		return common.ErrPayRouteLoop
	}
	return nil
}

// checkSeqNum is used to to give ErrInvalidSeqNum higher priority over other errors.
// It is only called when another error has already been found
func (h *CelerMsgHandler) checkSeqNum(
	request *rpc.CondPayRequest,
	cid ctype.CidType,
	peerFrom string,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {
	storedSimplex, _, err := h.dal.GetSimplexPaymentChannel(cid, peerFrom)
	if err != nil {
		log.Errorln(err, cid)
		return common.ErrSimplexStateNotFound // db error
	}
	logEntry.SeqNums.Stored = storedSimplex.SeqNum
	// verify sequence number
	if !validRecvdSeqNum(storedSimplex.SeqNum, recvdSimplex.SeqNum, request.GetBaseSeq()) {
		log.Errorln(common.ErrInvalidSeqNum,
			"current", storedSimplex.SeqNum,
			"received", recvdSimplex.SeqNum,
			"base", request.GetBaseSeq())
		return common.ErrInvalidSeqNum // packet loss
	}
	return nil
}

func (h *CelerMsgHandler) processCondPayRequestTx(tx *storage.DALTx, args ...interface{}) error {
	request := args[0].(*rpc.CondPayRequest)
	cid := args[1].(ctype.CidType)
	peerFrom := args[2].(string)
	payID := args[3].(ctype.PayIDType)
	pay := args[4].(*entity.ConditionalPay)
	recvdState := args[5].(*rpc.SignedSimplexState)
	recvdSimplex := args[6].(*entity.SimplexPaymentChannel)
	logEntry := args[7].(*pem.PayEventMessage)
	routeLoop := args[8].(bool)

	// channel state machine
	err := fsm.OnPscUpdateSimplex(tx, cid)
	if err != nil {
		log.Error(err)
		return err // channel closed or in dispute
	}

	if request.GetDirectPay() {
		// verify request
		err = h.verifyDirectPayRequestTx(tx, request, cid, peerFrom, payID, pay, recvdSimplex, logEntry)
		if err != nil {
			return err
		}

		// payment state machine
		_, _, err = fsm.OnPayIngressDirectCoSignedPaid(tx, payID, cid)
		if err != nil {
			log.Error(err)
			return err // should not happen if peer follows the same protocol
		}
	} else {
		// verify request
		err = h.verifyCondPayRequestTx(tx, request, cid, peerFrom, payID, pay, recvdSimplex, logEntry)
		if err != nil {
			return err
		}
		// pay state should NOT be updated on routeLoop detected
		// as it will overwrite pay state set by the first time processing the pay
		if !routeLoop {
			// payment state machine
			_, _, err = fsm.OnPayIngressCoSignedPending(tx, payID, cid)
			if err != nil {
				log.Error(err)
				return err // should not happen if peer follows the same protocol
			}
		}
	}

	// record
	err = tx.PutSimplexState(cid, peerFrom, recvdState)
	if err != nil {
		log.Error(err)
		return err // rare db error
	}
	if routeLoop {
		return nil
	}

	err = tx.PutConditionalPay(request.CondPay)
	if err != nil {
		log.Error(err)
		return err // rare db error
	}

	err = tx.PutPayNote(pay, request.Note)
	if err != nil {
		log.Error(err)
		return err // rare db error
	}
	return nil
}

// verifyCondPayRequest verifies recvdSimplex from the condpay request
func (h *CelerMsgHandler) verifyCondPayRequestTx(
	tx *storage.DALTx,
	request *rpc.CondPayRequest,
	cid ctype.CidType,
	peerFrom string,
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {

	// Get stored simplex channel
	storedSimplex, _, err := tx.GetSimplexPaymentChannel(cid, peerFrom)
	if err != nil {
		log.Errorln(err, cid)
		return common.ErrSimplexStateNotFound // db error
	}
	logEntry.SeqNums.Stored = storedSimplex.GetSeqNum()

	// common pay verifications
	err = h.verifyCommonPayRequest(tx, cid, pay, storedSimplex, recvdSimplex, request)
	if err != nil {
		return err
	}

	// verify unconditional transfer
	if !bytes.Equal(storedSimplex.GetTransferToPeer().GetReceiver().GetAmt(),
		recvdSimplex.GetTransferToPeer().GetReceiver().GetAmt()) {
		log.Errorln(common.ErrInvalidTransferAmt,
			new(big.Int).SetBytes(storedSimplex.GetTransferToPeer().GetReceiver().GetAmt()),
			new(big.Int).SetBytes(recvdSimplex.GetTransferToPeer().GetReceiver().GetAmt()),
		)
		return common.ErrInvalidTransferAmt // corrupted peer
	}

	// verify pending pay list
	if len(recvdSimplex.PendingPayIds.PayIds) > int(rtconfig.GetMaxNumPendingPays()) {
		log.Errorln(common.ErrTooManyPendingPays, len(recvdSimplex.PendingPayIds.PayIds))
		return common.ErrTooManyPendingPays // should not happen if peer has the same config
	}
	newPayIDs, removedPayIDs, err := hashlist.SymmetricDifference(
		recvdSimplex.PendingPayIds.PayIds, storedSimplex.PendingPayIds.PayIds)
	if err != nil {
		log.Errorln("sym diff:", err,
			utils.PrintPayIdList(recvdSimplex.PendingPayIds),
			utils.PrintPayIdList(storedSimplex.PendingPayIds))
		return common.ErrInvalidPendingPays // corrupted peer
	}
	if len(removedPayIDs) > 0 || len(newPayIDs) != 1 {
		log.Errorln(common.ErrInvalidPendingPays,
			utils.PrintPayIdList(recvdSimplex.PendingPayIds),
			utils.PrintPayIdList(storedSimplex.PendingPayIds))
		return common.ErrInvalidPendingPays // corrupted peer
	}
	if !bytes.Equal(payID[:], newPayIDs[0]) {
		log.Errorln(common.ErrInvalidPendingPays, ctype.PayID2Hex(payID), ctype.Bytes2Hex(newPayIDs[0]))
		return common.ErrInvalidPendingPays // corrupted peer
	}

	// verify last pay resolve deadline
	deadline := storedSimplex.GetLastPayResolveDeadline()
	if pay.GetResolveDeadline() > deadline {
		deadline = pay.GetResolveDeadline()
	}
	if deadline != recvdSimplex.GetLastPayResolveDeadline() {
		log.Errorln(common.ErrInvalidLastPayDeadline, recvdSimplex.LastPayResolveDeadline, deadline)
		return common.ErrInvalidLastPayDeadline // corrupted peer
	}

	// verify pay resolver address
	payResolver := ctype.Bytes2Addr(pay.GetPayResolver())
	if payResolver != h.nodeConfig.GetPayResolverContract().GetAddr() {
		log.Errorln(common.ErrInvalidPayResolver, payResolver.Hex())
		return common.ErrInvalidPayResolver // should not happen if peer has the same config
	}

	// verify total pending amount
	storedPendingAmt := new(big.Int).SetBytes(storedSimplex.TotalPendingAmount)
	recvdPendingAmt := new(big.Int).SetBytes(recvdSimplex.TotalPendingAmount)
	recvdAmt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	if new(big.Int).Add(storedPendingAmt, recvdAmt).Cmp(recvdPendingAmt) != 0 {
		log.Errorln(common.ErrInvalidPendingAmt, storedPendingAmt, recvdAmt, recvdPendingAmt)
		return common.ErrInvalidPendingAmt // corrupted peer
	}

	return nil
}

func (h *CelerMsgHandler) verifyDirectPayRequestTx(
	tx *storage.DALTx,
	request *rpc.CondPayRequest,
	cid ctype.CidType,
	peerFrom string,
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {

	// Get stored simplex channel
	storedSimplex, _, err := tx.GetSimplexPaymentChannel(cid, peerFrom)
	if err != nil {
		log.Error(err)
		return err // db error
	}
	logEntry.SeqNums.Stored = storedSimplex.GetSeqNum()

	// common pay verifications
	err = h.verifyCommonPayRequest(tx, cid, pay, storedSimplex, recvdSimplex, request)
	if err != nil {
		return err
	}

	// verify unconditional transfer
	oldSendAmt := new(big.Int).SetBytes(storedSimplex.TransferToPeer.Receiver.Amt)
	newSendAmt := new(big.Int).SetBytes(recvdSimplex.TransferToPeer.Receiver.Amt)
	deltaAmt := new(big.Int).Sub(newSendAmt, oldSendAmt)
	payAmt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	if deltaAmt.Cmp(payAmt) != 0 {
		log.Errorln(common.ErrInvalidTransferAmt, deltaAmt, payAmt)
		return common.ErrInvalidTransferAmt
	}

	return nil
}

func (h *CelerMsgHandler) verifyCommonPayRequest(
	tx *storage.DALTx,
	cid ctype.CidType,
	pay *entity.ConditionalPay,
	stored, recvd *entity.SimplexPaymentChannel,
	req *rpc.CondPayRequest) error {

	// verify peerFrom
	sPeer, rPeer := stored.GetPeerFrom(), recvd.GetPeerFrom()
	if !bytes.Equal(sPeer, rPeer) {
		log.Errorln(common.ErrInvalidChannelPeerFrom,
			ctype.Bytes2Hex(sPeer), ctype.Bytes2Hex(rPeer))
		return common.ErrInvalidChannelPeerFrom // corrupted peer
	}

	// verify sequence number
	baseSeqNum := req.GetBaseSeq()
	if !validRecvdSeqNum(stored.SeqNum, recvd.SeqNum, baseSeqNum) {
		log.Errorln(common.ErrInvalidSeqNum,
			"current", stored.SeqNum,
			"received", recvd.SeqNum,
			"base", baseSeqNum)
		return common.ErrInvalidSeqNum // packet loss
	}

	// verify balance
	blkNum := h.monitorService.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalanceTx(tx, cid, h.nodeConfig.GetOnChainAddr(), blkNum)
	if err != nil {
		log.Errorln(err, "unabled to find balance for cid", cid.Hex())
		return err // local db error
	}
	recvdAmt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	if recvdAmt.Cmp(balance.PeerFree) == 1 {
		// Peer does not have enough free balance
		log.Errorln("Not enough balance to receive on", cid.Hex(), "need:", recvdAmt, "have:", balance.PeerFree)
		return fmt.Errorf("%s, peer free %s", common.ErrNoEnoughBalance, balance.PeerFree.String()) // corrupted peer
	}

	return nil
}

func (h *CelerMsgHandler) condPayRequestOutbound(msg *common.MsgFrame) error {
	condPayRequest := msg.Message.GetCondPayRequest()
	if condPayRequest.GetDirectPay() {
		log.Debugln("Skip pay receipt for direct pay")
		return nil
	}

	var condPay entity.ConditionalPay
	err := proto.Unmarshal(condPayRequest.CondPay, &condPay)
	if err != nil {
		return errors.New("CANNOT_PARSE_COND_PAY_REQUEST")
	}
	payID := ctype.Pay2PayID(&condPay)

	dst := ctype.Bytes2Hex(condPay.GetDest())
	logEntry := msg.LogEntry
	if logEntry.GetPayId() == "" {
		logEntry.PayId = ctype.PayID2Hex(payID)
	} else if logEntry.GetPayId() != ctype.PayID2Hex(payID) {
		logEntry.Error = append(logEntry.Error, "different payID:"+ctype.PayID2Hex(payID))
	}
	if dst == h.nodeConfig.GetOnChainAddr() {
		// reply conPay receipt
		log.Debugln("Reply pay receipt", payID.Hex())
		sigOfCondPay, err2 := h.crypto.Sign(condPayRequest.CondPay)
		if err2 != nil {
			return err2
		}
		receipt := &rpc.CondPayReceipt{
			PayId:      payID[:],
			PayDestSig: sigOfCondPay,
		}
		// There is possibility that receipt is received before cond pay ack. We intentionally add a small
		// delay to mitigate the misorder issue. Correct behavior may be maintaining a state machine of a pay and keep
		// message ordering. This is fine since sink usually happens on client which doesn't relay payment.
		time.Sleep(5 * time.Millisecond)
		celerMsg := &rpc.CelerMsg{
			ToAddr: condPay.Src,
			Message: &rpc.CelerMsg_CondPayReceipt{
				CondPayReceipt: receipt,
			},
		}
		err2 = h.streamWriter.WriteCelerMsg(ctype.Addr2Hex(msg.PeerAddr), celerMsg)
		if err2 != nil {
			log.Warn("Cannot send receipt to " + ctype.Bytes2Hex(condPay.Src) + ":" + err2.Error())
			return errors.New(err2.Error() + " FAIL_SEND_RECEIPT")
		}
	} else {
		// Forward condPay to next hop
		log.Debugln("Forward", payID.Hex())
		err = h.messager.SendCondPayRequest(&condPay, condPayRequest.GetNote(), logEntry)
		if err != nil {
			log.Error(err)
			logEntry.Error = append(logEntry.Error, err.Error()+" DST_UNREACHABLE")
			// Cancel the payment upfront
			payDstNoRoute.Add(1)
			h.rejectUnreachablePay(payID, logEntry)
		}
	}
	return nil
}

func (h *CelerMsgHandler) rejectUnreachablePay(payID ctype.PayIDType, logEntry *pem.PayEventMessage) {
	err := h.messager.SendOnePaySettleProof(
		payID,
		rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE,
		logEntry,
	)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error()+" SendOnePaySettleProof")
		log.Error(err)
	}
}
