// Copyright 2019-2020 Celer Network
//
// cNode helper code for multi-server support.

package cnode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/celer-network/goCeler/ctype"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/lrucache"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goutils/log"
	"google.golang.org/grpc"
)

const (
	serverForwarderMaxRetry    = 3
	serverForwarderRetryDelay  = 3 * time.Second
	serverForwarderDialTimeout = 2 * time.Second
)

// Keep track of open gRPC client connections to other servers.
// It is cached and used for inter-server communication.
type serverClient struct {
	addr   string
	client rpc.MultiServerClient
	conn   *grpc.ClientConn
}

// Create a new gRPC client connection from this server to another server.
func newServerClient(rpcAddr string) (*serverClient, error) {
	conn, err := grpc.Dial(rpcAddr, grpc.WithInsecure(),
		grpc.WithBlock(), grpc.WithTimeout(serverForwarderDialTimeout))
	if err != nil {
		return nil, err
	}

	c := &serverClient{
		addr:   rpcAddr,
		client: rpc.NewMultiServerClient(conn),
		conn:   conn,
	}
	return c, nil
}

// Callback from the LRU cache of OSP client connections to close the
// connection when it is evicted from the cache.
func dropServerClient(rpcAddr string, client interface{}) {
	log.Debugf("dropServerClient: %s conn evicted from LRU cache", rpcAddr)
	oc := client.(*serverClient)
	oc.Close()
}

// Close this OSP client connection.
func (oc *serverClient) Close() {
	if oc.conn != nil {
		oc.conn.Close()
		oc.conn = nil
		oc.client = nil
		oc.addr = ""
	}
}

// Fetch an OSP client connection to the given OSP server.
// Reuse the one in the cache, or create a new one and cache it.
func (c *CNode) getServerClient(rpcAddr string) (*serverClient, error) {
	c.serverCacheLock.Lock()
	defer c.serverCacheLock.Unlock()

	v, found := c.serverCache.Get(rpcAddr)
	if found {
		return v.(*serverClient), nil
	}

	oc, err := newServerClient(rpcAddr)
	if err != nil {
		return nil, err
	}
	c.serverCache.Put(rpcAddr, oc)
	return oc, nil
}

// Find the OSP server that owns the given client.  Reuse the info
// in the cache, or fetch it from the storage server and cache it.
// If "refresh" is set, reload the data from storage.
func (c *CNode) getClientConnServer(client ctype.Addr, refresh bool) (string, error) {
	if !refresh {
		v, found := c.clientCache.Get(ctype.Addr2Hex(client))
		if found {
			return v.(string), nil
		}
	}

	svr, found, err := c.dal.GetPeerServer(client)
	if err != nil {
		return "", err
	} else if !found || svr == "" {
		return "", common.ErrPeerNotFound
	}
	c.clientCache.Put(ctype.Addr2Hex(client), svr)
	return svr, nil
}

// This function is used in the multi-server setup as a callback from the
// connection manager to register a newly connected client at this server.
// This lets other servers know to reach this client and this server.
func (c *CNode) registerClientForServer(client ctype.Addr) {
	myAddr := c.GetRPCAddr()
	err := c.dal.UpsertPeerServer(client, myAddr)
	if err != nil {
		log.Errorf("Cannot register client %x at server %s. error: %s", client, myAddr, err)
	}
}

