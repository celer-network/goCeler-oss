// Copyright 2019-2020 Celer Network

// this file defines externalsigner interface for app to implement
// and provides common.signer to cnode, also adds new API for create
// sdk client
package celersdk

import (
	"errors"
	"sync"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	ec "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// pkg level singleton so PublishSignedResult can be pkg level func
// we need this because RegisterStream require sign but at this time
// client isn't available to mobile yet. We could move it out and have
// separate api for connect, but the core requirement that during all init
// code, no sign can be requested is too strict
// esm is set/replaced whenever initClientWithSigner is called, same instance
// is also passed into celer client/cnode
var (
	esm     *extSignerManager
	esmLock sync.RWMutex
)

var (
	ErrNilExtSigner  = errors.New("external signer is nil")
	ErrResultTimeout = errors.New("timeout waiting for sign result")
	ErrInvalidReqID  = errors.New("invalid sign reqid")
	ErrNilResult     = errors.New("sign result is nil")
)

// wait at most 60s for external to return sign result
const SignTimeout = 60 * time.Second

// mobile needs to implement this as callback, corresponding cb will be
// called by goceler when need to sign stuff
type ExternalSignerCallback interface {
	OnSignMessage(reqid int, msg []byte)
	// OnSignTransaction rawtx is RLP encoded bytes of Eth transaction
	OnSignTransaction(reqid int, rawtx []byte)
}

// SDK API for mobile to call to send back sign result
// if mobile has error, send nil result so goCeler will know sign failed
func PublishSignedResult(reqid int, result []byte) error {
	log.Debugf("publish signed result %d %x", reqid, result)
	esmLock.RLock()
	defer esmLock.RUnlock()
	if esm == nil {
		return ErrNilExtSigner
	}
	if len(result) == 65 { // this should be sig for msg b/c rlp tx will be much longer
		s := ctype.Bytes2Sig(result)
		if s[64] == 27 || s[64] == 28 {
			// fix v so our RecoverSigner can work
			// note we also made RecoverSigner support both 0/1 and 27/28, but need to
			// consider previous deployed services so fix in both places are safer
			log.Debugf("fix v in sig, %d -> %d", s[64], s[64]-27)
			s[64] -= 27
		}
		return esm.SendSignResult(reqid, s.Bytes())
	}
	// signed tx case, return directly, could make a copy for extra safe
	return esm.SendSignResult(reqid, result)
}

// extSignerManager implements eth.Signer interface and triggers external callback
type extSignerManager struct {
	seq int                    // mono-increase per sign, correlate extCb and result api
	m   map[int](chan []byte)  // key is seq(reqid), value is chan of signed result
	mux sync.RWMutex           // protect seq and map
	cb  ExternalSignerCallback // mobile provided cb

	sigmap     map[ctype.Hash]ctype.Sig // map from msg hash to sig, to avoid sign requests for same data
	sigmapLock sync.RWMutex
}

// ❌❌❌ WARNING: REPLACE pkg level var esm ❌❌❌
func newExtSignerMgr(ecb ExternalSignerCallback) *extSignerManager {
	ret := &extSignerManager{
		seq:    int(time.Now().Unix()),
		m:      make(map[int]chan []byte),
		cb:     ecb,
		sigmap: make(map[ctype.Hash]ctype.Sig),
	}
	esmLock.Lock()
	esm = ret // update pkg level esm
	esmLock.Unlock()
	return ret
}

// return es.sigmap[key].Bytes, or nil if key not found
func (es *extSignerManager) getSig(key ctype.Hash) []byte {
	es.sigmapLock.RLock()
	defer es.sigmapLock.RUnlock()
	ret, ok := es.sigmap[key]
	if !ok {
		return nil
	}
	return ret.Bytes()
}

// set sigmap[key] = ctype.Bytes2Sig(value)
func (es *extSignerManager) putSig(key ctype.Hash, value []byte) {
	es.sigmapLock.Lock()
	defer es.sigmapLock.Unlock()
	es.sigmap[key] = ctype.Bytes2Sig(value)
}

// eth.Signer, blocking till mobile calls result api with matched reqid
func (es *extSignerManager) SignEthMessage(msg []byte) ([]byte, error) {
	if es.cb == nil {
		return nil, ErrNilExtSigner
	}
	msghash := crypto.Keccak256(msg)
	sigmapKey := ec.BytesToHash(msghash)
	cachedSig := es.getSig(sigmapKey)
	if cachedSig != nil {
		log.Debugf("sign eth message cache hit. msg: %x, sig: %x", msg, cachedSig)
		return cachedSig, nil
	}
	id, c := es.newSeqChan()
	log.Debugf("sign eth message %d %x", id, msg)
	if _, ok := es.cb.(eth.Signer); ok {
		// cb also implements eth.Signer, this is for our e2e test case
		// extSigner in api_server.go embed a cobj.CelerSigner which does hash internally so we don't double hash
		go es.cb.OnSignMessage(id, msg) // trigger cb
	} else {
		// external signer, only sign the hash result to avoid they do hash
		// eg. samsung keystore signEthPersonalMessage adds prefix automatically
		log.Debugf("hash for extsigner to sign: %x", msghash)
		go es.cb.OnSignMessage(id, msghash)
	}
	t := time.NewTimer(SignTimeout)
	select {
	case <-t.C:
		log.Debug("sign eth message timeout")
		return nil, ErrResultTimeout
	case ret := <-c: // received result bytes, m[id]chan will be closed by SendSignResult
		t.Stop()
		if ret == nil {
			log.Debug("sign eth message nil result")
			return nil, ErrNilResult
		}
		log.Debugf("sign eth message res %x", ret)
		es.putSig(sigmapKey, ret)
		return ret, nil
	}
}

// eth.Signer, blocking till mobile calls result api with matched reqid
func (es *extSignerManager) SignEthTransaction(rawtx []byte) ([]byte, error) {
	if es.cb == nil {
		return nil, ErrNilExtSigner
	}
	id, c := es.newSeqChan()
	log.Debugf("sign eth tx %d %x", id, rawtx)
	go es.cb.OnSignTransaction(id, rawtx) // trigger cb
	t := time.NewTimer(SignTimeout)
	select {
	case <-t.C:
		log.Debug("sign eth tx timeout")
		return nil, ErrResultTimeout
	case ret := <-c: // received result bytes, m[id]chan will be closed by SendSignResult
		t.Stop()
		if ret == nil {
			log.Debug("sign eth tx nil result")
			return nil, ErrNilResult
		}
		log.Debugf("sign eth tx res %x", ret)
		return ret, nil
	}
}

// called by mobile to send back signed result, write to matched chan to unblock SignXXX
func (es *extSignerManager) SendSignResult(reqid int, result []byte) error {
	es.mux.RLock()
	c, ok := es.m[reqid]
	es.mux.RUnlock()
	if !ok {
		return ErrInvalidReqID
	}
	c <- result // send result to c to unblock SignXXX
	es.mux.Lock()
	defer es.mux.Unlock()
	// cleanup, close c and delete map
	close(c)
	delete(es.m, reqid)
	return nil
}

// helper to +1 seq, and make chan to be used by SignXXX
func (es *extSignerManager) newSeqChan() (int, chan []byte) {
	newch := make(chan []byte, 1) // buffered chan size 1 so write won't be block even no reader
	es.mux.Lock()
	defer es.mux.Unlock()
	es.seq++
	es.m[es.seq] = newch
	return es.seq, newch
}
