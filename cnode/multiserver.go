// Copyright 2019 Celer Network
//
// cNode helper code for multi-server support.

package cnode

import (
	"fmt"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
)

// Is the destination client locally connected on this server?
func (c *CNode) IsLocalPeer(client string) bool {
	return c.connManager.GetCelerStream(client) != nil
}

// Return the number of clients locally connected on this server?
func (c *CNode) NumClients() int {
	return c.connManager.GetNumCelerStreams()
}

// Return the server's internal RPC address.
func (c *CNode) GetRPCAddr() string {
	return c.nodeConfig.GetRPCAddr()
}

// For the default singleton-server setup (with local storage), all clients
// are connected locally (true value), no forwarding is done.
func (c *CNode) defServerForwarder(dest string, msg interface{}) (bool, error) {
	if c.IsLocalPeer(dest) {
		return true, nil
	}
	err := fmt.Errorf("failed to forward dest %s msg: no peer connection", dest)
	log.Errorln(err)
	return false, err
}

func (c *CNode) ForwardMsgToPeer(req *rpc.FwdReq) error {
	msg := req.GetMessage()
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.ForwardingExternal = true
	frame := &common.MsgFrame{
		Message:  msg,
		LogEntry: logEntry,
	}
	var err error
	switch msg.GetMessage().(type) {
	case *rpc.CelerMsg_CondPayRequest:
		logEntry.Type = pem.PayMessageType_COND_PAY_REQUEST
		err = c.messager.ForwardCondPayRequestMsg(frame)

	case *rpc.CelerMsg_PaymentSettleProof:
		logEntry.Type = pem.PayMessageType_PAY_SETTLE_PROOF
		err = c.messager.ForwardPaySettleProofMsg(frame)

	case *rpc.CelerMsg_CondPayReceipt:
		logEntry.Type = pem.PayMessageType_COND_PAY_RECEIPT
		logEntry.PayId = ctype.Bytes2Hex(msg.GetCondPayReceipt().GetPayId())
		err = c.streamWriter.WriteCelerMsg(req.GetDest(), msg)
	case *rpc.CelerMsg_RevealSecret:
		logEntry.Type = pem.PayMessageType_REVEAL_SECRET
		logEntry.PayId = ctype.Bytes2Hex(msg.GetRevealSecret().GetPayId())
		err = c.streamWriter.WriteCelerMsg(req.GetDest(), msg)
	case *rpc.CelerMsg_RevealSecretAck:
		logEntry.Type = pem.PayMessageType_REVEAL_SECRET_ACK
		logEntry.PayId = ctype.Bytes2Hex(msg.GetRevealSecretAck().GetPayId())
		err = c.streamWriter.WriteCelerMsg(req.GetDest(), msg)

	case *rpc.CelerMsg_PaymentSettleRequest:
		logEntry.Type = pem.PayMessageType_PAY_SETTLE_REQUEST
		err = c.messager.ForwardPaySettleRequestMsg(frame)

	default:
		log.Error(common.ErrInvalidMsgType)
		err = common.ErrInvalidMsgType
	}
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

// Initialize the multi-server structures of this cNode object.
func (c *CNode) initMultiServer(profile *common.CProfile) {
	c.isMultiServer = false
}
