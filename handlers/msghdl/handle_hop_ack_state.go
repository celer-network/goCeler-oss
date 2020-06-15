// Copyright 2018-2020 Celer Network

package msghdl

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

func (h *CelerMsgHandler) HandleHopAckState(frame *common.MsgFrame) error {
	if frame.Message.GetCondPayResponse() == nil && frame.Message.GetPaymentSettleResponse() == nil {
		return common.ErrInvalidMsgType
	}
	var ackState *rpc.SignedSimplexState
	var ackErr *rpc.Error
	logEntry := frame.LogEntry
	if frame.Message.GetCondPayResponse() != nil {
		ackState = frame.Message.GetCondPayResponse().GetStateCosigned()
		ackErr = frame.Message.GetCondPayResponse().GetError()
	} else {
		ackState = frame.Message.GetPaymentSettleResponse().GetStateCosigned()
		ackErr = frame.Message.GetPaymentSettleResponse().GetError()
	}
	if ackErr != nil {
		logEntry.Nack = ackErr
	}

	var ackSimplex entity.SimplexPaymentChannel
	err := proto.Unmarshal(ackState.GetSimplexState(), &ackSimplex)
	if err != nil {
		return errors.New(err.Error() + "cannot parse simplex state")
	}
	log.Debugln("Receive hop ack simplex:", utils.PrintSimplexChannel(&ackSimplex))

	// Verify signature
	sigValid := eth.IsSignatureValid(frame.PeerAddr, ackState.GetSimplexState(), ackState.GetSigOfPeerTo())
	if !sigValid {
		log.Errorln(common.ErrInvalidSig, ctype.Addr2Hex(frame.PeerAddr), utils.PrintSimplexChannel(&ackSimplex))
		return common.ErrInvalidSig
	}

	ackSeqNum := ackSimplex.GetSeqNum()
	logEntry.SeqNums.Ack = ackSeqNum
	if ackSeqNum > 0 {
		sigValid = eth.IsSignatureValid(h.nodeConfig.GetOnChainAddr(), ackState.GetSimplexState(), ackState.GetSigOfPeerFrom())
		if !sigValid {
			log.Errorln(common.ErrInvalidSig, ctype.Addr2Hex(frame.PeerAddr), utils.PrintSimplexChannel(&ackSimplex))
			return common.ErrInvalidSig
		}
	}

	cid := ctype.Bytes2Cid(ackSimplex.ChannelId)
	logEntry.FromCid = ctype.Cid2Hex(cid)

	var ackedMsgs []*rpc.CelerMsg
	var nackedErrMsg *rpc.CelerMsg
	var nackedInflightMsgs []*rpc.CelerMsg
	var routeLoopPayMsg *rpc.CelerMsg
	var lastNackSeqNum uint64
	err = h.dal.Transactional(
		h.handleHopAckTx, ackState, &ackSimplex, ackErr, cid,
		&ackedMsgs, &nackedErrMsg, &nackedInflightMsgs, &lastNackSeqNum, &routeLoopPayMsg)
	if err != nil {
		log.Error(err)
	} else {
		err = h.messager.AckMsgQueue(cid, ackSeqNum, lastNackSeqNum)
		if err != nil {
			log.Error(err)
		}
		if ackErr != nil {
			if ackErr.GetCode() == rpc.ErrCode_INVALID_SEQ_NUM {
				// resend if the errored seq is greater than acked seq, and there is no inflight nacked messages
				if ackSeqNum < ackErr.GetSeq() && (ackSeqNum > lastNackSeqNum || lastNackSeqNum == 0) {
					err = h.messager.ResendMsgQueue(cid, ackSeqNum+1)
					if err != nil {
						log.Error(err)
					}
				}
			}
		}
	}

	// acknowledge ackedMsgs
	h.processAckMsgs(ackedMsgs, logEntry)

	// handle nackedErrMsg
	if nackedErrMsg != nil {
		logEntry.SeqNums.LastInflight = lastNackSeqNum
		if nackedErrMsg.GetCondPayRequest() != nil {
			request := nackedErrMsg.GetCondPayRequest()
			condPayBytes := request.GetCondPay()
			payID := ctype.PayBytes2PayID(request.GetCondPay())
			logEntry.NackPayIds = append(logEntry.NackPayIds, ctype.PayID2Hex(payID))
			var pay entity.ConditionalPay
			err = proto.Unmarshal(condPayBytes, &pay)
			if err != nil {
				log.Error(err)
				return err
			}
			h.notifyPayError(payID, &pay, ackErr.GetReason())
		} else if nackedErrMsg.GetPaymentSettleRequest() != nil {
			for _, settledPay := range nackedErrMsg.GetPaymentSettleRequest().GetSettledPays() {
				payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
				logEntry.PayId = ctype.PayID2Hex(payID)
				logEntry.NackPayIds = append(logEntry.NackPayIds, ctype.PayID2Hex(payID))
				pay, _, found, err2 := h.dal.GetPayment(payID)
				if err2 != nil {
					logEntry.Error = append(logEntry.Error, err2.Error())
					log.Errorln(err2, payID.Hex())
					continue
				}
				if !found {
					logEntry.Error = append(logEntry.Error, common.ErrPayNotFound.Error())
					log.Errorln(common.ErrPayNotFound, payID.Hex())
					continue
				}
				h.notifyPayError(payID, pay, ackErr.GetReason())
			}
		}
	}

	// resend inflight msgs after nackErr
	for _, msg := range nackedInflightMsgs {
		resendLogEntry := pem.NewPem(h.nodeConfig.GetRPCAddr())
		if msg.GetCondPayRequest() != nil {
			req := msg.GetCondPayRequest()
			var pay entity.ConditionalPay
			err = proto.Unmarshal(req.GetCondPay(), &pay)
			if err != nil {
				log.Error(err)
				return err
			}
			payID := ctype.Pay2PayID(&pay)
			directPay := req.GetDirectPay()
			log.Debugln("resend pay request", payID.Hex(), ", direct", directPay)
			resendLogEntry.Type = pem.PayMessageType_COND_PAY_REQUEST
			resendLogEntry.PayId = ctype.PayID2Hex(payID)
			resendLogEntry.Dst = ctype.Bytes2Hex(pay.GetDest())
			resendLogEntry.DirectPay = directPay
			err = h.messager.SendCondPayRequest(req.GetCondPay(), req.GetNote(), resendLogEntry)
			if err != nil {
				log.Error(err)
				resendLogEntry.Error = append(resendLogEntry.Error, err.Error())
				h.notifyPayError(payID, &pay, err.Error())
			}
		} else if msg.GetPaymentSettleRequest() != nil {
			var payIDs []ctype.PayIDType
			var pays []*entity.ConditionalPay
			var amts []*big.Int
			var reason rpc.PaymentSettleReason
			for _, settledPay := range msg.GetPaymentSettleRequest().GetSettledPays() {
				payID := ctype.Bytes2Cid(settledPay.GetSettledPayId())
				log.Debugln("resend settle request", payID.Hex())
				pay, _, found, err2 := h.dal.GetPayment(payID)
				if err2 != nil {
					log.Error(err2, payID.Hex())
					continue
				}
				if !found {
					log.Errorln(common.ErrPayNotFound, payID.Hex())
					continue
				}
				payIDs = append(payIDs, payID)
				pays = append(pays, pay)
				amts = append(amts, new(big.Int).SetBytes(settledPay.GetAmount()))
				reason = settledPay.GetReason()
			}
			resendLogEntry.Type = pem.PayMessageType_PAY_SETTLE_REQUEST
			_, err = h.messager.SendPaysSettleRequest(pays, amts, reason, resendLogEntry)
			if err != nil {
				log.Error(err)
				resendLogEntry.Error = append(resendLogEntry.Error, err.Error())
				for i, payID := range payIDs {
					pay := pays[i]
					h.notifyPayError(payID, pay, err.Error())
				}
			}
		}
		logEntry.Resend = append(logEntry.Resend, resendLogEntry)
	}

	// send settle request to clear route loop msg
	if routeLoopPayMsg != nil {
		req := routeLoopPayMsg.GetCondPayRequest()
		var pay entity.ConditionalPay
		err = proto.Unmarshal(req.GetCondPay(), &pay)
		if err != nil {
			log.Error(err)
			return err
		}
		if !h.payFromSelf(&pay) {
			payID := ctype.PayBytes2PayID(req.GetCondPay())
			ingressPeer, found, err := h.dal.GetPayIngressPeer(payID)
			if err != nil {
				return fmt.Errorf("GetPayIngressPeer err: %w", err)
			}
			if !found {
				return fmt.Errorf("GetPayIngressPeer err: %w", common.ErrPayNoIngress)
			}
			payHop := &rpc.PayHop{
				PayId:       payID.Bytes(),
				PrevHopAddr: ingressPeer.Bytes(),
				NextHopAddr: frame.PeerAddr.Bytes(),
				Err:         &rpc.Error{Code: rpc.ErrCode_PAY_ROUTE_LOOP},
			}
			payPath := &rpc.PayPath{}
			err = h.prependPayPath(payPath, payHop)
			if err != nil {
				return err
			}
			err = h.dal.PutPayPath(payID, payPath)
			if err != nil {
				return fmt.Errorf("PutPayPath err: %w", err)
			}
		}
		return h.messager.SendOnePaySettleRequest(
			&pay, new(big.Int).SetUint64(0), rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE, logEntry)
	}

	return nil
}

