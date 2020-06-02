// Copyright 2018-2020 Celer Network
//
// The connection manager keeps track of the open hop and flow streams
// per destination address and their receiving goroutines.  It detects
// stream disconnects and notifies the application so it can reconnect
// by registering new streams for the destination.

package rpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	grpc "google.golang.org/grpc"
)

type CelerStream interface {
	Send(*CelerMsg) error
	Recv() (*CelerMsg, error)
}

type ErrCallbackFunc func(addr ctype.Addr, err error)
type RegisterClientCallbackFunc func(clientAddr ctype.Addr)
type MsgQueueCallbackFunc func(peer ctype.Addr) error

// syncInfo holds data used to synchronize between the goroutines of
// an address and coordinate the handling of teardown and cleanup
type syncInfo struct {
	wg   sync.WaitGroup // wait for goroutines to break their loops
	stop chan bool      // tell the goroutines to terminate
}

type SafeSendCelerStream struct {
	celerStream CelerStream
	sendLock    sync.Mutex
	done        chan bool
}

func (s *SafeSendCelerStream) SafeSend(msg *CelerMsg) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()
	return s.celerStream.Send(msg)
}

type ConnectionManager struct {
	// Eth->CelerStream
	celerStreams map[ctype.Addr]*SafeSendCelerStream
	// Eth->Connection
	conns        map[ctype.Addr]*grpc.ClientConn
	syncInfos    map[ctype.Addr]*syncInfo
	errCallbacks map[ctype.Addr]ErrCallbackFunc
	lock         *sync.RWMutex
	// Callback used in the multi-server setup to register that
	// the client connected to this OSP.
	regClient RegisterClientCallbackFunc
	// Callbacks used to enable/disable message queue processing for
	// a peer when its  connection is broken.
	enableMsgQueue  MsgQueueCallbackFunc
	disableMsgQueue MsgQueueCallbackFunc
}

func NewConnectionManager(regClient RegisterClientCallbackFunc) *ConnectionManager {
	return &ConnectionManager{
		celerStreams: make(map[ctype.Addr]*SafeSendCelerStream),
		conns:        make(map[ctype.Addr]*grpc.ClientConn),
		syncInfos:    make(map[ctype.Addr]*syncInfo),
		errCallbacks: make(map[ctype.Addr]ErrCallbackFunc),
		lock:         &sync.RWMutex{},
		regClient:    regClient,
	}
}

func (m *ConnectionManager) SetMsgQueueCallback(enable, disable MsgQueueCallbackFunc) {
	m.enableMsgQueue = enable
	m.disableMsgQueue = disable
}

func (m *ConnectionManager) AddConnection(peerAddr ctype.Addr, cc *grpc.ClientConn) {
	if cc == nil {
		log.Panicln("error: nil conn given is invalid", peerAddr.Hex())
	}

	//log.Debug("waiting for addConnection lock")
	m.lock.Lock()
	//log.Debug("grab addConnection lock")
	defer m.lock.Unlock()
	m.conns[peerAddr] = cc
}

func (m *ConnectionManager) GetClient(peerAddr ctype.Addr) (RpcClient, error) {
	//log.Debug("waiting for GetClient lock")
	m.lock.RLock()
	//log.Debug("grab GetClient lock")
	defer m.lock.RUnlock()

	cc := m.conns[peerAddr]
	if cc == nil {
		log.Errorln("error: nil rpc connection for", peerAddr.Hex())
		return nil, fmt.Errorf("no RPC connection for %x", peerAddr)
	}
	return NewRpcClient(cc), nil
}

func (m *ConnectionManager) GetClientByAddr(peerAddr ctype.Addr) (RpcClient, error) {
	return m.GetClient(peerAddr)
}

// CloseNoRetry remove errCallbacks so no more retry, then close the underlying grpc conn
// expect to be only called in cNode.Close
func (m *ConnectionManager) CloseNoRetry(onchainAddr ctype.Addr) {
	// we delete callback first because conn.Close will cause .Recv to err and exit
	// loop then it calls CloseConnection again and tries to trigger callback
	// delete first avoid race condition
	m.lock.Lock()
	delete(m.errCallbacks, onchainAddr)
	m.lock.Unlock()
	m.CloseConnection(onchainAddr)
}

func (m *ConnectionManager) CloseConnection(peerAddr ctype.Addr) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if cc, exist := m.conns[peerAddr]; exist {
		cc.Close()
		delete(m.conns, peerAddr)
	}
}

type Rpc_CelerOneServer interface {
	Send(*CelerMsg) error
	Recv() (*CelerMsg, error)
}

