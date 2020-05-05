// Copyright 2018-2020 Celer Network

package dispatchers

import (
	"sync"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/dispute"
	"github.com/celer-network/goCeler/handlers"
	"github.com/celer-network/goCeler/handlers/msghdl"
	"github.com/celer-network/goCeler/messager"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
)

type CooperativeWithdraw interface {
	ProcessRequest(*common.MsgFrame) error
	ProcessResponse(*common.MsgFrame) error
}

type CelerMsgDispatcher struct {
	stop                bool
	nodeConfig          common.GlobalNodeConfig
	streamWriter        common.StreamWriter
	signer              common.Signer
	monitorService      intfs.MonitorService
	dal                 *storage.DAL
	cooperativeWithdraw CooperativeWithdraw
	serverForwarder     handlers.ForwardToServerCallback
	onReceivingToken    event.OnReceivingTokenCallback
	tokenCallbackLock   *sync.RWMutex
	onSendingToken      event.OnSendingTokenCallback
	sendingCallbackLock *sync.RWMutex
	disputer            *dispute.Processor
	routeForwarder      *route.Forwarder
	routeController     *route.Controller
	messager            *messager.Messager
}

func NewCelerMsgDispatcher(
	nodeConfig common.GlobalNodeConfig,
	streamWriter common.StreamWriter,
	signer common.Signer,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	cooperativeWithdraw CooperativeWithdraw,
	serverForwarder handlers.ForwardToServerCallback,
	disputer *dispute.Processor,
	routeForwarder *route.Forwarder,
	routeController *route.Controller,
	messager *messager.Messager,
) *CelerMsgDispatcher {
	d := &CelerMsgDispatcher{
		nodeConfig:          nodeConfig,
		streamWriter:        streamWriter,
		signer:              signer,
		monitorService:      monitorService,
		dal:                 dal,
		stop:                false,
		cooperativeWithdraw: cooperativeWithdraw,
		serverForwarder:     serverForwarder,
		tokenCallbackLock:   &sync.RWMutex{},
		sendingCallbackLock: &sync.RWMutex{},
		disputer:            disputer,
		routeForwarder:      routeForwarder,
		routeController:     routeController,
		messager:            messager,
	}
	return d
}

var msgNames = []string{
	msghdl.CondPayRequestMsgName,
	msghdl.PaySettleProofMsgName,
	msghdl.PaySettleRequestMsgName,
	msghdl.HopAckStateMsgName,
	msghdl.RevealSecretMsgName,
	msghdl.RevealSecretAckMsgName,
	msghdl.CondPayReceiptMsgName,
	msghdl.CondPayResultMsgName,
	msghdl.WithdrawRequestMsgName,
	msghdl.WithdrawResponseMsgName,
	msghdl.UnkownMsgName,
}

func (d *CelerMsgDispatcher) OnReceivingToken(callback event.OnReceivingTokenCallback) {
	d.tokenCallbackLock.Lock()
	defer d.tokenCallbackLock.Unlock()
	d.onReceivingToken = callback
}
func (d *CelerMsgDispatcher) OnSendingToken(callback event.OnSendingTokenCallback) {
	d.sendingCallbackLock.Lock()
	defer d.sendingCallbackLock.Unlock()
	d.onSendingToken = callback
}

func (d *CelerMsgDispatcher) NewStream(peerAddr ctype.Addr) chan *rpc.CelerMsg {
	in := make(chan *rpc.CelerMsg)
	go d.Start(in, peerAddr)
	return in
}

func (d *CelerMsgDispatcher) Start(input chan *rpc.CelerMsg, peerAddr ctype.Addr) {
	// This dispatcher dispatch messages coming from one stream implementation
	log.Debug("CelerMsgDispatcher Running")
	for !d.stop {
		msg, ok := <-input
		if !ok {
			return
		}
		if msg.GetMessage() == nil {
			continue
		}
		log.Traceln("CelerMsg detail: ", msg)

		var handler handlers.CelerMsgHandler
		logEntry := pem.NewPem(d.nodeConfig.GetRPCAddr())
		logEntry.MsgFrom = ctype.Addr2Hex(peerAddr)
		msgFrame := &common.MsgFrame{
			Message:  msg,
			PeerAddr: peerAddr,
			LogEntry: logEntry,
		}

		handler = d.NewMsgHandler()

		start := time.Now()
		err := handler.Run(msgFrame)
		msgname := handler.GetMsgName()
		metrics.IncDispatcherMsgCnt(msgname)
		metrics.IncDispatcherMsgProcDur(start, msgname)
		if err != nil {
			logEntry.Error = append(logEntry.Error, err.Error())
			metrics.IncDispatcherErrCnt(msgname)
		}
		pem.CommitPem(logEntry)
	}
}

func (d *CelerMsgDispatcher) Stop() {
	d.stop = true
}

func (d *CelerMsgDispatcher) NewMsgHandler() *msghdl.CelerMsgHandler {
	return msghdl.NewCelerMsgHandler(
		d.nodeConfig,
		d.streamWriter,
		d.signer,
		d.monitorService,
		d.serverForwarder,
		d.onReceivingToken,
		d.tokenCallbackLock,
		d.onSendingToken,
		d.sendingCallbackLock,
		d.disputer,
		d.cooperativeWithdraw,
		d.routeForwarder,
		d.routeController,
		d.messager,
		d.dal,
	)
}