// HandleMysimplexFromAuthAck calls internal funcs to update db and go over msg queue
// note when this is called by handleAuthAck, msgqueue doesn't have the peer in memory yet
func (h *CelerMsgHandler) HandleMysimplexFromAuthAck(tx *storage.DALTx, cid ctype.CidType, ackState *rpc.SignedSimplexState) error {
	var ackSimplex entity.SimplexPaymentChannel
	err := proto.Unmarshal(ackState.GetSimplexState(), &ackSimplex)
	if err != nil {
		return fmt.Errorf("fail parse simplex, err: %w", err)
	}
	var ackErr *rpc.Error
	var ackedMsgs []*rpc.CelerMsg
	var nackedErrMsg *rpc.CelerMsg
	var nackedInflightMsgs []*rpc.CelerMsg
	var routeLoopPayMsg *rpc.CelerMsg
	var lastNackSeqNum uint64
	err = h.handleHopAckTx(tx, ackState, &ackSimplex, ackErr, cid,
		&ackedMsgs, &nackedErrMsg, &nackedInflightMsgs, &lastNackSeqNum, &routeLoopPayMsg)
	if err != nil {
		return err
	}
	logEntry := new(pem.PayEventMessage)
	h.processAckMsgs(ackedMsgs, logEntry)
	return nil
}

