// Copyright 2018 Celer Network
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

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	ethcommon "github.com/ethereum/go-ethereum/common"
	grpc "google.golang.org/grpc"
)

type CelerStream interface {
	Send(*CelerMsg) error
	Recv() (*CelerMsg, error)
}

type ErrCallbackFunc func(addr string, err error)
type RegisterClientCallbackFunc func(clientAddr string)
type MsgQueueCallbackFunc func(peer string) error

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
	celerStreams map[string]*SafeSendCelerStream
	// Eth->Connection
	conns        map[string]*grpc.ClientConn
	syncInfos    map[string]*syncInfo
	errCallbacks map[string]ErrCallbackFunc
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
		celerStreams: make(map[string]*SafeSendCelerStream),
		conns:        make(map[string]*grpc.ClientConn),
		syncInfos:    make(map[string]*syncInfo),
		errCallbacks: make(map[string]ErrCallbackFunc),
		lock:         &sync.RWMutex{},
		regClient:    regClient,
	}
}

func (m *ConnectionManager) SetMsgQueueCallback(enable, disable MsgQueueCallbackFunc) {
	m.enableMsgQueue = enable
	m.disableMsgQueue = disable
}

func (m *ConnectionManager) AddConnection(onchainAddr string, cc *grpc.ClientConn) {
	if cc == nil {
		log.Panicln("error: nil conn given is invalid", onchainAddr)
	}

	//log.Debug("waiting for addConnection lock")
	m.lock.Lock()
	//log.Debug("grab addConnection lock")
	defer m.lock.Unlock()
	m.conns[onchainAddr] = cc
}

func (m *ConnectionManager) GetClient(onchainAddr string) (RpcClient, error) {
	//log.Debug("waiting for GetClient lock")
	m.lock.RLock()
	//log.Debug("grab GetClient lock")
	defer m.lock.RUnlock()

	cc := m.conns[onchainAddr]
	if cc == nil {
		log.Errorln("error: nil rpc connection for", onchainAddr)
		return nil, fmt.Errorf("no RPC connection for %s", onchainAddr)
	}
	return NewRpcClient(cc), nil
}

func (m *ConnectionManager) GetClientByAddr(peerAddr ethcommon.Address) (RpcClient, error) {
	return m.GetClient(ctype.Addr2Hex(peerAddr))
}

func (m *ConnectionManager) CloseConnection(onchainAddr string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if cc, exist := m.conns[onchainAddr]; exist {
		cc.Close()
		delete(m.conns, onchainAddr)
	}
}

type Rpc_CelerOneServer interface {
	Send(*CelerMsg) error
	Recv() (*CelerMsg, error)
}

func (m *ConnectionManager) AddCelerStream(onchainAddr string, stream CelerStream, msgProcessor chan *CelerMsg) context.Context {
	m.lock.Lock()
	defer m.lock.Unlock()
	if oldSafeSend, exist := m.celerStreams[onchainAddr]; exist {
		delete(m.celerStreams, onchainAddr)
		close(oldSafeSend.done)
	}
	safeSend := &SafeSendCelerStream{
		celerStream: stream,
		done:        make(chan bool),
	}
	m.celerStreams[onchainAddr] = safeSend
	if m.regClient != nil {
		m.regClient(onchainAddr)
	}
	if m.enableMsgQueue != nil {
		err := m.enableMsgQueue(onchainAddr)
		if err != nil {
			log.Warnln("CelerStream: enable peer message queue error:", onchainAddr, ":", err)
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
				log.Infoln("CelerStream error for", onchainAddr, ":", err)
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
			err = m.disableMsgQueue(onchainAddr)
			if err != nil {
				log.Warnln("CelerStream: disable peer message queue error:", onchainAddr, ":", err)
			}
		}
		m.CloseConnection(onchainAddr)
		m.lock.Lock()
		delete(m.celerStreams, onchainAddr)
		cb, hasCb := m.errCallbacks[onchainAddr]
		m.lock.Unlock()

		cancel()
		if hasCb {
			log.Debugln("CelerStream app callback", onchainAddr)
			cb(onchainAddr, err)
		}
	}()
	return ctx
}

func (m *ConnectionManager) AddErrorCallback(onchainAddr string, cb ErrCallbackFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.errCallbacks[onchainAddr] = cb
}

// Helper to return the syncInfo entry for an address, and create one if
// it does yet exist.  Note: must be called with the "lock" held.
func (m *ConnectionManager) getSyncInfo(onchainAddr string) *syncInfo {
	si, ok := m.syncInfos[onchainAddr]
	if !ok {
		log.Debugln("getSyncInfo: created", onchainAddr)
		si = &syncInfo{
			stop: make(chan bool),
		}
		si.wg.Add(2) // will later start 2 goroutines (hop & flow)
		m.syncInfos[onchainAddr] = si
	}
	return si
}

// CleanupStreams terminate and cleanup the existing (old) hop & flow stream goroutines to
// allow for new streams & goroutines to be registered for the same address.
// It helps call from external to aquire lock.
func (m *ConnectionManager) CleanupStreams(addr string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cleanupStreams(addr)
}

// Terminate and cleanup the existing (old) hop & flow stream goroutines to
// allow for new streams & goroutines to be registered for the same address.
// Note: this function must be called with the "lock" held.
func (m *ConnectionManager) cleanupStreams(addr string) {
	si := m.syncInfos[addr]
	if si == nil {
		return // no goroutines are setup
	}

	// A new stream registration on top of an old one happens when a client
	// reconnects after a disconnect, but before the server has realized that
	// the disconnect happened.  This means the 2 goroutines (hop & flow)
	// have not yet received an error and are still running.  Tell them to
	// stop at the next opportunity (effectively they become useless).
	log.Debugln("cleanupStreams for", addr)
	close(si.stop)

	delete(m.celerStreams, addr)
	delete(m.syncInfos, addr)
}

func mustStop(si *syncInfo) bool {
	select {
	case <-si.stop: // is this channel closed?
		return true
	default:
		return false
	}
}

func (m *ConnectionManager) GetCelerStream(onchainAddr string) *SafeSendCelerStream {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.celerStreams[onchainAddr]
}

func (m *ConnectionManager) GetNumCelerStreams() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.celerStreams)
}
