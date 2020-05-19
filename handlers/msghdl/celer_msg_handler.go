// Copyright 2018-2020 Celer Network

package msghdl

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/dispute"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/handlers"
	"github.com/celer-network/goCeler/messager"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

type CooperativeWithdraw interface {
	ProcessRequest(*common.MsgFrame) error
	ProcessResponse(*common.MsgFrame) error
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
	RoutingRequestMsgName   = "RoutingRequestMessage"
	UnkownMsgName           = "UnkownMessage"
)

type CelerMsgHandler struct {
	nodeConfig          common.GlobalNodeConfig
	streamWriter        common.StreamWriter
	signer              common.Signer
	monitorService      intfs.MonitorService
	serverForwarder     handlers.ForwardToServerCallback
	onReceivingToken    event.OnReceivingTokenCallback
	tokenCallbackLock   *sync.RWMutex
	onSendingToken      event.OnSendingTokenCallback
	sendingCallbackLock *sync.RWMutex
	disputer            *dispute.Processor
	cooperativeWithdraw CooperativeWithdraw
	routeForwarder      *route.Forwarder
	routeController     *route.Controller
	messager            *messager.Messager
	dal                 *storage.DAL
	msgName             string
}

func NewCelerMsgHandler(
	nodeConfig common.GlobalNodeConfig,
	streamWriter common.StreamWriter,
	signer common.Signer,
	monitorService intfs.MonitorService,
	serverForwarder handlers.ForwardToServerCallback,
	onReceivingToken event.OnReceivingTokenCallback,
	tokenCallbackLock *sync.RWMutex,
	onSendingToken event.OnSendingTokenCallback,
	sendingCallbackLock *sync.RWMutex,
	disputer *dispute.Processor,
	cooperativeWithdraw CooperativeWithdraw,
	routeForwarder *route.Forwarder,
	routeController *route.Controller,
	messager *messager.Messager,
	dal *storage.DAL,
) *CelerMsgHandler {
	h := &CelerMsgHandler{
		nodeConfig:          nodeConfig,
		streamWriter:        streamWriter,
		signer:              signer,
		monitorService:      monitorService,
		serverForwarder:     serverForwarder,
		onReceivingToken:    onReceivingToken,
		tokenCallbackLock:   tokenCallbackLock,
		onSendingToken:      onSendingToken,
		sendingCallbackLock: sendingCallbackLock,
		disputer:            disputer,
		cooperativeWithdraw: cooperativeWithdraw,
		routeForwarder:      routeForwarder,
		routeController:     routeController,
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
		err = h.cooperativeWithdraw.ProcessRequest(frame)
	case *rpc.CelerMsg_WithdrawResponse:
		h.msgName = WithdrawResponseMsgName
		frame.LogEntry.Type = pem.PayMessageType_WITHDRAW_RESPONSE
		err = h.cooperativeWithdraw.ProcessResponse(frame)
	case *rpc.CelerMsg_RoutingRequest:
		h.msgName = RoutingRequestMsgName
		frame.LogEntry.Type = pem.PayMessageType_ROUTING_REQUEST
		err = h.HandleRoutingRequest(frame)
	default:
		h.msgName = UnkownMsgName
		log.Errorln("Can't find hop handler for", frame.Message, ctype.Addr2Hex(frame.PeerAddr))
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

func (h *CelerMsgHandler) payFromSelf(pay *entity.ConditionalPay) bool {
	return bytes.Compare(pay.GetSrc(), h.nodeConfig.GetOnChainAddr().Bytes()) == 0
}

func (h *CelerMsgHandler) prependPayPath(payPath *rpc.PayPath, payHop *rpc.PayHop) error {
	payHopBytes, err := proto.Marshal(payHop)
	if err != nil {
		return fmt.Errorf("marshal payHop err: %w", err)
	}
	sig, err := h.signer.SignEthMessage(payHopBytes)
	if err != nil {
		return fmt.Errorf("sign payHop err: %w", err)
	}
	signedPayHop := &rpc.SignedPayHop{
		PayHopBytes: payHopBytes,
		Sig:         sig,
	}
	payPath.Hops = append([]*rpc.SignedPayHop{signedPayHop}, payPath.Hops...)
	return nil
}