func (m *ConnectionManager) AddCelerStream(peerAddr ctype.Addr, stream CelerStream, msgProcessor chan *CelerMsg) context.Context {
	m.lock.Lock()
	defer m.lock.Unlock()
	if oldSafeSend, exist := m.celerStreams[peerAddr]; exist {
		delete(m.celerStreams, peerAddr)
		close(oldSafeSend.done)
	}
	safeSend := &SafeSendCelerStream{
		celerStream: stream,
		done:        make(chan bool),
	}
	m.celerStreams[peerAddr] = safeSend
	if m.regClient != nil {
		m.regClient(peerAddr)
	}
	if m.enableMsgQueue != nil {
		err := m.enableMsgQueue(peerAddr)
		if err != nil {
			log.Warnln("CelerStream: enable peer message queue error:", peerAddr.Hex(), ":", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var err error
		defer close(msgProcessor)
		replaced := false
		for !replaced {
			// Needed because the select-on-msgProcess does not
			// give priority to the safeSend.done channel.
			select {
			case <-safeSend.done:
				replaced = true
				continue
			default:
			}

			var msg *CelerMsg
			msg, err = stream.Recv()
			if err != nil {
				log.Infoln("CelerStream error for", peerAddr.Hex(), ":", err)
				break
			}

			// Needed to break out of blocking on msgProcessor channel.
			select {
			case msgProcessor <- msg:
			case <-safeSend.done:
				replaced = true
			}
		}

		// If terminated by the closing of the "done" channel, the old
		// stream is being replaced by a new stream. The new stream will
		// overwrite the old one, so don't do any cleanup to avoid race
		// conditions between the new setup and the old cleanup.
		if replaced {
			cancel()
			return
		}

		if m.disableMsgQueue != nil {
			err = m.disableMsgQueue(peerAddr)
			if err != nil {
				log.Warnf("CelerStream: disable peer %x message queue error: %s", peerAddr, err)
			}
		}
		m.CloseConnection(peerAddr)
		m.lock.Lock()
		delete(m.celerStreams, peerAddr)
		cb, hasCb := m.errCallbacks[peerAddr]
		m.lock.Unlock()

		cancel()
		if hasCb {
			log.Debugln("CelerStream app callback", peerAddr.Hex())
			cb(peerAddr, err)
		}
	}()
	return ctx
}

func (m *ConnectionManager) AddErrorCallback(peerAddr ctype.Addr, cb ErrCallbackFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.errCallbacks[peerAddr] = cb
}

// Helper to return the syncInfo entry for an address, and create one if
// it does yet exist.  Note: must be called with the "lock" held.
func (m *ConnectionManager) getSyncInfo(peerAddr ctype.Addr) *syncInfo {
	si, ok := m.syncInfos[peerAddr]
	if !ok {
		log.Debugln("getSyncInfo: created", peerAddr.Hex())
		si = &syncInfo{
			stop: make(chan bool),
		}
		si.wg.Add(2) // will later start 2 goroutines (hop & flow)
		m.syncInfos[peerAddr] = si
	}
	return si
}

// CleanupStreams terminate and cleanup the existing (old) hop & flow stream goroutines to
// allow for new streams & goroutines to be registered for the same address.
// It helps call from external to aquire lock.
func (m *ConnectionManager) CleanupStreams(peerAddr ctype.Addr) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cleanupStreams(peerAddr)
}

// Terminate and cleanup the existing (old) hop & flow stream goroutines to
// allow for new streams & goroutines to be registered for the same address.
// Note: this function must be called with the "lock" held.
func (m *ConnectionManager) cleanupStreams(peerAddr ctype.Addr) {
	si := m.syncInfos[peerAddr]
	if si == nil {
		return // no goroutines are setup
	}

	// A new stream registration on top of an old one happens when a client
	// reconnects after a disconnect, but before the server has realized that
	// the disconnect happened.  This means the 2 goroutines (hop & flow)
	// have not yet received an error and are still running.  Tell them to
	// stop at the next opportunity (effectively they become useless).
	log.Debugln("cleanupStreams for", peerAddr.Hex())
	close(si.stop)

	delete(m.celerStreams, peerAddr)
	delete(m.syncInfos, peerAddr)
}

func mustStop(si *syncInfo) bool {
	select {
	case <-si.stop: // is this channel closed?
		return true
	default:
		return false
	}
}

func (m *ConnectionManager) GetCelerStream(peerAddr ctype.Addr) *SafeSendCelerStream {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.celerStreams[peerAddr]
}

func (m *ConnectionManager) GetNumCelerStreams() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.celerStreams)
}