func (h *CelerMsgHandler) processAckMsgs(ackedMsgs []*rpc.CelerMsg, logEntry *pem.PayEventMessage) {
	for _, msg := range ackedMsgs {
		if msg.GetPaymentSettleRequest() != nil {
			for _, settledPay := range msg.GetPaymentSettleRequest().GetSettledPays() {
				payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
				logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
				pay, _, found, err2 := h.dal.GetPayment(payID)
				if err2 != nil {
					logEntry.Error = append(logEntry.Error, err2.Error())
					log.Errorln(err2, payID.Hex())
					continue
				}
				if !found {
					logEntry.Error = append(logEntry.Error, common.ErrPayNotFound.Error())
					log.Errorln(common.ErrPayNotFound, payID.Hex())
					continue
				}
				if h.payFromSelf(pay) {
					// I'm the sender, notify complete
					h.notifyPayComplete(settledPay, pay)
				} else {
					// I'm not the sender, forward request to upstream
					h.forwardToUpstream(settledPay, pay, logEntry)
				}
			}
		}
		if req := msg.GetCondPayRequest(); req != nil {
			payID := ctype.PayID2Hex(ctype.PayBytes2PayID(msg.GetCondPayRequest().GetCondPay()))
			logEntry.PayId = payID
			logEntry.PayIds = append(logEntry.PayIds, payID)
			if req.GetDirectPay() {
				h.notifyDirectPayComplete(msg)

				// Special case of a single ACKed direct-pay.
				if len(ackedMsgs) == 1 {
					logEntry.DirectPay = true
				}
			}
		}
	}
}

