// Copyright 2018-2019 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

func (h *CelerMsgHandler) HandleRevealSecret(frame *common.MsgFrame) error {
	msg := frame.Message.GetRevealSecret()
	logEntry := frame.LogEntry
	if msg == nil {
		return common.ErrInvalidMsgType
	}

	// Forward Msg if not destination
	dst := ctype.Bytes2Hex(frame.Message.GetToAddr())
	payID := ctype.Bytes2PayID(frame.Message.GetRevealSecret().PayId)
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.Secret = ctype.Bytes2Hex(msg.GetSecret())
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.channelRouter.LookupEgressChannelOnPay(payID)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debugf("Forwarding reveal secret to %s, next hop %s", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	secret := msg.GetSecret()
	pay, _, err := h.dal.GetConditionalPay(payID)
	if err != nil {
		log.Error("Can't find cond pay")
		return err
	}
	// verify pay destination
	if ctype.Bytes2Addr(pay.GetDest()) != ctype.Hex2Addr(h.nodeConfig.GetOnChainAddr()) {
		log.Errorln(common.ErrInvalidPayDst, utils.PrintConditionalPay(pay))
		return common.ErrInvalidPayDst
	}

	// verify hash preimage
	hash := crypto.Keccak256(secret)
	if len(pay.GetConditions()) == 0 {
		return errors.New("empty condition list " + payID.Hex())
	}
	if bytes.Compare(hash, pay.Conditions[0].GetHashLock()) != 0 {
		err = errors.New("hash lock verification failed for pay")
		log.Errorln(err, payID.Hex())
		return err
	}
	secretHash := ctype.Bytes2Hex(pay.Conditions[0].GetHashLock())
	preimage := ctype.Bytes2Hex(secret)
	log.Debugf("Saving secret(%s) of secretHash(%s) for pay(%s)", preimage, secretHash, payID.Hex())
	h.dal.PutSecretRegistry(secretHash, preimage)

	// send RevealSecretAck
	secretSig, _ := h.crypto.Sign(secret)
	ack := &rpc.RevealSecretAck{
		PayId:            payID[:],
		PayDestSecretSig: secretSig,
	}
	celerMsg := &rpc.CelerMsg{
		ToAddr: pay.Src,
		Message: &rpc.CelerMsg_RevealSecretAck{
			RevealSecretAck: ack,
		},
	}
	h.streamWriter.WriteCelerMsg(ctype.Addr2Hex(frame.PeerAddr), celerMsg)

	err = h.dal.Transactional(ingressLockRevealedTx, payID)
	if err != nil {
		log.Error(err)
	}

	note, err := h.dal.GetPayNote(payID)
	if err != nil {
		// Missing note is normal for most cpay.
		log.Traceln(err)
	}
	// Notify application of cond pay receiving.
	h.tokenCallbackLock.RLock()
	if h.onReceivingToken != nil {
		go h.onReceivingToken.HandleReceivingStart(payID, pay, note)
	}
	h.tokenCallbackLock.RUnlock()

	return nil
}

func (h *CelerMsgHandler) HandleRevealSecretAck(frame *common.MsgFrame) error {
	ack := frame.Message.GetRevealSecretAck()
	logEntry := frame.LogEntry
	if ack == nil {
		return common.ErrInvalidMsgType
	}

	// Forward Msg if not destination
	dst := ctype.Bytes2Hex(frame.Message.GetToAddr())
	payID := ctype.Bytes2PayID(ack.GetPayId())
	logEntry.PayId = ctype.PayID2Hex(payID)
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.channelRouter.LookupIngressChannelOnPay(payID)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debugf("Forwarding reveal secret ack to %s, next hop %s", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	pay, _, err := h.dal.GetConditionalPay(payID)
	if err != nil {
		return err
	}
	// verify pay source
	if ctype.Bytes2Addr(pay.GetSrc()) != ctype.Hex2Addr(h.nodeConfig.GetOnChainAddr()) {
		log.Errorln(common.ErrInvalidPaySrc, utils.PrintConditionalPay(pay))
		return common.ErrInvalidPaySrc
	}
	// verify signature
	if len(pay.GetConditions()) == 0 {
		return errors.New("empty condition list " + payID.Hex())
	}
	secret, err := h.dal.GetSecretRegistry(ctype.Bytes2Hex(pay.Conditions[0].GetHashLock()))
	if err != nil {
		log.Error("SECRET NOT FOUND", payID.Hex())
		return err
	}
	if !h.crypto.SigIsValid(ctype.Bytes2Hex(pay.Dest), ctype.Hex2Bytes(secret), ack.GetPayDestSecretSig()) {
		log.Errorln(common.ErrInvalidSig, payID.Hex())
		return common.ErrInvalidSig
	}
	// HL only, cPay
	if len(pay.GetConditions()) == 1 &&
		pay.Conditions[0].ConditionType == entity.ConditionType_HASH_LOCK {
		amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
		return h.messager.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, frame.LogEntry)
	}

	err = h.dal.Transactional(egressLockRevealedTx, payID)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func ingressLockRevealedTx(tx *storage.DALTx, args ...interface{}) error {
	payID := args[0].(ctype.PayIDType)
	_, _, err := fsm.OnPayIngressHashLockRevealed(tx, payID)
	return err
}

func egressLockRevealedTx(tx *storage.DALTx, args ...interface{}) error {
	payID := args[0].(ctype.PayIDType)
	_, _, err := fsm.OnPayEgressHashLockRevealed(tx, payID)
	return err
}
