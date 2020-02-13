// Copyright 2018-2019 Celer Network

package messager

import (
	"errors"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
)

func (m *Messager) SendOnePaySettleProof(
	payID ctype.PayIDType,
	reason rpc.PaymentSettleReason,
	logEntry *pem.PayEventMessage) error {
	return m.SendPaysSettleProof([]ctype.PayIDType{payID}, reason, logEntry)
}

func (m *Messager) SendPaysSettleProof(
	payIDs []ctype.PayIDType,
	reason rpc.PaymentSettleReason,
	logEntry *pem.PayEventMessage) error {
	if len(payIDs) == 0 {
		return errors.New("Empty pay ID list")
	}

	cid, peer, err := m.channelRouter.LookupIngressChannelOnPay(payIDs[0])
	if err != nil {
		log.Errorln(err, payIDs[0].Hex())
		return err
	}
	logEntry.MsgTo = peer
	logEntry.ToCid = ctype.Cid2Hex(cid)

	for i := 1; i < len(payIDs); i++ {
		cid2, _, _, err2 := m.dal.GetPayIngressState(payIDs[i])
		if err2 != nil {
			log.Error(err2)
			return err2
		}
		if cid != cid2 {
			return errors.New("cannot batch settle proof requests for pays with different ingress cids")
		}
	}
	logEntry.SettleReason = reason

	request := &rpc.PaymentSettleProof{}
	for _, payID := range payIDs {
		settledPay := &rpc.SettledPayment{
			SettledPayId: payID.Bytes(),
			Reason:       reason,
		}
		request.SettledPays = append(request.SettledPays, settledPay)
		logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))

		// if reason is pay rejected or dest unreachable, set pay ingress to rejected to record this status
		if reason == rpc.PaymentSettleReason_PAY_REJECTED ||
			reason == rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE ||
			reason == rpc.PaymentSettleReason_PAY_EXPIRED {
			err := m.dal.Transactional(m.runPaySettleProofTx, payID)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_PaymentSettleProof{
			PaymentSettleProof: request,
		},
	}
	return m.ForwardCelerMsg(peer, celerMsg)
}

func (m *Messager) runPaySettleProofTx(tx *storage.DALTx, args ...interface{}) error {
	payID := args[0].(ctype.PayIDType)

	_, _, err := fsm.OnPayIngressRejected(tx, payID)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (m *Messager) ForwardPaySettleProofMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	payProof := msg.GetPaymentSettleProof()
	var payID ctype.PayIDType
	if len(payProof.GetSettledPays()) > 0 {
		if len(payProof.SettledPays) > 1 {
			return errors.New("batched pay settle proof forwarding not supported yet")
		}
		payID = ctype.Bytes2PayID(payProof.SettledPays[0].SettledPayId)
		logEntry.PayId = ctype.PayID2Hex(payID)
	} else {
		return errors.New("empty settled pays in paymentSettleProof")
	}

	cid, peer, err := m.channelRouter.LookupIngressChannelOnPay(payID)
	if err != nil {
		log.Errorln(err, payID)
		return err
	}
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.MsgTo = peer
	return m.streamWriter.WriteCelerMsg(peer, msg)
}
