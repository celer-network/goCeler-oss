// Copyright 2018-2020 Celer Network

package msghdl

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goutils/log"
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
			var found bool
			pay, _, found, err = h.dal.GetPayment(payID)
			if err != nil {
				log.Errorln(err, payID.Hex())
				continue
			}
			if !found {
				log.Errorln(common.ErrPayNotFound, payID.Hex())
				continue
			}
			expiredPays = append(expiredPays, pay)
			payAmts = append(payAmts, new(big.Int).SetUint64(0))
		}
		_, err = h.messager.SendPaysSettleRequest(expiredPays, payAmts, reason, logEntry)
		if err != nil {
			err = fmt.Errorf("SendPaysSettleRequest err: %w", err)
		}
		// do not foward settle proof for expired pays.
		// Currently, we only consider single-account (multi-server) OSP.  Clients will actively call
		// SettleExpiredPays() on demand (e.g., periodically or when start)
		// TODO: handle multi-account OSP cases
		return err
	}

	if len(payProof.SettledPays) > 1 {
		return errors.New("Batched pay settle proof for unexpired pays not supported")
	}

	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	logEntry.PayId = ctype.PayID2Hex(payID)
	var found bool
	pay, _, found, err = h.dal.GetPayment(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return common.ErrPayNotFound
	}
	payAmt := new(big.Int).SetUint64(0)
	switch reason {
	case rpc.PaymentSettleReason_PAY_REJECTED:
	case rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN:
		payAmt, _, err = h.disputer.GetCondPayInfoFromRegistry(payID)
		if err != nil {
			return fmt.Errorf("GetCondPayInfoFromRegistry err: %w", err)
		}
	case rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE:
		h.notifyUnreachablility(payID, pay)
	default:
		return errors.New("Unsupported payment settle type")
	}

	err = h.messager.SendOnePaySettleRequest(pay, payAmt, reason, logEntry)
	if err != nil {
		return fmt.Errorf("SendOnePaySettleRequest err: %w", err)
	}
	return nil
}

func (h *CelerMsgHandler) notifyUnreachablility(payID ctype.PayIDType, pay *entity.ConditionalPay) {
	note, _, err := h.dal.GetPayNote(payID)
	if err != nil {
		log.Error(err)
	}
	h.sendingCallbackLock.RLock()
	if h.onSendingToken != nil {
		go h.onSendingToken.HandleDestinationUnreachable(payID, pay, note)
	}
	h.sendingCallbackLock.RUnlock()
}
