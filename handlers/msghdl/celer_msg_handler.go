// Copyright 2018-2019 Celer Network

package msghdl

import (
	"sync"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/dispute"
	"github.com/celer-network/goCeler-oss/handlers"
	"github.com/celer-network/goCeler-oss/messager"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
)

type CooperativeWithdraw interface {
	ProcessRequest(*rpc.CooperativeWithdrawRequest) error
	ProcessResponse(*rpc.CooperativeWithdrawResponse) error
}

const (
	CondPayRequestMsgName   = "CondPayRequestMessage"
	PaySettleProofMsgName   = "PaySettleProofMessage"
	PaySettleRequestMsgName = "PaySettleRequestMessage"
	HopAckStateMsgName      = "HopAckStateMessage"
	RevealSecretMsgName     = "RevealSecretMessage"
	RevealSecretAckMsgName  = "RevealSecretAckMessage"
	CondPayReceiptMsgName   = "CondPayReceiptMessage"
	CondPayResultMsgName    = "CondPayResultMessage"
	WithdrawRequestMsgName  = "WithdrawRequestMessage"
	WithdrawResponseMsgName = "WithdrawResponseMessage"
	UnkownMsgName           = "UnkownMessage"
)

type CelerMsgHandler struct {
	nodeConfig          common.GlobalNodeConfig
	streamWriter        common.StreamWriter
	crypto              common.Crypto
	channelRouter       common.StateChannelRouter
	monitorService      intfs.MonitorService
	serverForwarder     handlers.ForwardToServerCallback
	onReceivingToken    event.OnReceivingTokenCallback
	tokenCallbackLock   *sync.RWMutex
	onSendingToken      event.OnSendingTokenCallback
	sendingCallbackLock *sync.RWMutex
	disputer            *dispute.Processor
	cooperativeWithdraw CooperativeWithdraw
	messager            messager.Sender
	dal                 *storage.DAL
	msgName             string
}

func NewCelerMsgHandler(
	nodeConfig common.GlobalNodeConfig,
	streamWriter common.StreamWriter,
	crypto common.Crypto,
	channelRouter common.StateChannelRouter,
	monitorService intfs.MonitorService,
	serverForwarder handlers.ForwardToServerCallback,
	onReceivingToken event.OnReceivingTokenCallback,
	tokenCallbackLock *sync.RWMutex,
	onSendingToken event.OnSendingTokenCallback,
	sendingCallbackLock *sync.RWMutex,
	disputer *dispute.Processor,
	cooperativeWithdraw CooperativeWithdraw,
	messager messager.Sender,
	dal *storage.DAL,
) *CelerMsgHandler {
	h := &CelerMsgHandler{
		nodeConfig:          nodeConfig,
		streamWriter:        streamWriter,
		crypto:              crypto,
		channelRouter:       channelRouter,
		monitorService:      monitorService,
		serverForwarder:     serverForwarder,
		onReceivingToken:    onReceivingToken,
		tokenCallbackLock:   tokenCallbackLock,
		onSendingToken:      onSendingToken,
		sendingCallbackLock: sendingCallbackLock,
		disputer:            disputer,
		cooperativeWithdraw: cooperativeWithdraw,
		messager:            messager,
		dal:                 dal,
	}
	return h
}

func (h *CelerMsgHandler) Run(frame *common.MsgFrame) error {
	var err error
	switch frame.Message.GetMessage().(type) {
	case *rpc.CelerMsg_CondPayRequest:
		h.msgName = CondPayRequestMsgName
		frame.LogEntry.Type = pem.PayMessageType_COND_PAY_REQUEST
		err = h.HandleCondPayRequest(frame)
	case *rpc.CelerMsg_PaymentSettleProof:
		h.msgName = PaySettleProofMsgName
		frame.LogEntry.Type = pem.PayMessageType_PAY_SETTLE_PROOF
		err = h.HandlePaySettleProof(frame)
	case *rpc.CelerMsg_PaymentSettleRequest:
		h.msgName = PaySettleRequestMsgName
		frame.LogEntry.Type = pem.PayMessageType_PAY_SETTLE_REQUEST
		err = h.HandlePaySettleRequest(frame)
	case *rpc.CelerMsg_CondPayResponse:
		h.msgName = HopAckStateMsgName
		frame.LogEntry.Type = pem.PayMessageType_COND_PAY_RESPONSE
		err = h.HandleHopAckState(frame)
	case *rpc.CelerMsg_PaymentSettleResponse:
		h.msgName = HopAckStateMsgName
		frame.LogEntry.Type = pem.PayMessageType_PAY_SETTLE_RESPONSE
		err = h.HandleHopAckState(frame)
	case *rpc.CelerMsg_CondPayReceipt:
		h.msgName = CondPayReceiptMsgName
		frame.LogEntry.Type = pem.PayMessageType_COND_PAY_RECEIPT
		err = h.HandleCondPayReceipt(frame)
	case *rpc.CelerMsg_RevealSecret:
		h.msgName = RevealSecretMsgName
		frame.LogEntry.Type = pem.PayMessageType_REVEAL_SECRET
		err = h.HandleRevealSecret(frame)
	case *rpc.CelerMsg_RevealSecretAck:
		h.msgName = RevealSecretAckMsgName
		frame.LogEntry.Type = pem.PayMessageType_REVEAL_SECRET_ACK
		err = h.HandleRevealSecretAck(frame)
	case *rpc.CelerMsg_WithdrawRequest:
		h.msgName = WithdrawRequestMsgName
		frame.LogEntry.Type = pem.PayMessageType_WITHDRAW_REQUEST
		err = h.cooperativeWithdraw.ProcessRequest(frame.Message.GetWithdrawRequest())
	case *rpc.CelerMsg_WithdrawResponse:
		h.msgName = WithdrawResponseMsgName
		frame.LogEntry.Type = pem.PayMessageType_WITHDRAW_RESPONSE
		err = h.cooperativeWithdraw.ProcessResponse(frame.Message.GetWithdrawResponse())
	default:
		h.msgName = UnkownMsgName
		log.Errorln("Can't find hop handler for", frame.Message, frame.PeerAddr.Hex())
		err = common.ErrInvalidMsgType
	}
	return err
}

func (h *CelerMsgHandler) GetMsgName() string {
	return h.msgName
}

// -------------------------- Helper util functions ---------------------------

func validRecvdSeqNum(stored, recvd, base uint64) bool {
	return stored == base && recvd > stored
}
