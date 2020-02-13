// Copyright 2018-2019 Celer Network

package msghdl

import (
	"errors"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/utils"
)

func (h *CelerMsgHandler) HandleCondPayReceipt(frame *common.MsgFrame) error {
	receipt := frame.Message.GetCondPayReceipt()
	logEntry := frame.LogEntry
	if receipt == nil {
		return common.ErrInvalidMsgType
	}

	// Forward Msg if not destination
	dst := ctype.Bytes2Hex(frame.Message.GetToAddr())
	logEntry.Dst = dst
	payID := ctype.Bytes2PayID(receipt.PayId)
	logEntry.PayId = ctype.PayID2Hex(payID)
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.channelRouter.LookupIngressChannelOnPay(payID)
		if err != nil {
			log.Error(err)
			return errors.New(err.Error() + " LookupIngressChannelOnPay")
		}
		logEntry.MsgTo = peer
		log.Debugf("Forwarding cond pay receipt to %s, next hop %s", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	condPay, condPayBytes, err := h.dal.GetConditionalPay(payID)
	if err != nil {
		log.Warn("PAY NOT FOUND")
		return nil
	}

	// verify pay source
	if ctype.Bytes2Addr(condPay.GetSrc()) != ctype.Hex2Addr(h.nodeConfig.GetOnChainAddr()) {
		log.Errorln(common.ErrInvalidPaySrc, utils.PrintConditionalPay(condPay))
		return common.ErrInvalidPaySrc
	}

	// Check signature of receipt signed by destination
	if !h.crypto.SigIsValid(ctype.Bytes2Hex(condPay.Dest), condPayBytes, receipt.PayDestSig) {
		return errors.New("RECEIPT_NOT_SIGNED_BY_DEST")
	}

	// Return secret
	// The first condition is always HashLock
	if len(condPay.GetConditions()) == 0 {
		log.Warnln("empty condition list", payID.Hex())
		return nil
	}
	secretHash := ctype.Bytes2Hex(condPay.Conditions[0].GetHashLock())
	secret, err := h.dal.GetSecretRegistry(secretHash)
	if err != nil {
		log.Warn("SECRET NOT FOUND", secretHash)
		return nil
	}
	secretBytes := ctype.Hex2Bytes(secret)
	secretMsg := &rpc.RevealSecret{
		PayId:  receipt.PayId,
		Secret: secretBytes,
	}
	celerMsg := &rpc.CelerMsg{
		ToAddr: condPay.Dest,
		Message: &rpc.CelerMsg_RevealSecret{
			RevealSecret: secretMsg,
		},
	}
	err = h.streamWriter.WriteCelerMsg(ctype.Addr2Hex(frame.PeerAddr), celerMsg)
	if err != nil {
		log.Warn(err.Error())
		return err
	}

	return nil
}
