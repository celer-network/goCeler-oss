// Copyright 2018-2020 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/hashlist"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

const onchainCheckInterval = 5

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
	peerFrom := frame.PeerAddr
	request := frame.Message.GetCondPayRequest()
	var pay entity.ConditionalPay
	err := proto.Unmarshal(request.GetCondPay(), &pay)
	if err != nil {
		return fmt.Errorf("Unmarshal pay err %w", err)
	}
	payID := ctype.Pay2PayID(&pay)

	logEntry := frame.LogEntry
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.Token = utils.PrintTokenInfo(pay.GetTransferFunc().GetMaxTransfer().GetToken())
	logEntry.Src = ctype.Bytes2Hex(pay.GetSrc())
	logEntry.Dst = ctype.Bytes2Hex(pay.GetDest())

	if request.GetCrossNet().GetCrossing() {
		// proceed as crossnet payment
		return h.crossNetPayInbound(frame, pay, payID, logEntry)
	}

	// Sign the state in advance, verify request later
	mySig, err := h.signer.SignEthMessage(request.GetStateOnlyPeerFromSig().GetSimplexState())
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}
	// Copy signed simplex state to avoid modifying request.
	recvdState := *request.GetStateOnlyPeerFromSig()
	recvdState.SigOfPeerTo = mySig

	var recvdSimplex entity.SimplexPaymentChannel
	err = proto.Unmarshal(recvdState.SimplexState, &recvdSimplex)
	if err != nil {
		return common.ErrSimplexParse // corrupted peer
	}
	cid := ctype.Bytes2Cid(recvdSimplex.GetChannelId())
	directPay := request.GetDirectPay()
	logEntry.DirectPay = directPay
	log.Debugf("Receive pay request, payID %x, cid %x, peerFrom %x, direct %t", payID, cid, peerFrom, directPay)

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
		if errors.Is(requestErr, common.ErrInvalidSeqNum) {
			errMsg.Code = rpc.ErrCode_INVALID_SEQ_NUM
		}
		if errors.Is(requestErr, common.ErrPayRouteLoop) {
			errMsg.Code = rpc.ErrCode_PAY_ROUTE_LOOP
			response = &rpc.CondPayResponse{
				StateCosigned: &recvdState,
				Error:         errMsg,
			}
		} else {
			_, stateCosigned, _, err2 := h.dal.GetPeerSimplex(cid)
			if err2 != nil {
				logEntry.Error = append(logEntry.Error, fmt.Sprintf("GetPeerSimplex on requestErr err %s", err2))
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
	peerFrom ctype.Addr,
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	recvdState *rpc.SignedSimplexState,
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {

	// verify signature
	sig := recvdState.SigOfPeerFrom
	if !eth.IsSignatureValid(peerFrom, recvdState.SimplexState, sig) {
		return common.ErrInvalidSig // corrupted peer
	}

	// verify pay source
	if ctype.Bytes2Addr(pay.GetSrc()) == h.nodeConfig.GetOnChainAddr() {
		if seqErr := h.checkSeqNum(request, cid, recvdSimplex, logEntry); seqErr != nil {
			return seqErr
		}
		return common.ErrInvalidPaySrc // pay src is myself
	}

	// verify payment deadline is within limit
	blknum := h.monitorService.GetCurrentBlockNumber().Uint64()
	if pay.GetResolveDeadline() > blknum+rtconfig.GetMaxPaymentTimeout() {
		if seqErr := h.checkSeqNum(request, cid, recvdSimplex, logEntry); seqErr != nil {
			return seqErr
		}
		// should not happen if peer has the same config
		return fmt.Errorf("%w, deadline %d current %d", common.ErrInvalidPayDeadline, pay.GetResolveDeadline(), blknum)
	}
	var routeLoop bool
	err := h.dal.Transactional(
		h.processCondPayRequestTx, request, cid, payID, pay, recvdState, recvdSimplex, logEntry, &routeLoop)
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
	recvdSimplex *entity.SimplexPaymentChannel,
	logEntry *pem.PayEventMessage) error {
	storedSimplex, _, found, err := h.dal.GetPeerSimplex(cid)
	if err != nil {
		return common.ErrSimplexStateNotFound // db error
	}
	if !found {
		return common.ErrChannelNotFound
	}
	logEntry.SeqNums.Stored = storedSimplex.SeqNum
	// verify sequence number
	if !validRecvdSeqNum(storedSimplex.SeqNum, recvdSimplex.SeqNum, request.GetBaseSeq()) {
		return common.ErrInvalidSeqNum // packet loss
	}
	return nil
}

func (h *CelerMsgHandler) processCondPayRequestTx(tx *storage.DALTx, args ...interface{}) error {
	request := args[0].(*rpc.CondPayRequest)
	cid := args[1].(ctype.CidType)
	payID := args[2].(ctype.PayIDType)
	pay := args[3].(*entity.ConditionalPay)
	recvdState := args[4].(*rpc.SignedSimplexState)
	recvdSimplex := args[5].(*entity.SimplexPaymentChannel)
	logEntry := args[6].(*pem.PayEventMessage)
	retRouteLoop := args[7].(*bool)

	peer, chanState, onChainBalance, baseSeq, lastAckedSeq,
		selfSimplex, storedSimplex, found, err := tx.GetChanForRecvPayRequest(cid)
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

	// common pay verifications
	err = h.verifyCommonPayRequest(
		tx, request, cid, peer, pay, selfSimplex, storedSimplex, recvdSimplex, onChainBalance, baseSeq, lastAckedSeq)
	if err != nil {
		return err
	}

	if request.GetDirectPay() {
		// verify request
		err = h.verifyDirectPayRequest(storedSimplex, pay, recvdSimplex)
		if err != nil {
			return err
		}

		err = tx.InsertPayment(
			payID, request.GetCondPay(), pay, request.GetNote(), cid, structs.PayState_COSIGNED_PAID, ctype.ZeroCid, structs.PayState_NULL)
		if err != nil {
			return fmt.Errorf("InsertPayment err %w", err)
		}

	} else {
		// verify request
		err = h.verifyCondPayRequest(storedSimplex, payID, pay, recvdSimplex)
		if err != nil {
			return err
		}

		// TODO(xli): no need for this read, use sql write err message to tell if key already exists
		_, _, found, err2 := tx.GetPayEgress(payID)
		if err2 != nil {
			return fmt.Errorf("GetPayEgress err %w", err)
		}
		// routeLoop detected if pay info already exist
		*retRouteLoop = found
		if !found {
			err = tx.InsertPayment(
				payID, request.GetCondPay(), pay, request.GetNote(), cid, structs.PayState_COSIGNED_PENDING, ctype.ZeroCid, structs.PayState_NULL)
			if err != nil {
				return fmt.Errorf("InsertPayment err %w", err)
			}
		}
	}

	// record
	err = tx.UpdateChanForRecvRequest(cid, recvdState)
	if err != nil {
		return fmt.Errorf("UpdateChanForRecvRequest err %w", err) // rare db error
	}

	return nil
}

// verifyCondPayRequest verifies recvdSimplex from the condpay request
func (h *CelerMsgHandler) verifyCondPayRequest(
	storedSimplex *entity.SimplexPaymentChannel,
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	recvdSimplex *entity.SimplexPaymentChannel) error {

	// verify unconditional transfer
	oldAmt := new(big.Int).SetBytes(storedSimplex.TransferToPeer.Receiver.Amt)
	newAmt := new(big.Int).SetBytes(recvdSimplex.TransferToPeer.Receiver.Amt)
	if oldAmt.Cmp(newAmt) != 0 {
		// corrupted peer
		return fmt.Errorf("%w stored %s recvd %s", common.ErrInvalidTransferAmt, oldAmt, newAmt)
	}

	// verify pending pay list
	if len(recvdSimplex.PendingPayIds.PayIds) > int(rtconfig.GetMaxNumPendingPays()) {
		// should not happen if peer has the same config
		return fmt.Errorf("%w: %d", common.ErrTooManyPendingPays, len(recvdSimplex.PendingPayIds.PayIds))
	}
	newPayIDs, removedPayIDs, err := hashlist.SymmetricDifference(
		recvdSimplex.PendingPayIds.PayIds, storedSimplex.PendingPayIds.PayIds)
	if err != nil || len(removedPayIDs) > 0 || len(newPayIDs) != 1 {
		log.Errorln("sym diff:", err,
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

func (h *CelerMsgHandler) verifyDirectPayRequest(
	storedSimplex *entity.SimplexPaymentChannel,
	pay *entity.ConditionalPay,
	recvdSimplex *entity.SimplexPaymentChannel) error {

	// verify unconditional transfer
	oldSendAmt := new(big.Int).SetBytes(storedSimplex.TransferToPeer.Receiver.Amt)
	newSendAmt := new(big.Int).SetBytes(recvdSimplex.TransferToPeer.Receiver.Amt)
	deltaAmt := new(big.Int).Sub(newSendAmt, oldSendAmt)
	payAmt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	if deltaAmt.Cmp(payAmt) != 0 {
		return fmt.Errorf("%w delta %s pay %s", common.ErrInvalidTransferAmt, deltaAmt, payAmt)
	}

	return nil
}

func (h *CelerMsgHandler) verifyCommonPayRequest(
	tx *storage.DALTx,
	request *rpc.CondPayRequest,
	cid ctype.CidType,
	peer ctype.Addr,
	pay *entity.ConditionalPay,
	selfSimplex, storedSimplex, recvdSimplex *entity.SimplexPaymentChannel,
	onChainBalance *structs.OnChainBalance,
	baseSeq, lastAckedSeq uint64) error {

	// verify peerFrom
	sPeer, rPeer := storedSimplex.GetPeerFrom(), recvdSimplex.GetPeerFrom()
	if !bytes.Equal(sPeer, rPeer) {
		// corrupted peer
		return fmt.Errorf("%w stored %x recvd %x", common.ErrInvalidChannelPeerFrom, sPeer, rPeer)
	}

	storedToken := utils.GetTokenAddr(storedSimplex.GetTransferToPeer().GetToken())
	payToken := utils.GetTokenAddr(pay.GetTransferFunc().GetMaxTransfer().GetToken())
	if storedToken != payToken {
		return fmt.Errorf("%w stored %x recvd %x", common.ErrInvalidTokenAddress, storedToken, payToken)
	}

	// verify sequence number
	baseSeqNum := request.GetBaseSeq()
	if !validRecvdSeqNum(storedSimplex.SeqNum, recvdSimplex.SeqNum, baseSeqNum) {
		return common.ErrInvalidSeqNum // packet loss
	}

	// verify balance
	blkNum := h.monitorService.GetCurrentBlockNumber().Uint64()
	balance := ledgerview.ComputeBalance(
		selfSimplex, storedSimplex, onChainBalance, h.nodeConfig.GetOnChainAddr(), peer, blkNum)
	recvdAmt := new(big.Int).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	if recvdAmt.Cmp(balance.PeerFree) == 1 {
		if !h.isOSP {
			lastSyncBlk, _ := tx.GetQueryTime(config.QueryName_OnChainBalance)
			if blkNum-lastSyncBlk > onchainCheckInterval {
				log.Warnf("channel %x balance not enough, try sync with onchain balance once", cid)
				var err error
				onChainBalance, err = ledgerview.SyncOnChainBalanceTx(tx, cid, h.nodeConfig)
				if err != nil {
					log.Error(err)
				} else {
					err = tx.PutQueryTime(config.QueryName_OnChainBalance, blkNum)
					if err != nil {
						log.Error(err)
					}
					balance = ledgerview.ComputeBalance(
						selfSimplex, storedSimplex, onChainBalance, h.nodeConfig.GetOnChainAddr(), peer, blkNum)
					if recvdAmt.Cmp(balance.PeerFree) != 1 {
						return nil
					}
				}
			} else {
				log.Warnf("channel %x balance not enough, last sycned onchain balance at blk %d", cid, lastSyncBlk)
			}

		}
		// Peer does not have enough free balance
		return fmt.Errorf("%w, need %s free %s", common.ErrNoEnoughBalance, recvdAmt, balance.PeerFree) // corrupted peer
	}

	return nil
}

func (h *CelerMsgHandler) condPayRequestOutbound(frame *common.MsgFrame) error {
	peerFrom := frame.PeerAddr
	request := frame.Message.GetCondPayRequest()
	if request.GetDirectPay() {
		log.Debugln("Skip pay receipt for direct pay")
		return nil
	}
	payBytes := request.GetCondPay()
	var pay entity.ConditionalPay
	err := proto.Unmarshal(payBytes, &pay)
	if err != nil {
		return fmt.Errorf("Unmarshal payBytes err %w", err)
	}
	payID := ctype.Pay2PayID(&pay)

	dest := ctype.Bytes2Addr(pay.GetDest())
	logEntry := frame.LogEntry
	if logEntry.GetPayId() == "" {
		logEntry.PayId = ctype.PayID2Hex(payID)
	} else if logEntry.GetPayId() != ctype.PayID2Hex(payID) {
		logEntry.Error = append(logEntry.Error, "different payID:"+ctype.PayID2Hex(payID))
	}

	isRecipient := false
	var originalPayID ctype.PayIDType
	if dest == h.nodeConfig.GetOnChainAddr() {
		xnet := request.GetCrossNet()
		if xnet.GetDstNetId() == 0 {
			isRecipient = true
		} else {
			myNetId, err2 := h.dal.GetNetId()
			if err2 != nil {
				return fmt.Errorf("GetNetId err: %w", err2)
			}
			if myNetId == xnet.GetDstNetId() {
				isRecipient = true
				originalPayID = ctype.PayBytes2PayID(xnet.GetOriginalPay())
				err = h.dal.InsertCrossNetPay(
					payID, originalPayID, xnet.GetOriginalPay(), structs.CrossNetPay_DST,
					xnet.GetSrcNetId(), xnet.GetDstNetId(), ctype.ZeroAddr, 0)
				if err != nil {
					return fmt.Errorf("InsertCrossNetPay err: %w", err)
				}
			}
		}
	}

	if isRecipient {
		// reply conPay receipt
		log.Debugln("Reply pay receipt", payID.Hex())
		signedPayBytes := payBytes
		if request.GetCrossNet().GetDstNetId() != 0 {
			err = h.verifyCrossNetPay(pay, request.GetCrossNet().GetOriginalPay(), request.GetCrossNet().GetSrcNetId())
			if err != nil {
				return err
			}
			signedPayBytes = request.GetCrossNet().GetOriginalPay()
		}
		sigOfCondPay, err2 := h.signer.SignEthMessage(signedPayBytes)
		if err2 != nil {
			return err2
		}
		receipt := &rpc.CondPayReceipt{
			PayId:      payID.Bytes(),
			PayDestSig: sigOfCondPay,
		}
		if originalPayID != ctype.ZeroPayID {
			receipt.OriginalPayId = originalPayID.Bytes()
		}
		celerMsg := &rpc.CelerMsg{
			ToAddr: pay.Src,
			Message: &rpc.CelerMsg_CondPayReceipt{
				CondPayReceipt: receipt,
			},
		}
		err2 = h.streamWriter.WriteCelerMsg(peerFrom, celerMsg)
		if err2 != nil {
			return fmt.Errorf(err2.Error() + ", FAIL_SEND_RECEIPT")
		}
		return nil
	}

	// Forward condPay to next hop if I am not the destination
	log.Debugln("Forward", payID.Hex())
	delegable, proof, description := h.checkPayDelegable(&pay, ctype.Bytes2Addr(pay.GetDest()), logEntry)
	peerTo, err := h.messager.ForwardCondPayRequest(payBytes, request.GetNote(), delegable, request.GetCrossNet(), logEntry)
	if err != nil {
		if delegable && errors.Is(err, common.ErrPeerNotOnline) {
			return h.delegatePay(payID, &pay, payBytes, description, proof, peerFrom, dest, logEntry)
		}
		logEntry.Error = append(logEntry.Error, err.Error()+", DST_UNREACHABLE")
		errmsg := &rpc.Error{
			Reason: err.Error(),
		}
		if errors.Is(err, common.ErrPeerNotOnline) {
			errmsg.Code = rpc.ErrCode_PEER_NOT_ONLINE
		} else if errors.Is(err, common.ErrNoEnoughBalance) {
			errmsg.Code = rpc.ErrCode_NOT_ENOUGH_BALANCE
		} else if errors.Is(err, common.ErrRouteNotFound) {
			errmsg.Code = rpc.ErrCode_NO_ROUTE_TO_DST
		} else {
			errmsg.Code = rpc.ErrCode_MISC_ERROR
		}
		payHop := &rpc.PayHop{
			PayId:       payID.Bytes(),
			PrevHopAddr: peerFrom.Bytes(),
			NextHopAddr: peerTo.Bytes(),
			Err:         errmsg,
		}
		payPath := &rpc.PayPath{}
		err = h.prependPayPath(payPath, payHop)
		if err != nil {
			return err
		}

		// Cancel the payment upfront
		return h.messager.SendPayUnreachableSettleProof(payID, payPath, logEntry)
	}

	return nil
}

func (h *CelerMsgHandler) delegatePay(
	payID ctype.PayIDType,
	pay *entity.ConditionalPay,
	payBytes []byte,
	description *rpc.DelegationDescription,
	proof *rpc.DelegationProof,
	peerFrom ctype.Addr,
	dest ctype.Addr,
	logEntry *pem.PayEventMessage) error {

	log.Debugf("Delegating pay %x", payID)
	// Unable to send to dest but I'm authorized to delegate receiving the payment.
	logEntry.DelegationDescription = description
	sigOfCondPay, err := h.signer.SignEthMessage(payBytes)
	if err != nil {
		return fmt.Errorf("sign delegate pay err %w", err)
	}
	receipt := &rpc.CondPayReceipt{
		PayId:           payID[:],
		PayDelegatorSig: sigOfCondPay,
		DelegationProof: proof,
	}
	celerMsg := &rpc.CelerMsg{
		ToAddr: pay.Src,
		Message: &rpc.CelerMsg_CondPayReceipt{
			CondPayReceipt: receipt,
		},
	}
	err = h.streamWriter.WriteCelerMsg(peerFrom, celerMsg)
	if err != nil {
		return fmt.Errorf("send delegation receipt err %w", err)
	}
	err = h.dal.InsertDelegatedPay(payID, dest, structs.DelegatedPayStatus_RECVING)
	if err != nil {
		return fmt.Errorf("InsertDelegatedPay err %w", err)
	}
	log.Debugln("Inserted delegated pay", payID.Hex())

	return nil
}

func (h *CelerMsgHandler) checkPayDelegable(
	pay *entity.ConditionalPay, dest ctype.Addr, logEntry *pem.PayEventMessage) (
	bool, *rpc.DelegationProof, *rpc.DelegationDescription) {
	// Only able to delegate if
	// 1. payment doesn't have condition other than HashLock, aka only delegate cPay.
	// 2. found a description. if not found or error, treat the pay as not delegated.
	// 3. authorized to delegate by dest on the token type.
	// 4. authorization isn't expired.
	conditions := pay.GetConditions()
	if len(conditions) != 1 || conditions[0].GetConditionType() != entity.ConditionType_HASH_LOCK {
		return false, nil, nil
	}

	proof, found, err := h.dal.GetPeerDelegateProof(dest)
	if err != nil {
		logEntry.Error = append(logEntry.Error, "GetPeerDelegateProof:"+err.Error())
		return false, nil, nil
	}
	if !found || proof == nil {
		return false, nil, nil
	}
	description, err := utils.UnmarshalDelegationDescription(proof)
	if err != nil {
		logEntry.Error = append(logEntry.Error, "UnmarshalDelegationDescription:"+err.Error())
		return false, nil, nil
	}

	token := pay.GetTransferFunc().GetMaxTransfer().GetToken().GetTokenAddress()
	delegable := ctype.Bytes2Addr(description.GetDelegatee()) == dest &&
		hashlist.Exist(description.GetTokenToDelegate(), token) &&
		description.GetExpiresAfterBlock() > h.monitorService.GetCurrentBlockNumber().Int64()

	return delegable, proof, description
}

func (h *CelerMsgHandler) crossNetPayInbound(
	frame *common.MsgFrame, pay entity.ConditionalPay, payID ctype.PayIDType, logEntry *pem.PayEventMessage) error {
	request := frame.Message.GetCondPayRequest()
	xnet := request.GetCrossNet()
	originalPayId := ctype.PayBytes2PayID(xnet.GetOriginalPay())

	logEntry.Xnet.SrcNetId = xnet.GetSrcNetId()
	logEntry.Xnet.DstNetId = xnet.GetDstNetId()
	logEntry.Xnet.OriginalPayId = ctype.PayID2Hex(originalPayId)
	logEntry.Xnet.FromPayId = ctype.PayID2Hex(payID)
	logEntry.Xnet.State = pem.CrossNetPayState_XNET_INGRESS

	bridgeAddr := frame.PeerAddr
	bridgeNetId, found, err := h.dal.GetNetBridge(bridgeAddr)
	if err != nil {
		return fmt.Errorf("GetNetBridge err: %w", err)
	}
	if !found {
		return fmt.Errorf("bridgeNetId not found")
	}
	logEntry.Xnet.FromBridgeId = bridgeNetId

	if ctype.Bytes2Addr(pay.GetDest()) == h.nodeConfig.GetOnChainAddr() {
		return fmt.Errorf("ingress bridge cannot be the dest addr")
	}

	newPay := pay
	if pay.GetTransferFunc().GetMaxTransfer().GetToken().GetTokenType() != entity.TokenType_ETH {
		localToken, found, err2 := h.dal.GetLocalToken(bridgeNetId, pay.GetTransferFunc().GetMaxTransfer().GetToken())
		if err2 != nil {
			return fmt.Errorf("GetLocalToken err: %w", err2)
		}
		if !found {
			return fmt.Errorf("local token not found")
		}
		newPay.TransferFunc.MaxTransfer.Token = localToken
	}
	// TODO: update resolve dealine and timeout, check conditions
	newPay.ResolveDeadline = h.monitorService.GetCurrentBlockNumber().Uint64() + xnet.GetTimeout()
	newPay.ResolveTimeout = config.PayResolveTimeout
	newPay.PayResolver = h.nodeConfig.GetPayResolverContract().GetAddr().Bytes()
	newPayID := ctype.Pay2PayID(&newPay)
	newPayBytes, err := proto.Marshal(&newPay)
	if err != nil {
		return err
	}

	err = h.dal.InsertCrossNetPay(
		newPayID, originalPayId, xnet.GetOriginalPay(), structs.CrossNetPay_INGRESS, xnet.GetSrcNetId(), xnet.GetDstNetId(),
		bridgeAddr, bridgeNetId)
	if err != nil {
		return fmt.Errorf("InsertCrossNetPay err: %w", err)
	}

	xnet.Crossing = false
	xnet.BridgeAddr = nil
	xnet.BridgeNetId = 0
	xnet.Timeout = 0

	frame.Message = &rpc.CelerMsg{
		Message: &rpc.CelerMsg_CondPayRequest{
			CondPayRequest: &rpc.CondPayRequest{
				CondPay:  newPayBytes,
				Note:     request.GetNote(),
				CrossNet: xnet,
			},
		},
	}

	logEntry.PayId = ctype.PayID2Hex(newPayID)
	logEntry.Token = utils.PrintTokenInfo(newPay.GetTransferFunc().GetMaxTransfer().GetToken())

	return nil
}

func (h *CelerMsgHandler) verifyCrossNetPay(pay entity.ConditionalPay, originalPayBytes []byte, srcNetId uint64) error {
	var originalPay entity.ConditionalPay
	err := proto.Unmarshal(originalPayBytes, &originalPay)
	if err != nil {
		return err
	}
	if originalPay.GetTransferFunc().GetMaxTransfer().GetToken().GetTokenType() != entity.TokenType_ETH {
		token, found, err2 := h.dal.GetLocalToken(srcNetId, originalPay.GetTransferFunc().GetMaxTransfer().GetToken())
		if err2 != nil {
			return fmt.Errorf("GetLocalToken err: %w", err2)
		}
		if !found {
			return fmt.Errorf("local token not found")
		}
		originalPay.TransferFunc.MaxTransfer.Token = token
	}

	pay.ResolveDeadline = 0
	pay.ResolveTimeout = 0
	pay.PayResolver = nil
	originalPay.ResolveDeadline = 0
	originalPay.ResolveTimeout = 0
	originalPay.PayResolver = nil

	b1, err := proto.Marshal(&pay)
	if err != nil {
		return err
	}
	b2, err := proto.Marshal(&originalPay)
	if err != nil {
		return err
	}
	if bytes.Compare(b1, b2) != 0 {
		return fmt.Errorf("original pay not match")
	}

	return nil
}