func (h *CelerMsgHandler) notifyPayComplete(
	settledPay *rpc.SettledPayment, pay *entity.ConditionalPay) {
	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	amt := new(big.Int).SetBytes(settledPay.GetAmount())
	paid := !(amt.Cmp(new(big.Int).SetUint64(0)) == 0)
	log.Debugln("notify pay complete", payID.Hex(), "paid:", paid)
	note, _, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Error(err)
	}
	h.sendingCallbackLock.RLock()
	if h.onSendingToken != nil {
		go h.onSendingToken.HandleSendComplete(payID, pay, note, settledPay.GetReason())
	}
	h.sendingCallbackLock.RUnlock()
}

func (h *CelerMsgHandler) notifyDirectPayComplete(msg *rpc.CelerMsg) {
	req := msg.GetCondPayRequest()
	var pay entity.ConditionalPay
	err := proto.Unmarshal(req.GetCondPay(), &pay)
	if err != nil {
		log.Errorln("cannot notify direct pay complete", err)
		return
	}

	payID := ctype.Pay2PayID(&pay)
	note := req.GetNote()
	reason := rpc.PaymentSettleReason_PAY_PAID_MAX
	log.Debugln("notify direct pay complete", payID.Hex())

	h.sendingCallbackLock.RLock()
	if h.onSendingToken != nil {
		go h.onSendingToken.HandleSendComplete(payID, &pay, note, reason)
	}
	h.sendingCallbackLock.RUnlock()
}

func (h *CelerMsgHandler) notifyPayError(
	payID ctype.PayIDType, pay *entity.ConditionalPay, errMsg string) {
	if !h.payFromSelf(pay) {
		// I am not the sender
		return
	}
	log.Warnln("notify pay error", payID.Hex(), errMsg)
	note, _, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Error(err)
	}
	h.sendingCallbackLock.RLock()
	if h.onSendingToken != nil {
		go h.onSendingToken.HandleSendFail(payID, pay, note, errMsg)
	}
	h.sendingCallbackLock.RUnlock()
}

func (h *CelerMsgHandler) forwardToUpstream(
	settledPay *rpc.SettledPayment, pay *entity.ConditionalPay, logEntry *pem.PayEventMessage) {

	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	if settledPay.GetReason() == rpc.PaymentSettleReason_PAY_REJECTED ||
		settledPay.GetReason() == rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN {
		log.Debugln("forward pay to upstream", payID.Hex(), settledPay.GetReason())
		err := h.messager.SendOnePaySettleProof(payID, settledPay.GetReason(), logEntry)
		if err != nil {
			logEntry.Error = append(logEntry.Error, "SendOnePaySettleProof err: "+err.Error())
			return
		}
	} else if settledPay.GetReason() == rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE {
		payPath, err := h.dal.GetPayPath(payID)
		if err != nil {
			logEntry.Error = append(logEntry.Error, "GetPayPath err: "+err.Error())
			payPath = nil
		} else {
			err = h.dal.DeletePayPath(payID)
			if err != nil {
				logEntry.Error = append(logEntry.Error, "DeletePayPath err: "+err.Error())
			}
		}
		err = h.messager.SendPayUnreachableSettleProof(payID, payPath, logEntry)
		if err != nil {
			logEntry.Error = append(logEntry.Error, "SendPayUnreachableSettleProof err: "+err.Error())
			return
		}
	}
}