// Is the destination client locally connected on this server?
func (c *CNode) IsLocalPeer(client ctype.Addr) bool {
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
func (c *CNode) defServerForwarder(dest ctype.Addr, retry bool, msg interface{}) (bool, error) {
	if c.IsLocalPeer(dest) {
		return true, nil
	}
	log.Errorln("failed to forward dest", ctype.Addr2Hex(dest), common.ErrPeerNotOnline)
	return false, common.ErrPeerNotOnline
}

// For the multi-server setup, find which server has the connection to the
// destination client and forward the message to it.
func (c *CNode) multiServerForwarder(dest ctype.Addr, retry bool, msg interface{}) (bool, error) {
	req := rpc.FwdReq{
		Dest: ctype.Addr2Hex(dest),
	}

	switch msg.(type) {

	case *rpc.CelerMsg:
		req.Message = msg.(*rpc.CelerMsg)

	default:
		err := errors.New("multiServerForwarder: invalid message type")
		log.Error(err)
		return false, err
	}

	// The destination client may be reconnecting to another server
	// during this operation, so keep trying unless successful.
	myAddr := c.GetRPCAddr()
	refresh := false
	retryNum := 1
	if retry {
		retryNum = serverForwarderMaxRetry
	}
	for i := 0; i < retryNum; i++ {
		if c.IsLocalPeer(dest) {
			log.Debugf("multiServerForwarder: dest %x is local", dest)
			return true, nil
		}

		server, err := c.getClientConnServer(dest, refresh)
		if err != nil {
			return false, err
		} else if server == myAddr {
			log.Infof("multiServerForwarder: dest %x is local after all (%s), try again to get its stream", dest, server)
		} else {
			// Try to forward the message to the target server.
			oc, err := c.getServerClient(server)
			if err != nil {
				log.Error(err)
				return false, err
			}

			reply, err := oc.client.FwdMsg(context.Background(), &req)
			if err != nil {
				log.Error(err)
				return false, err
			}
			accepted := reply.GetAccepted()
			if accepted {
				log.Debugf("multiServerForwarder: sent to remote dest %x", dest)
				return false, nil
			}
			if reply.GetErr() != "" {
				err = fmt.Errorf("failed to forward dest %x msg, err: %s", dest, reply.GetErr())
				log.Errorln(err)
				return false, err
			}

			log.Infof("multiServerForwarder: failed to forward dest %x msg at #%d try", dest, i+1)
		}

		if i < retryNum-1 {
			time.Sleep(serverForwarderRetryDelay)
		}
		refresh = true
	}

	log.Errorln("failed to forward dest", ctype.Addr2Hex(dest), common.ErrPeerNotOnline)
	return false, common.ErrPeerNotOnline
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
		err = c.streamWriter.WriteCelerMsg(ctype.Hex2Addr(req.GetDest()), msg)

	case *rpc.CelerMsg_RevealSecret:
		logEntry.Type = pem.PayMessageType_REVEAL_SECRET
		logEntry.PayId = ctype.Bytes2Hex(msg.GetRevealSecret().GetPayId())
		err = c.streamWriter.WriteCelerMsg(ctype.Hex2Addr(req.GetDest()), msg)

	case *rpc.CelerMsg_RevealSecretAck:
		logEntry.Type = pem.PayMessageType_REVEAL_SECRET_ACK
		logEntry.PayId = ctype.Bytes2Hex(msg.GetRevealSecretAck().GetPayId())
		err = c.streamWriter.WriteCelerMsg(ctype.Hex2Addr(req.GetDest()), msg)

	case *rpc.CelerMsg_PaymentSettleRequest:
		logEntry.Type = pem.PayMessageType_PAY_SETTLE_REQUEST
		err = c.messager.ForwardPaySettleRequestMsg(frame)

	default:
		err = common.ErrInvalidMsgType
	}
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}

	// filter no celer stream error
	if err == nil || strings.Contains(err.Error(), common.ErrNoCelerStream.Error()) {
		metrics.IncSvrFwdMsgCnt(int32(logEntry.Type), metrics.SvrFwdMsgSucceed)
	}

	metrics.IncSvrFwdMsgCnt(int32(logEntry.Type), metrics.SvrFwdMsgAttempt)
	pem.CommitPem(logEntry)
	return err
}

// Initialize the multi-server structures of this cNode object.
// The multi-server setup is identified by having both SelfRPC
// (know myself) and StoreSql (remote database) configured.
func (c *CNode) initMultiServer(profile *common.CProfile) {
	if profile.SelfRPC == "" || profile.StoreSql == "" {
		c.isMultiServer = false
		return
	}

	c.isMultiServer = true
	c.clientCache = lrucache.NewLRUCache(config.ClientCacheSize, nil)
	c.serverCache = lrucache.NewLRUCache(config.ServerCacheSize, dropServerClient)
}

// Send a routing information request to a list of peer OSPs.
func (c *CNode) bcastSend(req *rpc.RoutingRequest, osps []string) {
	if !c.isMultiServer {
		c.BcastRoutingInfo(req, osps)
		return
	}

	// In a multi-server setup, fetch the list of servers, and send
	// the same message and list of peer OPSs to all of them so they
	// each send out the message to the OPSs connected to them.
	servers, err := c.dal.GetAllPeerServers()
	if err != nil {
		log.Errorln("bcastSend: cannot get servers:", err)
		return
	}

	bcastReq := &rpc.BcastRoutingRequest{
		Req:  req,
		Osps: osps,
	}

	for _, svr := range servers {
		c.sendBcastToServer(bcastReq, svr)
	}
}

// The listener/router server uses gRPC to ask an OSP server to
// broadcast the routing information to its peer OSPs.
func (c *CNode) sendBcastToServer(bcastReq *rpc.BcastRoutingRequest, server string) {
	log.Debugln("sendBcastToServer: sending bcast request to:", server)

	oc, err := c.getServerClient(server)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = oc.client.BcastRoutingInfo(context.Background(), bcastReq)
	if err != nil {
		log.Error(err)
		return
	}
}

// An OSP server sends out the message to its peer OSPs.
func (c *CNode) BcastRoutingInfo(req *rpc.RoutingRequest, osps []string) {
	msg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_RoutingRequest{
			RoutingRequest: req,
		},
	}

	for _, osp := range osps {
		dest := ctype.Hex2Addr(osp)
		if c.IsLocalPeer(dest) {
			log.Debugln("send routing request to", osp)
			if err := c.streamWriter.WriteCelerMsg(dest, msg); err != nil {
				log.Errorln("BcastRoutingInfo:", osp, err)
			}
		}
	}
}
