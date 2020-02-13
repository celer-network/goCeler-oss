// Copyright 2018-2019 Celer Network

package messager

import (
	"bytes"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/handlers"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/golang/protobuf/ptypes/any"
)

type Sender interface {
	SendCondPayRequest(pay *entity.ConditionalPay, note *any.Any, logEntry *pem.PayEventMessage) error

	ForwardCondPayRequestMsg(*common.MsgFrame) error

	SendOnePaySettleRequest(
		pay *entity.ConditionalPay,
		payAmt *big.Int,
		reason rpc.PaymentSettleReason, logEntry *pem.PayEventMessage) error

	SendPaysSettleRequest(
		pays []*entity.ConditionalPay,
		payAmts []*big.Int, // total amount for all setted payments
		reason rpc.PaymentSettleReason, logEntry *pem.PayEventMessage) ([]*entity.ConditionalPay, error)

	ForwardPaySettleRequestMsg(frame *common.MsgFrame) error

	SendOnePaySettleProof(
		payID ctype.PayIDType,
		reason rpc.PaymentSettleReason,
		logEntry *pem.PayEventMessage) error

	SendPaysSettleProof(
		payIDs []ctype.PayIDType,
		reason rpc.PaymentSettleReason,
		logEntry *pem.PayEventMessage) error

	ForwardPaySettleProofMsg(frame *common.MsgFrame) error

	ForwardCelerMsg(peerTo string, msg *rpc.CelerMsg) error

	AckMsgQueue(cid ctype.CidType, ack uint64, nack uint64) error

	ResendMsgQueue(cid ctype.CidType, seqnum uint64) error

	GetMsgQueue(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, bool)

	IsDirectPay(pay *entity.ConditionalPay, peer string) bool
}

// Messager sends hop and flow messages
type Messager struct {
	nodeConfig      common.GlobalNodeConfig
	signer          common.Signer
	streamWriter    common.StreamWriter
	channelRouter   common.StateChannelRouter
	monitorService  intfs.MonitorService
	serverForwarder handlers.ForwardToServerCallback
	dal             *storage.DAL
	msgQueue        *MsgQueue
}

func NewMessager(
	nodeConfig common.GlobalNodeConfig,
	signer common.Signer,
	streamWriter common.StreamWriter,
	channelRouter common.StateChannelRouter,
	monitorService intfs.MonitorService,
	serverForwarder handlers.ForwardToServerCallback,
	dal *storage.DAL,
) *Messager {
	return &Messager{
		nodeConfig:      nodeConfig,
		signer:          signer,
		streamWriter:    streamWriter,
		channelRouter:   channelRouter,
		monitorService:  monitorService,
		serverForwarder: serverForwarder,
		dal:             dal,
		msgQueue:        NewMsqQueue(dal, streamWriter, nodeConfig.GetOnChainAddr()),
	}
}

func (m *Messager) ForwardCelerMsg(peerTo string, msg *rpc.CelerMsg) error {
	if isLocalPeer, err := m.serverForwarder(peerTo, msg); err != nil {
		return err
	} else if isLocalPeer {
		return m.streamWriter.WriteCelerMsg(peerTo, msg)
	}
	return nil
}

// Enable message queue processing for this peer address.
func (m *Messager) EnableMsgQueue(peer string) error {
	return m.msgQueue.AddPeer(peer)
}

// Disable message queue processing for this peer address.
func (m *Messager) DisableMsgQueue(peer string) error {
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
func (m *Messager) IsDirectPay(pay *entity.ConditionalPay, peer string) bool {
	if len(pay.GetConditions()) > 0 {
		return false
	}

	myAddr := m.nodeConfig.GetOnChainAddrBytes()
	if bytes.Compare(pay.GetSrc(), myAddr) != 0 {
		return false
	}

	dest := ctype.Bytes2Hex(pay.GetDest())
	if peer == "" {
		var err error
		token := pay.GetTransferFunc().GetMaxTransfer().GetToken()
		tokenAddr := utils.GetTokenAddrStr(token)
		_, peer, err = m.channelRouter.LookupNextChannelOnToken(dest, tokenAddr)
		if err != nil {
			log.Warnln("IsDirectPay: cannot determine next-hop peer:", dest, tokenAddr, err)
			return false
		}
	}

	return dest == peer
}
