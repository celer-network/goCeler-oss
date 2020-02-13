// Copyright 2018-2019 Celer Network

package msghdl

import (
	"errors"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rpc"
)

func (h *CelerMsgHandler) HandlePaySettleProof(frame *common.MsgFrame) error {
	payProof := frame.Message.GetPaymentSettleProof()
	if payProof == nil {
		return common.ErrInvalidMsgType
	}

	peer := frame.PeerAddr
	logEntry := frame.LogEntry
	var pay *entity.ConditionalPay
	var err error
	if len(payProof.VouchedCondPayResults) > 0 {
		log.Debugln("received settle proof with vouched results from", ctype.Addr2Hex(peer))
		return errors.New("Vouched pay result not accepted yet")
	}
	if len(payProof.SettledPays) == 0 {
		return errors.New("Empty payment settle proof")
	}

	log.Debugln("received settle proof with settled pays from", ctype.Addr2Hex(peer))
	settledPay := payProof.SettledPays[0]
	reason := settledPay.GetReason()
	logEntry.SettleReason = reason
	if reason == rpc.PaymentSettleReason_PAY_EXPIRED {
		// no need to verify if pay already expired,
		// it's the hop receiver's responsibility to check before ask for canceling expired pays
		var expiredPays []*entity.ConditionalPay
		var payAmts []*big.Int
		for _, sp := range payProof.SettledPays {
			if sp.Reason != rpc.PaymentSettleReason_PAY_EXPIRED {
				return errors.New("batched pay settle proof with different reasons not supported")
			}
			payID := ctype.Bytes2PayID(sp.SettledPayId)
			logEntry.PayId = ctype.PayID2Hex(payID)
			logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
			pay, _, err = h.dal.GetConditionalPay(payID)
			if err != nil {
				log.Errorln("Cannot find pay", payID.Hex(), err)
				continue
			}
			expiredPays = append(expiredPays, pay)
			payAmts = append(payAmts, new(big.Int).SetUint64(0))
		}
		_, err = h.messager.SendPaysSettleRequest(expiredPays, payAmts, reason, logEntry)
		if err != nil {
			log.Error(err)
			err = errors.New(err.Error() + " SendPaysSettleRequest")
		}
		// do not foward settle proof for expired pays.
		// Currently, we only consider single-account (multi-server) OSP.  Clients will actively call
		// SettleExpiredPays() on demand (e.g., periodically or when start)
		return err
	}
	if len(payProof.SettledPays) > 1 {
		return errors.New("Batched pay settle proof for unexpired pays not supported")
	}
	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	pay, _, err = h.dal.GetConditionalPay(payID)
	if err != nil {
		log.Errorln("Cannot find pay", payID.Hex(), err)
		return errors.New(err.Error() + " GetCondPayId " + payID.Hex())
	}
	payAmt := new(big.Int).SetUint64(0)
	switch reason {
	case rpc.PaymentSettleReason_PAY_REJECTED:
	case rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN:
		payAmt, _, err = h.disputer.GetCondPayInfoFromRegistry(payID)
		if err != nil {
			log.Errorln("Get info from PayRegistry error", payID.Hex(), err)
			return errors.New(err.Error() + " GetCondPayInfoFromRegistry")
		}
	case rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		h.notifyUnreachablility(payID, pay)
	default:
		return errors.New("Unsupported payment settle type")
	}

	err = h.messager.SendOnePaySettleRequest(pay, payAmt, reason, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error()+" SendOnePaySettleRequest"+payID.Hex())
		log.Error(err)
		return errors.New(err.Error() + " SendOnePaySettleRequest")
	}
	if h.nodeConfig.GetOnChainAddr() != ctype.Bytes2Hex(pay.Src) {
		if reason == rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN {
			_, peer, err2 := h.channelRouter.LookupIngressChannelOnPay(ctype.Pay2PayID(pay))
			if err2 != nil {
				log.Error(err2)
				return errors.New(err2.Error() + " LookupIngressChannelOnPay")
			}
			err2 = h.messager.ForwardCelerMsg(peer, frame.Message)
			if err2 != nil {
				logEntry.Error = append(logEntry.Error, err2.Error()+" ForwardCelerMsg"+peer)
				log.Errorln("Foward pay settle proof error", err2, "peer", peer, "reason", reason)
			}
		}
	}
	return nil
}

func (h *CelerMsgHandler) notifyUnreachablility(payID ctype.PayIDType, pay *entity.ConditionalPay) {
	note, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Traceln(err)
	}
	h.sendingCallbackLock.RLock()
	if h.onSendingToken != nil {
		go h.onSendingToken.HandleDestinationUnreachable(payID, pay, note)
	}
	h.sendingCallbackLock.RUnlock()
}
