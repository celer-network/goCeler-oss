// Copyright 2018-2019 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
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

	peerTo := ctype.Addr2Hex(frame.PeerAddr)
	// Verify signature
	sigValid := h.crypto.SigIsValid(peerTo, ackState.GetSimplexState(), ackState.GetSigOfPeerTo())
	if !sigValid {
		log.Errorln(common.ErrInvalidSig, peerTo, utils.PrintSimplexChannel(&ackSimplex))
		return common.ErrInvalidSig
	}

	ackSeqNum := ackSimplex.GetSeqNum()
	logEntry.SeqNums.Ack = ackSeqNum
	if ackSeqNum > 0 {
		sigValid = h.crypto.SigIsValid(h.nodeConfig.GetOnChainAddr(), ackState.GetSimplexState(), ackState.GetSigOfPeerFrom())
		if !sigValid {
			log.Errorln(common.ErrInvalidSig, peerTo, utils.PrintSimplexChannel(&ackSimplex))
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
	for _, msg := range ackedMsgs {
		if msg.GetPaymentSettleRequest() != nil {
			for _, settledPay := range msg.GetPaymentSettleRequest().GetSettledPays() {
				payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
				logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
				pay, _, err2 := h.dal.GetConditionalPay(payID)
				if err2 != nil {
					logEntry.Error = append(logEntry.Error, err2.Error())
					log.Errorln("Can't find completed pay", payID.Hex(), err2)
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
				pay, _, err2 := h.dal.GetConditionalPay(payID)
				if err2 != nil {
					logEntry.Error = append(logEntry.Error, err2.Error())
					log.Errorln("Can't find failed pay", payID.Hex(), err2)
					return err2
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
			err = h.messager.SendCondPayRequest(&pay, req.GetNote(), resendLogEntry)
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
				pay, _, err2 := h.dal.GetConditionalPay(payID)
				if err2 != nil {
					log.Error(err2, payID.Hex())
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
		payAmt := new(big.Int).SetUint64(0)
		return h.messager.SendOnePaySettleRequest(
			&pay, payAmt, rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE, logEntry)
	}

	return nil
}

func (h *CelerMsgHandler) notifyPayComplete(
	settledPay *rpc.SettledPayment, pay *entity.ConditionalPay) {
	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	amt := new(big.Int).SetBytes(settledPay.GetAmount())
	paid := !(amt.Cmp(new(big.Int).SetUint64(0)) == 0)
	log.Debugln("notify pay complete", payID.Hex(), "paid:", paid)
	note, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Traceln(err)
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
	note, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Traceln(err)
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
		settledPay.GetReason() == rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE {

		log.Debugln("forward rejected pay to upstream", payID.Hex())
		err := h.messager.SendOnePaySettleProof(payID, settledPay.GetReason(), logEntry)
		if err != nil {
			log.Error(err)
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
	myAddr := h.nodeConfig.GetOnChainAddr()

	// channel state machine
	err := fsm.OnPscUpdateSimplex(tx, cid)
	if err != nil {
		log.Error(err)
		return err
	}

	seqNums, err := tx.GetChannelSeqNums(cid)
	if err != nil {
		_, seqNums, err = ledgerview.GetChannelSeqNumsFromSimplexState(tx, cid, myAddr)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	lastNacked := seqNums.LastNacked
	// Handle Nacked Error message
	if ackErr != nil {
		if ackErr.GetCode() == rpc.ErrCode_PAY_ROUTE_LOOP {
			errSeq := ackErr.GetSeq()
			routeLoopPayMsg, _ := h.getChannelMessage(tx, cid, errSeq)
			*retRouteLoopPayMsg = routeLoopPayMsg
		} else if ackErr.GetCode() != rpc.ErrCode_INVALID_SEQ_NUM {
			// recover from critical failure
			errSeq := ackErr.GetSeq()
			// set nackedMsg
			nackedMsg, _ := h.getChannelMessage(tx, cid, errSeq)
			*retNackedMsg = nackedMsg
			req := nackedMsg.GetCondPayRequest()
			if req != nil {
				payID := ctype.PayBytes2PayID(req.GetCondPay())
				_, _, err = fsm.OnPayEgressNacked(tx, payID)
				if err != nil {
					log.Error(err)
				}
			}
			seqNums.Base = ackSimplex.GetSeqNum()
			// all in-flight messages become invalid and need to be resent with higher sequence numbers
			seqNums.LastNacked = seqNums.LastUsed
			nackedInflightMsgs := []*rpc.CelerMsg{}
			log.Warnln(cid.Hex(), "ackErr:", ackErr, "update base seq to", seqNums.Base, "last inflight", seqNums.LastNacked)
			for i := errSeq + 1; i <= seqNums.LastNacked; i++ {
				msg, _ := h.getChannelMessage(tx, cid, i)
				if msg != nil {
					nackedInflightMsgs = append(nackedInflightMsgs, msg)
				}
			}
			*retNackeInflightMsgs = nackedInflightMsgs
		}
	}

	// newly acked seq, handle acked messages
	if ackSimplex.GetSeqNum() > seqNums.LastAcked {
		from := seqNums.LastAcked + 1
		// process previously nacked msg
		if lastNacked > seqNums.LastAcked {
			for i := from; i <= lastNacked; i++ {
				msg, err2 := h.getChannelMessage(tx, cid, i)
				if err2 != nil {
					log.Warnln(err2)
					continue
				}
				if msg.GetCondPayRequest() != nil {
					payID := ctype.PayBytes2PayID(msg.GetCondPayRequest().GetCondPay())
					_, _, err = fsm.OnPayEgressUpdateAfterNack(tx, payID)
					if err != nil {
						log.Error(err)
						return err
					}
				}
				err = tx.DeleteChannelMessage(cid, i)
				if err != nil {
					log.Warnf("cannot delete NACKed msg %d for %x: %s", i, cid, err)
				}
			}
			from = lastNacked + 1
		}

		// process each acked message
		ackedMsgs := []*rpc.CelerMsg{}
		for i := from; i <= ackSimplex.GetSeqNum(); i++ {
			msg, err2 := h.getChannelMessage(tx, cid, i)
			if err2 != nil {
				log.Warnln(err2)
				continue
			}
			if msg.GetCondPayRequest() != nil {
				payID := ctype.PayBytes2PayID(msg.GetCondPayRequest().GetCondPay())
				directPay := msg.GetCondPayRequest().GetDirectPay()
				if directPay {
					_, _, err = fsm.OnPayEgressDirectCoSignedPaid(tx, payID, cid)
				} else {
					_, _, err = fsm.OnPayEgressDelivered(tx, payID)
				}
				if err != nil {
					log.Error(err)
					return err
				}
				log.Debugln("Receive ACK for cond pay request", payID.Hex(), "direct", directPay)
			} else if msg.GetPaymentSettleRequest() != nil {
				for _, pay := range msg.GetPaymentSettleRequest().GetSettledPays() {
					payID := ctype.Bytes2PayID(pay.GetSettledPayId())
					amt := new(big.Int).SetBytes(pay.GetAmount())
					paid := !((*amt).Cmp(new(big.Int).SetUint64(0)) == 0)
					if paid {
						_, _, err = fsm.OnPayEgressCoSignedPaid(tx, payID)
					} else {
						_, _, err = fsm.OnPayEgressCoSignedCanceled(tx, payID)
					}
					if err != nil {
						log.Error(err)
						return err
					}
					log.Debugln("Receive ACK pay settle request", payID.Hex(), "paid:", paid)
				}
			} else {
				log.Errorln("invalid message type", msg)
				continue
			}
			ackedMsgs = append(ackedMsgs, msg)
			err = tx.DeleteChannelMessage(cid, i)
			if err != nil {
				log.Warnf("cannot delete ACKed msg %d for %x: %s", i, cid, err)
			}
		}

		err = tx.PutSimplexState(cid, myAddr, ackState)
		if err != nil {
			return errors.New("put simplex failed:" + err.Error())
		}
		*retAckedMsgs = ackedMsgs
	}

	seqNums.LastAcked = ackSimplex.GetSeqNum()
	*retLastNackSeqNum = seqNums.LastNacked
	err = tx.PutChannelSeqNums(cid, seqNums)
	if err != nil {
		return errors.New("PutChannelSeqNums failed:" + err.Error())
	}

	return nil
}

func (h *CelerMsgHandler) getChannelMessage(tx *storage.DALTx, cid ctype.CidType, seq uint64) (*rpc.CelerMsg, error) {
	var err error
	msg, ok := h.messager.GetMsgQueue(cid, seq)
	if !ok {
		log.Warnf("cannot get msg %d for %x from msgQueue", seq, cid)
		msg, err = tx.GetChannelMessage(cid, seq)
		if err != nil {
			err = fmt.Errorf("cannot get msg %d for %x from database: %s", seq, cid, err)
			return nil, err
		}
	}
	return msg, nil
}

func (h *CelerMsgHandler) payFromSelf(pay *entity.ConditionalPay) bool {
	return bytes.Compare(pay.GetSrc(), h.nodeConfig.GetOnChainAddrBytes()) == 0
}
