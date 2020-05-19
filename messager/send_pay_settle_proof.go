// Copyright 2018-2020 Celer Network

package messager

import (
	"fmt"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
)

func (m *Messager) SendOnePaySettleProof(
	payID ctype.PayIDType,
	reason rpc.PaymentSettleReason,
	logEntry *pem.PayEventMessage) error {
	return m.SendPaysSettleProof([]ctype.PayIDType{payID}, reason, nil, logEntry)
}

func (m *Messager) SendPayUnreachableSettleProof(
	payID ctype.PayIDType,
	path *rpc.PayPath,
	logEntry *pem.PayEventMessage) error {
	return m.SendPaysSettleProof(
		[]ctype.PayIDType{payID}, rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE, []*rpc.PayPath{path}, logEntry)
}

func (m *Messager) SendPaysSettleProof(
	payIDs []ctype.PayIDType,
	reason rpc.PaymentSettleReason,
	payPaths []*rpc.PayPath,
	logEntry *pem.PayEventMessage) error {
	logEntry.SettleReason = reason
	if len(payIDs) == 0 {
		return fmt.Errorf("Empty pay ID list")
	}

	var peer ctype.Addr
	request := &rpc.PaymentSettleProof{}
	err := m.dal.Transactional(m.runPaySettleProofTx, payIDs, reason, payPaths, logEntry, &peer, &request)
	if err != nil {
		return err
	}
	celerMsg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_PaymentSettleProof{
			PaymentSettleProof: request,
		},
	}
	return m.ForwardCelerMsg(peer, celerMsg)
}

func (m *Messager) runPaySettleProofTx(tx *storage.DALTx, args ...interface{}) error {
	payIDs := args[0].([]ctype.PayIDType)
	reason := args[1].(rpc.PaymentSettleReason)
	payPaths := args[2].([]*rpc.PayPath)
	logEntry := args[3].(*pem.PayEventMessage)
	retPeer := args[4].(*ctype.Addr)
	retRequest := args[5].(**rpc.PaymentSettleProof)

	var rejecting bool
	if reason == rpc.PaymentSettleReason_PAY_REJECTED ||
		reason == rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE ||
		reason == rpc.PaymentSettleReason_PAY_EXPIRED {
		rejecting = true
	}

	request := &rpc.PaymentSettleProof{}
	var igcid ctype.CidType
	for i, payID := range payIDs {
		logEntry.PayIds = append(logEntry.PayIds, ctype.PayID2Hex(payID))

		var cid ctype.CidType
		var state int
		var found bool
		var err error
		if rejecting {
			cid, state, found, err = tx.GetPayIngress(payID)
		} else {
			cid, found, err = tx.GetPayIngressChannel(payID)
		}
		if err != nil {
			return fmt.Errorf("GetPayIngress err: %x, %w", payID, err)
		}
		if !found {
			return fmt.Errorf("%w, %x", common.ErrPayNotFound, payID)
		}
		if rejecting {
			err = fsm.OnPayIngressRejected(tx, payID, state)
			if err != nil {
				return fmt.Errorf("OnPayIngressRejected err: %x, %w", payID, err)
			}
		}
		if igcid == ctype.ZeroCid {
			igcid = cid
		} else if igcid != cid {
			return fmt.Errorf("cannot batch pay settle proof for multiple ingress cids")
		}
		settledPay := &rpc.SettledPayment{
			SettledPayId: payID.Bytes(),
			Reason:       reason,
		}
		if reason == rpc.PaymentSettleReason_PAY_DEST_UNREACHABLE {
			if len(payIDs) != len(payPaths) {
				log.Warnf("pay paths not provided for unreachable pay %x", payID)
			} else {
				settledPay.Path = payPaths[i]
			}
		}

		request.SettledPays = append(request.SettledPays, settledPay)
	}

	peer, found, err := tx.GetChanPeer(igcid)
	if err != nil {
		return fmt.Errorf("GetChanPeer err: %x, %w", igcid, err)
	}
	if !found {
		return fmt.Errorf("%x %w", igcid, common.ErrChannelNotFound)
	}

	logEntry.MsgTo = ctype.Addr2Hex(peer)
	logEntry.ToCid = ctype.Cid2Hex(igcid)

	*retPeer = peer
	*retRequest = request
	return nil
}

func (m *Messager) ForwardPaySettleProofMsg(frame *common.MsgFrame) error {
	msg := frame.Message
	logEntry := frame.LogEntry
	payProof := msg.GetPaymentSettleProof()
	var payID ctype.PayIDType
	if len(payProof.GetSettledPays()) > 0 {
		if len(payProof.SettledPays) > 1 {
			return fmt.Errorf("batched pay settle proof forwarding not supported yet")
		}
		payID = ctype.Bytes2PayID(payProof.SettledPays[0].SettledPayId)
		logEntry.PayId = ctype.PayID2Hex(payID)
	} else {
		return fmt.Errorf("empty settled pays in paymentSettleProof")
	}

	cid, peer, err := m.routeForwarder.LookupIngressChannelOnPay(payID)
	if err != nil {
		log.Errorln(err, payID)
		return err
	}
	logEntry.ToCid = ctype.Cid2Hex(cid)
	logEntry.MsgTo = ctype.Addr2Hex(peer)
	return m.streamWriter.WriteCelerMsg(peer, msg)
}
