// Copyright 2018-2020 Celer Network

package msghdl

import (
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
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
		return fmt.Errorf("Vouched pay result not accepted yet")
	}
	if len(payProof.SettledPays) == 0 {
		return fmt.Errorf("Empty payment settle proof")
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
				return fmt.Errorf("batched pay settle proof with different reasons not supported")
			}
			payID := ctype.Bytes2PayID(sp.SettledPayId)
			logEntry.PayId = ctype.PayID2Hex(payID)
			logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))
			var found bool
			var egressPeer ctype.Addr
			pay, egressPeer, found, err = h.dal.GetPayForRecvSettleProof(payID)
			if err != nil {
				log.Errorln(err, payID.Hex())
				continue
			}
			if !found {
				log.Errorln(common.ErrPayNotFound, payID.Hex())
				continue
			}
			if egressPeer != peer {
				return fmt.Errorf("settle proof sender and egress peer not match")
			}
			expiredPays = append(expiredPays, pay)
			payAmts = append(payAmts, new(big.Int).SetUint64(0))
		}
		if len(expiredPays) == 0 {
			return fmt.Errorf("no valid expired pays to settle")
		}
		_, err = h.messager.SendPaysSettleRequest(expiredPays, payAmts, reason, logEntry)
		if err != nil {
			err = fmt.Errorf("SendPaysSettleRequest err: %w", err)
		}
		// do not foward settle proof for expired pays.
		// each pair of peers are responsible for settling expired pays among themselves
		return err
	}

	if len(payProof.SettledPays) > 1 {
		return fmt.Errorf("Batched pay settle proof for unexpired pays not supported")
	}

	payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
	logEntry.PayId = ctype.PayID2Hex(payID)
	var found bool
	var ingressPeer, egressPeer ctype.Addr
	pay, egressPeer, found, err = h.dal.GetPayForRecvSettleProof(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return common.ErrPayNotFound
	}
	if egressPeer != peer {
		return fmt.Errorf("settle proof sender and egress peer not match")
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
		payPath := settledPay.GetPath()
		logEntry.PayPath = utils.PrintPayPath(payPath, payID)
		if h.payFromSelf(pay) {
			h.notifyUnreachablility(payID, pay)
		} else {
			ingressPeer, found, err = h.dal.GetPayIngressPeer(payID)
			if err != nil {
				return fmt.Errorf("GetPayIngressPeer err: %w", err)
			}
			if !found {
				return fmt.Errorf("GetPayIngressPeer err: %w", common.ErrPayNoIngress)
			}
			payHop := &rpc.PayHop{
				PayId:       payID.Bytes(),
				PrevHopAddr: ingressPeer.Bytes(),
				NextHopAddr: egressPeer.Bytes(),
			}
			err = h.prependPayPath(payPath, payHop)
			if err != nil {
				return err
			}
			err = h.dal.PutPayPath(payID, payPath)
			if err != nil {
				return fmt.Errorf("PutPayPath err: %w", err)
			}
		}
	default:
		return fmt.Errorf("Unsupported payment settle type")
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
