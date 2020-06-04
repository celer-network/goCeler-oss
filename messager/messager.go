// Copyright 2018-2020 Celer Network

package messager

import (
	"bytes"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/handlers"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
)

// Messager sends hop and flow messages
type Messager struct {
	nodeConfig       common.GlobalNodeConfig
	signer           eth.Signer
	streamWriter     common.StreamWriter
	routeForwarder   *route.Forwarder
	monitorService   intfs.MonitorService
	serverForwarder  handlers.ForwardToServerCallback
	depositProcessor *deposit.Processor
	dal              *storage.DAL
	msgQueue         *MsgQueue
	isOSP            bool
}

func NewMessager(
	nodeConfig common.GlobalNodeConfig,
	signer eth.Signer,
	streamWriter common.StreamWriter,
	routeForwarder *route.Forwarder,
	monitorService intfs.MonitorService,
	serverForwarder handlers.ForwardToServerCallback,
	depositProcessor *deposit.Processor,
	dal *storage.DAL,
	isOSP bool,
) *Messager {
	return &Messager{
		nodeConfig:       nodeConfig,
		signer:           signer,
		streamWriter:     streamWriter,
		routeForwarder:   routeForwarder,
		monitorService:   monitorService,
		serverForwarder:  serverForwarder,
		depositProcessor: depositProcessor,
		dal:              dal,
		isOSP:            isOSP,
		msgQueue:         NewMsqQueue(dal, streamWriter, nodeConfig.GetOnChainAddr()),
	}
}

func (m *Messager) ForwardCelerMsg(peerTo ctype.Addr, msg *rpc.CelerMsg) error {
	if isLocalPeer, err := m.serverForwarder(peerTo, true, msg); err != nil {
		return err
	} else if isLocalPeer {
		return m.streamWriter.WriteCelerMsg(peerTo, msg)
	}
	return nil
}

// Enable message queue processing for this peer address.
func (m *Messager) EnableMsgQueue(peer ctype.Addr) error {
	return m.msgQueue.AddPeer(peer)
}

// Disable message queue processing for this peer address.
func (m *Messager) DisableMsgQueue(peer ctype.Addr) error {
	return m.msgQueue.RemovePeer(peer)
}

// ACK a message in a channel queue.
func (m *Messager) AckMsgQueue(cid ctype.CidType, ack, nack uint64) error {
	return m.msgQueue.AckMsg(cid, ack, nack)
}

// Resend a message in a channel queue.
func (m *Messager) ResendMsgQueue(cid ctype.CidType, seqnum uint64) error {
	return m.msgQueue.ResendMsg(cid, seqnum)
}

// Get a message from a channel queue.
func (m *Messager) GetMsgQueue(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, bool) {
	return m.msgQueue.GetMsg(cid, seqnum)
}

// Is this a direct payment from me to this peer?  The peer is an optional
// parameter, if it is not given (an empty string), the next hop peer is
// looked up.  For now only consider unconditional payments where I am the
// source and the destination is my next hop peer.  This is typical of fee
// (client to OSP) and prize (OSP to client) payments in centralized games.
func (m *Messager) IsDirectPay(pay *entity.ConditionalPay, peer ctype.Addr) bool {
	if len(pay.GetConditions()) > 0 {
		return false
	}

	if bytes.Compare(pay.GetSrc(), m.nodeConfig.GetOnChainAddr().Bytes()) != 0 {
		return false
	}

	dest := ctype.Bytes2Addr(pay.GetDest())
	if peer == ctype.ZeroAddr {
		var err error
		tokenAddr := utils.GetTokenAddr(pay.GetTransferFunc().GetMaxTransfer().GetToken())
		_, peer, err = m.routeForwarder.LookupNextChannelOnToken(dest, tokenAddr)
		if err != nil {
			log.Warnln("IsDirectPay: cannot determine next-hop peer:", dest.Hex(), tokenAddr.Hex(), err)
			return false
		}
	}

	return dest == peer
}
