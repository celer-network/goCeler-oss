// Copyright 2018-2020 Celer Network

package webapi

import (
	"github.com/celer-network/goCeler/celersdk"
	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goutils/log"
)

type sendErr struct {
	pay *celersdkintf.Payment
	e   *celersdkintf.E
}

type callbackImpl struct {
	clientReady      chan *celersdk.Client
	clientInitErr    chan *celersdkintf.E
	channelOpened    chan string
	openChannelError chan string
	recvStart        chan *celersdkintf.Payment
	recvDone         chan *celersdkintf.Payment
	sendComplete     chan *celersdkintf.Payment
	sendErr          chan *sendErr
}

func NewCallbackImpl() *callbackImpl {
	return &callbackImpl{
		clientReady:      make(chan *celersdk.Client),
		clientInitErr:    make(chan *celersdkintf.E),
		channelOpened:    make(chan string),
		openChannelError: make(chan string),
		recvStart:        make(chan *celersdkintf.Payment),
		recvDone:         make(chan *celersdkintf.Payment),
		sendComplete:     make(chan *celersdkintf.Payment),
		sendErr:          make(chan *sendErr),
	}
}

func (c *callbackImpl) HandleClientReady(client *celersdk.Client) {
	c.clientReady <- client
}

func (c *callbackImpl) HandleClientInitErr(e *celersdkintf.E) {
	c.clientInitErr <- e
}

func (c *callbackImpl) HandleChannelOpened(token, cid string) {
	c.channelOpened <- cid
}

func (c *callbackImpl) HandleOpenChannelError(token, reason string) {
	c.openChannelError <- reason
}

// Callback triggered when secret revealed.
func (c *callbackImpl) HandleRecvStart(pay *celersdkintf.Payment) {
	select {
	case c.recvStart <- pay:
	default:
	}
}

// Callback triggered when pay settle request processed.
func (c *callbackImpl) HandleRecvDone(pay *celersdkintf.Payment) {
	log.Infoln("payID", pay.UID, "recv done. Status:", PayStatusName(pay.Status))
	select {
	case c.recvDone <- pay:
	default:
	}
}

func (c *callbackImpl) HandleSendComplete(pay *celersdkintf.Payment) {
	log.Infoln("payID", pay.UID, "send complete. Status:", PayStatusName(pay.Status))
	select {
	case c.sendComplete <- pay:
	default:
	}
}

func (c *callbackImpl) HandleSendErr(pay *celersdkintf.Payment, e *celersdkintf.E) {
	log.Errorln("payID", pay.UID, "send err:", e)
	select {
	case c.sendErr <- &sendErr{pay: pay, e: e}:
	default:
	}
}

func PayStatusName(status int) string {
	switch status {
	case celersdkintf.PAY_STATUS_INVALID:
		return "PAY_STATUS_INVALID"
	case celersdkintf.PAY_STATUS_PENDING:
		return "PAY_STATUS_PENDING"
	case celersdkintf.PAY_STATUS_PAID:
		return "PAY_STATUS_PAID"
	case celersdkintf.PAY_STATUS_PAID_RESOLVED_ONCHAIN:
		return "PAY_STATUS_PAID_RESOLVED_ONCHAIN"
	case celersdkintf.PAY_STATUS_UNPAID:
		return "PAY_STATUS_UNPAID"
	case celersdkintf.PAY_STATUS_UNPAID_EXPIRED:
		return "PAY_STATUS_UNPAID_EXPIRED"
	case celersdkintf.PAY_STATUS_UNPAID_REJECTED:
		return "PAY_STATUS_UNPAID_REJECTED"
	case celersdkintf.PAY_STATUS_UNPAID_DEST_UNREACHABLE:
		return "PAY_STATUS_UNPAID_DEST_UNREACHABLE"
	case celersdkintf.PAY_STATUS_INITIALIZING:
		return "PAY_STATUS_INITIALIZING"
	default:
		return "PAY_STATUS_NULL"
	}
}
