// Copyright 2018-2019 Celer Network

package dispatchers

import (
	"expvar"
	"sync"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/dispute"
	"github.com/celer-network/goCeler-oss/handlers"
	"github.com/celer-network/goCeler-oss/handlers/msghdl"
	"github.com/celer-network/goCeler-oss/messager"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/zserge/metric"
)

type CooperativeWithdraw interface {
	ProcessRequest(*rpc.CooperativeWithdrawRequest) error
	ProcessResponse(*rpc.CooperativeWithdrawResponse) error
}

type CelerMsgDispatcher struct {
	stop                bool
	nodeConfig          common.GlobalNodeConfig
	streamWriter        common.StreamWriter
	crypto              common.Crypto
	channelRouter       common.StateChannelRouter
	monitorService      intfs.MonitorService
	dal                 *storage.DAL
	cooperativeWithdraw CooperativeWithdraw
	serverForwarder     handlers.ForwardToServerCallback
	onReceivingToken    event.OnReceivingTokenCallback
	tokenCallbackLock   *sync.RWMutex
	onSendingToken      event.OnSendingTokenCallback
	sendingCallbackLock *sync.RWMutex
	disputer            *dispute.Processor
	messager            messager.Sender
}

func NewCelerMsgDispatcher(
	nodeConfig common.GlobalNodeConfig,
	streamWriter common.StreamWriter,
	crypto common.Crypto,
	channelRouter common.StateChannelRouter,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	cooperativeWithdraw CooperativeWithdraw,
	serverForwarder handlers.ForwardToServerCallback,
	disputer *dispute.Processor,
	messager *messager.Messager,
) *CelerMsgDispatcher {
	d := &CelerMsgDispatcher{
		nodeConfig:          nodeConfig,
		streamWriter:        streamWriter,
		crypto:              crypto,
		channelRouter:       channelRouter,
		monitorService:      monitorService,
		dal:                 dal,
		stop:                false,
		cooperativeWithdraw: cooperativeWithdraw,
		serverForwarder:     serverForwarder,
		tokenCallbackLock:   &sync.RWMutex{},
		sendingCallbackLock: &sync.RWMutex{},
		disputer:            disputer,
		messager:            messager,
	}
	return d
}

var msgProcessingWall = make(map[string]metric.Metric)
var errHandlingCounts = make(map[string]metric.Metric)
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

func init() {
	for _, name := range msgNames {
		msgProcessingWall[name] = metric.NewHistogram("10d1h", "24h10m", "15m10s")
		expvar.Publish(name+"-wall", msgProcessingWall[name])
		errHandlingCounts[name] = metric.NewCounter("10d1h", "24h10m", "15m10s")
		expvar.Publish(name+"-err", errHandlingCounts[name])
	}
	errHandlingCounts["unamed"] = metric.NewCounter("10d1h", "24h10m", "15m10s")
	expvar.Publish("unamed-err", errHandlingCounts["unamed"])
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

		handler = msghdl.NewCelerMsgHandler(
			d.nodeConfig,
			d.streamWriter,
			d.crypto,
			d.channelRouter,
			d.monitorService,
			d.serverForwarder,
			d.onReceivingToken,
			d.tokenCallbackLock,
			d.onSendingToken,
			d.sendingCallbackLock,
			d.disputer,
			d.cooperativeWithdraw,
			d.messager,
			d.dal,
		)

		start := time.Now()
		err := handler.Run(msgFrame)
		msgname := handler.GetMsgName()
		msgMetric := msgProcessingWall[msgname]
		if msgMetric != nil {
			msgMetric.Add(time.Since(start).Seconds())
		}
		metrics.IncDispatcherMsgCnt(msgname)
		metrics.IncDispatcherMsgProcDur(start, msgname)
		if err != nil {
			logEntry.Error = append(logEntry.Error, err.Error())
			errHandlingCount := errHandlingCounts[msgname]
			if errHandlingCount != nil {
				errHandlingCount.Add(1)
			}
			metrics.IncDispatcherErrCnt(msgname)
		}
		pem.CommitPem(logEntry)
	}
}

func (d *CelerMsgDispatcher) Stop() {
	d.stop = true
}
