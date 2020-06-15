// Copyright 2018-2020 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/ptypes/any"
)

func (h *CelerMsgHandler) HandleRevealSecret(frame *common.MsgFrame) error {
	msg := frame.Message.GetRevealSecret()
	logEntry := frame.LogEntry
	if msg == nil {
		return common.ErrInvalidMsgType
	}
	dst := ctype.Bytes2Addr(frame.Message.GetToAddr())
	payID := ctype.Bytes2PayID(frame.Message.GetRevealSecret().PayId)
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.Secret = ctype.Bytes2Hex(msg.GetSecret())
	logEntry.Dst = ctype.Addr2Hex(dst)

	// Forward Msg if not destination
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.routeForwarder.LookupEgressChannelOnPay(payID)
		if err != nil {
			return fmt.Errorf("LookupEgressChannelOnPay err %w", err)
		}
		log.Debugf("Forwarding reveal secret to %x, next hop %x", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	secret := msg.GetSecret()
	var pay *entity.ConditionalPay
	var note *any.Any
	err := h.dal.Transactional(h.recvSecretTx, payID, secret, &pay, &note)
	if err != nil {
		if errors.Is(err, common.ErrPayOffChainResolved) {
			log.Warnln(err, payID.Hex())
			return nil
		}
		return err
	}

	// send RevealSecretAck
	secretSig, err := h.signer.SignEthMessage(secret)
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}
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
	err = h.streamWriter.WriteCelerMsg(frame.PeerAddr, celerMsg)
	if err != nil {
		log.Error(err)
	}

	// Notify application of cond pay receiving.
	h.tokenCallbackLock.RLock()
	if h.onReceivingToken != nil {
		go h.onReceivingToken.HandleReceivingStart(payID, pay, note)
	}
	h.tokenCallbackLock.RUnlock()

	return nil
}

func (h *CelerMsgHandler) recvSecretTx(tx *storage.DALTx, args ...interface{}) error {
	payID := args[0].(ctype.PayIDType)
	secret := args[1].([]byte)
	retPay := args[2].(**entity.ConditionalPay)
	retNote := args[3].(**any.Any)

	pay, note, inState, found, err := tx.GetPayForRecvSecret(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return common.ErrPayNotFound
	}
	*retPay = pay
	*retNote = note

	err = fsm.OnPayIngressSecretRevealed(tx, payID, inState)
	if err != nil {
		return err
	}

	// verify pay destination or I'm delegating the pay.
	delegateStatus, found, err := tx.GetDelegatedPayStatus(payID)
	if err != nil {
		return fmt.Errorf("GetDelegatedPayStatus err %w", err)
	}
	delegatingPay := (found && delegateStatus == structs.DelegatedPayStatus_RECVING)
	if ctype.Bytes2Addr(pay.GetDest()) != h.nodeConfig.GetOnChainAddr() && !delegatingPay {
		return fmt.Errorf("delegatingPay err %w", common.ErrInvalidPayDst)
	}

	// verify hash preimage
	hash := crypto.Keccak256(secret)
	if len(pay.GetConditions()) == 0 {
		return common.ErrZeroConditions
	}
	if bytes.Compare(hash, pay.Conditions[0].GetHashLock()) != 0 {
		return fmt.Errorf("hash lock verification failed")
	}
	log.Debugf("Saving secret(%x) of hash(%x) for pay(%x)", secret, hash, payID)
	err = tx.InsertSecret(ctype.Bytes2Hex(hash), ctype.Bytes2Hex(secret), payID)
	if err != nil {
		return fmt.Errorf("InsertSecret %x err: %w", secret, err)
	}

	return nil
}

func (h *CelerMsgHandler) HandleRevealSecretAck(frame *common.MsgFrame) error {
	ack := frame.Message.GetRevealSecretAck()
	logEntry := frame.LogEntry
	if ack == nil {
		return common.ErrInvalidMsgType
	}
	dst := ctype.Bytes2Addr(frame.Message.GetToAddr())
	payID := ctype.Bytes2PayID(ack.GetPayId())
	logEntry.PayId = ctype.PayID2Hex(payID)

	// Forward Msg if not destination
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.routeForwarder.LookupIngressChannelOnPay(payID)
		if err != nil {
			return fmt.Errorf("LookupIngressChannelOnPay err %w", err)
		}
		log.Debugf("Forwarding reveal secret ack to %x, next hop %x", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	var pay *entity.ConditionalPay
	err := h.dal.Transactional(h.recvSecretAckTx, payID, ack, &pay)
	if err != nil {
		if errors.Is(err, common.ErrPayOffChainResolved) {
			log.Warnln(err, payID.Hex())
			return nil
		}
		return err
	}

	// HL only, cPay
	if len(pay.GetConditions()) == 1 &&
		pay.Conditions[0].ConditionType == entity.ConditionType_HASH_LOCK {
		amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
		return h.messager.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, frame.LogEntry)
	}
	// TODO: notify client on pay state change

	return nil
}

func (h *CelerMsgHandler) recvSecretAckTx(tx *storage.DALTx, args ...interface{}) error {
	payID := args[0].(ctype.PayIDType)
	ack := args[1].(*rpc.RevealSecretAck)
	retPay := args[2].(**entity.ConditionalPay)

	pay, _, outState, found, err := tx.GetPayAndEgressState(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return common.ErrPayNotFound
	}
	*retPay = pay

	err = fsm.OnPayEgressSecretRevealed(tx, payID, outState)
	if err != nil {
		return err
	}

	// verify pay source
	if ctype.Bytes2Addr(pay.GetSrc()) != h.nodeConfig.GetOnChainAddr() {
		return common.ErrInvalidPaySrc
	}
	// verify signature
	if len(pay.GetConditions()) == 0 {
		return fmt.Errorf("empty condition list")
	}
	hash := ctype.Bytes2Hex(pay.Conditions[0].GetHashLock())
	secret, found, err := tx.GetSecret(hash)
	if err != nil {
		return fmt.Errorf("GetSecret err %w hash %x", err, hash)
	}
	if !found {
		return fmt.Errorf("%w, hash %x", common.ErrSecretNotRevealed, hash)
	}

	delegator, found, err := tx.GetPayDelegator(payID)
	if err != nil {
		return fmt.Errorf("GetPayDelegator err %w", err)
	}
	expectedSigner := ctype.Bytes2Addr(pay.Dest)
	if found && delegator != ctype.ZeroAddr {
		expectedSigner = delegator
	}
	if !eth.IsSignatureValid(expectedSigner, ctype.Hex2Bytes(secret), ack.GetPayDestSecretSig()) {
		return common.ErrInvalidSig
	}

	return nil
}