func (h *CelerMsgHandler) handleHopAckTx(tx *storage.DALTx, args ...interface{}) error {
	ackState := args[0].(*rpc.SignedSimplexState)
	ackSimplex := args[1].(*entity.SimplexPaymentChannel)
	ackErr := args[2].(*rpc.Error)
	cid := args[3].(ctype.CidType)
	retAckedMsgs := args[4].(*[]*rpc.CelerMsg)
	retNackedMsg := args[5].(**rpc.CelerMsg)
	retNackeInflightMsgs := args[6].(*[]*rpc.CelerMsg)
	retLastNackSeqNum := args[7].(*uint64)
	retRouteLoopPayMsg := args[8].(**rpc.CelerMsg)
	*retAckedMsgs = nil
	*retNackedMsg = nil
	*retNackeInflightMsgs = nil
	*retRouteLoopPayMsg = nil

	_, chanState, baseSeq, lastUsedSeq, lastAckedSeq, lastNackedSeq, found, err := tx.GetChanForRecvResponse(cid)
	if err != nil {
		return fmt.Errorf("GetChanForRecvResponse err %w", err)
	}
	if !found {
		return common.ErrChannelNotFound
	}
	err = fsm.OnChannelUpdate(cid, chanState)
	if err != nil {
		return fmt.Errorf("OnChannelUpdate err %w", err)
	}

	lastNacked := lastNackedSeq
	// Handle Nacked Error message
	if ackErr != nil {
		if ackErr.GetCode() == rpc.ErrCode_PAY_ROUTE_LOOP {
			errSeq := ackErr.GetSeq()
			routeLoopPayMsg, err2 := h.getChanMessage(tx, cid, errSeq)
			if err2 != nil {
				log.Error(err2)
			}
			*retRouteLoopPayMsg = routeLoopPayMsg
		} else if ackErr.GetCode() != rpc.ErrCode_INVALID_SEQ_NUM {
			// recover from critical failure
			errSeq := ackErr.GetSeq()
			// set nackedMsg
			nackedMsg, err2 := h.getChanMessage(tx, cid, errSeq)
			if err2 != nil {
				log.Error(err2)
			}
			*retNackedMsg = nackedMsg
			req := nackedMsg.GetCondPayRequest()
			if req != nil {
				payID := ctype.PayBytes2PayID(req.GetCondPay())
				err = fsm.OnPayEgressNacked(tx, payID)
				if err != nil {
					log.Error(err)
				}
			}
			baseSeq = ackSimplex.GetSeqNum()
			// all in-flight messages become invalid and need to be resent with higher sequence numbers
			lastNackedSeq = lastUsedSeq
			nackedInflightMsgs := []*rpc.CelerMsg{}
			log.Warnln(cid.Hex(), "ackErr:", ackErr, "update base seq to", baseSeq, "last inflight", lastNackedSeq)
			for i := errSeq + 1; i <= lastNackedSeq; i++ {
				msg, err2 := h.getChanMessage(tx, cid, i)
				if err2 != nil {
					log.Error(err2)
				}
				if msg != nil {
					nackedInflightMsgs = append(nackedInflightMsgs, msg)
				}
			}
			*retNackeInflightMsgs = nackedInflightMsgs
		}
	}

	// newly acked seq, handle acked messages
	if ackSimplex.GetSeqNum() > lastAckedSeq {
		from := lastAckedSeq + 1
		// process previously nacked msg
		if lastNacked > lastAckedSeq {
			for i := from; i <= lastNacked; i++ {
				msg, err2 := h.getChanMessage(tx, cid, i)
				if err2 != nil {
					log.Error(err2)
					continue
				}
				if msg.GetCondPayRequest() != nil {
					payID := ctype.PayBytes2PayID(msg.GetCondPayRequest().GetCondPay())
					err = fsm.OnPayEgressUpdateAfterNack(tx, payID)
					if err != nil {
						return fmt.Errorf("OnPayEgressUpdateAfterNack err %w", err)
					}
				}
				err = tx.DeleteChanMessage(cid, i)
				if err != nil {
					log.Warnf("cannot delete NACKed msg %d for %x: %s", i, cid, err)
				}
			}
			from = lastNacked + 1
		}

		// process each acked message
		ackedMsgs := []*rpc.CelerMsg{}
		for i := from; i <= ackSimplex.GetSeqNum(); i++ {
			msg, err2 := h.getChanMessage(tx, cid, i)
			if err2 != nil {
				log.Error(err2)
				continue
			}
			if msg.GetCondPayRequest() != nil {
				payID := ctype.PayBytes2PayID(msg.GetCondPayRequest().GetCondPay())
				directPay := msg.GetCondPayRequest().GetDirectPay()
				if directPay {
					err = fsm.OnPayEgressCoSignedPaid(tx, payID)
				} else {
					err = fsm.OnPayEgressDelivered(tx, payID)
				}
				if err != nil {
					return err
				}
				log.Debugln("Receive ACK for cond pay request", payID.Hex(), "direct", directPay)
			} else if msg.GetPaymentSettleRequest() != nil {
				for _, pay := range msg.GetPaymentSettleRequest().GetSettledPays() {
					payID := ctype.Bytes2PayID(pay.GetSettledPayId())
					amt := new(big.Int).SetBytes(pay.GetAmount())
					paid := !((*amt).Cmp(new(big.Int).SetUint64(0)) == 0)
					if paid {
						err = fsm.OnPayEgressCoSignedPaid(tx, payID)
					} else {
						err = fsm.OnPayEgressCoSignedCanceled(tx, payID)
					}
					if err != nil {
						return err
					}
					log.Debugln("Receive ACK pay settle request", payID.Hex(), "paid:", paid)

					err = tx.DeleteSecretByPayID(payID)
					if err != nil {
						log.Errorln("DeleteSecretByPayID err", err, payID.Hex())
					}
				}
			} else {
				log.Errorln("invalid message type", msg)
				continue
			}
			ackedMsgs = append(ackedMsgs, msg)
			err = tx.DeleteChanMessage(cid, i)
			if err != nil {
				log.Warnf("cannot delete ACKed msg %d for %x: %s", i, cid, err)
			}
		}
		lastAckedSeq = ackSimplex.GetSeqNum()
		err = tx.UpdateChanForRecvResponse(cid, baseSeq, lastUsedSeq, lastAckedSeq, lastNackedSeq, ackState)
		if err != nil {
			return errors.New("UpdateChanForRecvResponse failed:" + err.Error())
		}
		*retAckedMsgs = ackedMsgs
	} else {
		err = tx.UpdateChanSeqNums(cid, baseSeq, lastUsedSeq, lastAckedSeq, lastNackedSeq)
		if err != nil {
			return errors.New("UpdateChanSeqNums failed:" + err.Error())
		}
	}
	*retLastNackSeqNum = lastNackedSeq
	return nil
}

func (h *CelerMsgHandler) getChanMessage(tx *storage.DALTx, cid ctype.CidType, seq uint64) (*rpc.CelerMsg, error) {
	var err error
	var found bool
	msg, ok := h.messager.GetMsgQueue(cid, seq)
	if !ok {
		log.Warnf("cannot get msg %d for %x from msgQueue, fetching from database", seq, cid)
		msg, found, err = tx.GetChanMessage(cid, seq)
		if err != nil {
			return nil, fmt.Errorf("cannot get msg %d for %x from database, err: %w", seq, cid, err)
		}
		if !found {
			return nil, fmt.Errorf("cannot get msg %d for %x from database", seq, cid)
		}
	}
	return msg, nil
}
