// Copyright 2020 Celer Network

package cnode

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

// WARNING: current auth implmentation will always sync the multi-login devices to the correct
// cosigned state and balance, but may lead to incorrect pay state in local db in corner cases.
// Example: device 1 and 2 share the same user account, latest cosigned seq from user is s.
// step 1: device 1 tried to send payment p1 with seq s+1, but p1 was not delivered to OSP.
// step 2: device 2 auth sync to seq s, then sent p2 with seq s+1, acked by OSP.
// step 3: device 1 auth sync to seq s+1. It would think p1 succeeded, but actually p1 was
//         never delivered. It was p2 from device 2 which reused the same seq (s+1) succeed.

// clientAuthTsMap has address to last authreq ts from client
// new authReq must have a ts that's larger than last to avoid replay attack
// also due to ts is in seconds, we automatically restrict multiple connections within one second
// also user who adjust their clock forward then backward won't be able to connect
var (
	clientAuthTsMap = make(map[ctype.Addr]uint64)
	tsMapLock       sync.RWMutex
)

func (c *CNode) getAuthReq(peerAddr ctype.Addr) (*rpc.AuthReq, error) {
	ts, tsSig := utils.GetTsAndSig(func(b []byte) []byte {
		sig, err := c.signer.SignEthMessage(b)
		if err != nil {
			log.Error(err)
			return nil
		}
		return sig
	})
	if tsSig == nil {
		return nil, fmt.Errorf("auth ts sig error")
	}
	authReq := &rpc.AuthReq{
		MyAddr:          c.EthAddress.Bytes(),
		Timestamp:       ts,
		MySig:           tsSig,
		ExpectPeer:      peerAddr.Bytes(),
		ProtocolVersion: config.AuthProtocolVersion, // >= 1 means support sync in auth
	}
	channels, err := c.dal.GetChannelsForAuthReq(peerAddr)
	if err == nil && len(channels) > 0 {
		authReq.OpenedChannels = channels
	}
	return authReq, nil
}

// HandleAuthReq verifies AuthReq and return msg to send back or error
// if error is nil, guarantee *rpc.CelerMsg isn't nil
func (c *CNode) HandleAuthReq(req *rpc.AuthReq) (*rpc.CelerMsg, error) {
	srcAddr := ctype.Bytes2Addr(req.GetMyAddr())
	if !isTimestampValid(srcAddr, req.GetTimestamp()) {
		return nil, fmt.Errorf("invalid timestamp, src addr %s", ctype.Addr2Hex(srcAddr))
	}
	tsByte := utils.Uint64ToBytes(req.Timestamp)
	if !eth.IsSignatureValid(srcAddr, tsByte, req.GetMySig()) {
		return nil, fmt.Errorf("invalid signature, src addr %s", ctype.Addr2Hex(srcAddr))
	}
	// only update ts after verify sig
	putTs(srcAddr, req.GetTimestamp())
	ret := new(rpc.CelerMsg)
	// if protocol version >=1 , means support sync in auth
	if req.GetProtocolVersion() >= 1 {
		syncChans := c.getSyncChannelsForAuthAck(srcAddr, req.GetOpenedChannels())
		ret.Message = &rpc.CelerMsg_AuthAck{
			AuthAck: &rpc.AuthAck{
				SyncChannels: syncChans,
			},
		}
	}
	return ret, nil
}

// HandleAuthAck tries to update db based on sync info in AuthAck msg
// NOTE this func doesn't return error as it's best effort to sync state
// some code are similar to open_channel logic but couldn't re-use due to openchan assumes init state
func (c *CNode) HandleAuthAck(peer ctype.Addr, ack *rpc.AuthAck) {
	if ack == nil { // do nothing
		return
	}
	var err error
	for _, ch := range ack.SyncChannels {
		chjson, _ := utils.PbToJSONHexBytes(ch)
		log.Infoln("HandleAuthAck, syncing channel:", chjson)
		chState := int(ch.ChannelState)
		cid := ctype.Bytes2Cid(ch.Cid)
		if chState == structs.ChanState_NULL || chState == structs.ChanState_CLOSED {
			log.Errorln("invalid channel state in msg:", ch)
			continue
		}
		mySimplex := ch.AuthreqSimplex
		peerSimplex := ch.AuthackSimplex
		// if simplex isn't nil, means osp sees need to sync
		if peerSimplex != nil {
			err = c.checkSignedSimplex(peerSimplex, cid, peer, c.EthAddress)
			if err != nil {
				log.Error(err)
				continue
			}
		}
		if mySimplex != nil {
			err = c.checkSignedSimplex(mySimplex, cid, c.EthAddress, peer)
			if err != nil {
				log.Error(err)
				continue
			}
		}

		shouldUsePeerLedger := false
		peerLedgerAddr := ctype.Bytes2Addr(ch.GetLedgerAddr())
		peerLedgerContract := c.nodeConfig.GetLedgerContractOn(peerLedgerAddr)
		if peerLedgerContract == nil {
			log.Errorf("Sync channel %x error, doesn't have ledger %x in profile.", ch.GetCid(), ch.GetLedgerAddr())
			continue
		}
		peepMyLedgerAddr, peepFound, peepErr := c.dal.GetChanLedger(cid)
		if peepErr == nil && peepFound && peepMyLedgerAddr != peerLedgerAddr {
			// Only check operability if my ledger is different to peer ledger and
			// we'll only update local ledger addr if peer ledger addr is operable.
			chanStatus, stateErr := ledgerview.GetOnChainChannelStatusOnLedger(cid, c.nodeConfig, peerLedgerAddr)
			if stateErr != nil {
				log.Errorln(stateErr)
				continue
			}
			if chanStatus != ledgerview.OnChainStatus_OPERABLE {
				continue
			}
			shouldUsePeerLedger = true
		}

		myChState, found, err := c.dal.GetChanState(cid)
		if err != nil {
			log.Error(err)
			continue
		}
		var txF storage.TxFunc
		if !found { // no local view, just save everything
			if mySimplex == nil || peerSimplex == nil {
				// not found case require both simplex to be non-nill
				log.Errorf("unexpected nil simplex my: %v, peer: %v", mySimplex, peerSimplex)
				continue
			}
			seqN, tkInfo, err2 := getSeqAndToken(mySimplex)
			if err2 != nil {
				log.Error(err2)
				continue
			}
			bal, err2 := c.getOnChainBalanceForAuth(cid, chState, ch.OpenChannelResponse.GetChannelInitializer(), peerLedgerAddr)
			if err2 != nil {
				log.Error(err2)
				continue
			}

			txF = func(tx *storage.DALTx, args ...interface{}) error {
				err3 := tx.InsertChan(cid, peer, tkInfo, peerLedgerAddr, chState, ch.OpenChannelResponse, bal, seqN, seqN, seqN, 0 /*lastNackedSeqNum*/, mySimplex, peerSimplex)
				if err3 != nil {
					return err3
				}
				err3 = tx.UpdatePeerCid(peer, cid, true)
				if err3 != nil {
					return err3
				}
				for _, pay := range ch.AuthackPays { // from peer to me. ok to not check if in simplex because no harm
					err3 = insertPaymentFromAuth(tx, pay.Pay, pay.Note, cid, int(pay.State), true)
					if err3 != nil {
						log.Error(err3) // only log no early return for best effort adding pays
					}
				}
				for _, pay := range ch.AuthreqPays { // my pays
					err3 = insertPaymentFromAuth(tx, pay.Pay, pay.Note, cid, int(pay.State), false)
					if err3 != nil {
						log.Error(err3)
					}
				}
				return nil
			} // txF Ends
		} else { // has local view, need to compare local view vs. AuthAck msg
			// it's possible OSP has a newer state eg. from trust_opened to opened
			// but the new state transition must be valid
			if !fsm.IsChanStateChangeValid(myChState, chState) {
				log.Errorf("invalid chState, my: %s, recved: %s", fsm.ChanStateName(myChState), fsm.ChanStateName(chState))
				continue
			}
			txF = func(tx *storage.DALTx, args ...interface{}) error {
				var err3 error
				if myChState != chState {
					err3 = tx.UpdateChanState(cid, chState)
					if err3 != nil {
						return err3
					}
				}
				myLedgerAddr, foundLedger, err3 := tx.GetChanLedger(cid)
				if err3 != nil {
					return err3
				}
				if !foundLedger {
					// This shouldn't happen as we UpdateChanState above making sure existence of cid.
					log.Errorf("Can't find cid %x", cid)
					return fmt.Errorf("Can't find cid %x", cid)
				}
				if peepMyLedgerAddr == myLedgerAddr && shouldUsePeerLedger {
					// make sure local ledger doesn't change from time of peep to now.
					err3 = tx.UpdateChanLedger(cid, peerLedgerAddr)
					if err3 != nil {
						return err3
					}
				}
				if mySimplex != nil {
					mh := c.celerMsgDispatcher.NewMsgHandler()
					err3 = mh.HandleMysimplexFromAuthAck(tx, cid, mySimplex)
					if err3 != nil {
						log.Error(err3) // don't return b/c pays and peerSimplex could still be updated
					}
					// update last added seqNum
					var mySimplexCh entity.SimplexPaymentChannel
					err3 = proto.Unmarshal(mySimplex.GetSimplexState(), &mySimplexCh)
					if err3 != nil {
						log.Error(err3)
					}
					_, lastUsedSeq, lastAckedSeq, lastNackedSeq, found, err4 := tx.GetChanSeqNums(cid)
					if err4 != nil || !found {
						log.Error(err4, found)
					} else if mySimplexCh.GetSeqNum() > lastUsedSeq {
						err3 = tx.UpdateChanSeqNums(cid, mySimplexCh.GetSeqNum(), mySimplexCh.GetSeqNum(), lastAckedSeq, lastNackedSeq)
						if err3 != nil {
							log.Error(err3)
						}
					}
					for _, pay := range ch.AuthreqPays { // my pays
						err3 = insertPaymentFromAuth(tx, pay.Pay, pay.Note, cid, int(pay.State), false)
						if err3 != nil {
							log.Error(err3) // only log err to try adding more pays
						}
					}
				} // mySimplex done
				if peerSimplex != nil {
					// need to sync peer
					peerSeqN, _, err3 := getSeqAndToken(peerSimplex)
					if err3 != nil {
						return err3
					}
					peerChannel, _, found, err3 := tx.GetPeerSimplex(cid)
					if err3 != nil {
						return err3
					}
					if found {
						// peerSeqN must be larger
						if peerSeqN <= peerChannel.GetSeqNum() {
							return common.ErrInvalidSeqNum
						}
					}
					err3 = tx.UpdateChanForRecvRequest(cid, peerSimplex)
					if err3 != nil {
						return err3
					}
					for _, pay := range ch.AuthackPays { // peer pays
						err3 = insertPaymentFromAuth(tx, pay.Pay, pay.Note, cid, int(pay.State), true)
						if err3 != nil {
							log.Error(err3) // only log err to try adding more pays
						}
					}
				} // peerSimplex done
				return nil
			} // txF Ends
		}
		// exec txF
		if err = c.dal.Transactional(txF); err != nil {
			log.Error(err)
			continue
		}
	}
}

// helper struct for seq num in AuthRequest
type seqNum struct {
	MySeq, PeerSeq uint64
}

// get all opened channels with this peer and compare with input, add to return if my seq numbers are bigger or input doesn't have same cid
// this is to support client side missing newly opened channels
func (c *CNode) getSyncChannelsForAuthAck(peer ctype.Addr, input []*rpc.ChannelSummary) []*rpc.ChannelInAuth {
	// save channel Summary list as a map for easy lookup later
	inMap := make(map[ctype.CidType]*seqNum)
	for _, c := range input {
		inMap[ctype.Bytes2Cid(c.GetChannelId())] = &seqNum{
			MySeq:   c.GetMySeqNum(),
			PeerSeq: c.GetPeerSeqNum(),
		}
	}
	// channels is type []*storage.chanForAuthAck
	channels, err := c.dal.GetChannelsForAuthAck(peer)
	if err != nil {
		log.Warnln(err)
	}
	var ret []*rpc.ChannelInAuth
	// only return channels if my seq is higher or client doesn't have it
	for _, ch := range channels {
		var toAdd *rpc.ChannelInAuth // default nil, only malloc when need to add
		cid := ctype.Hex2Cid(ch.Cid)
		if req, ok := inMap[cid]; ok {
			// compare seq and decide whether to add chan
			if ch.MySeq > req.PeerSeq || ch.PeerSeq > req.MySeq { // note the swap of my vs. peer
				// need to add, so malloc now
				toAdd = &rpc.ChannelInAuth{
					Cid:                 cid.Bytes(),
					ChannelState:        ch.State,
					OpenChannelResponse: ch.OpenChanResp, // could be nil
					LedgerAddr:          ch.LedgerAddr.Bytes(),
				}
				log.Infof("sync cid %s peer %s mysimplex [%d, %d] peersimplex [%d, %d]",
					ctype.Cid2Hex(cid), ctype.Addr2Hex(peer), ch.MySeq, req.PeerSeq, ch.PeerSeq, req.MySeq)
			}
			if ch.MySeq > req.PeerSeq {
				toAdd.AuthackSimplex = ch.MySigned
			}
			if ch.PeerSeq > req.MySeq {
				toAdd.AuthreqSimplex = ch.PeerSigned
			}
		} else {
			// client didn't include this cid, should return
			toAdd = &rpc.ChannelInAuth{
				Cid:                 cid.Bytes(),
				ChannelState:        ch.State,
				OpenChannelResponse: ch.OpenChanResp, // could be nil
				AuthackSimplex:      ch.MySigned,
				AuthreqSimplex:      ch.PeerSigned,
				LedgerAddr:          ch.LedgerAddr.Bytes(),
			}
			log.Infof("sync cid %s peer %s due to not in AuthReq", ctype.Cid2Hex(cid), ctype.Addr2Hex(peer))
		}
		if toAdd != nil {
			// add pays to toAdd
			c.addPays(toAdd)
			ret = append(ret, toAdd)
		}
	}
	return ret
}

// go over self/peer simplex in ch and add pays
func (c *CNode) addPays(ch *rpc.ChannelInAuth) {
	if ch.AuthackSimplex != nil {
		payIDs := getPayIDs(ch.AuthackSimplex)
		mypays, err := c.dal.GetPaysForAuthAck(payIDs, true) // pay in my simplex, use outState
		if err != nil {
			log.Warn(err)
		} else {
			ch.AuthackPays = mypays
		}
	}
	if ch.AuthreqSimplex != nil {
		payIDs := getPayIDs(ch.AuthreqSimplex)
		// CANNOT reuse mypays because the actual serialization happens later by grpc/proto runtime
		peerpays, err := c.dal.GetPaysForAuthAck(payIDs, false) // peer's simplex, use instate
		if err != nil {
			log.Warn(err)
		} else {
			ch.AuthreqPays = peerpays
		}
	}
}

// check and handle special seqNum 0 case
// if stateChannel seq is 0, add my sig and remove peer's
func (c *CNode) checkSignedSimplex(ss *rpc.SignedSimplexState, expCid ctype.CidType, expFrom, expTo ctype.Addr) error {
	var ch entity.SimplexPaymentChannel
	err := proto.Unmarshal(ss.GetSimplexState(), &ch)
	if err != nil {
		return err
	}
	if ctype.Bytes2Cid(ch.ChannelId) != expCid {
		return common.ErrInvalidChannelID
	}
	if ctype.Bytes2Addr(ch.PeerFrom) != expFrom {
		return fmt.Errorf("mismatch peerfrom got %s, exp %s", ctype.Bytes2Hex(ch.PeerFrom), ctype.Addr2Hex(expFrom))
	}

	if ch.SeqNum != 0 {
		// must be correctly co-signed by both
		if !eth.IsSignatureValid(expFrom, ss.GetSimplexState(), ss.GetSigOfPeerFrom()) ||
			!eth.IsSignatureValid(expTo, ss.GetSimplexState(), ss.GetSigOfPeerTo()) {
			return common.ErrInvalidSig
		}
		return nil
	}
	// seq num 0 case, simplexChannel must match emptySimplex
	amtBytes := ch.TransferToPeer.GetReceiver().GetAmt()
	pendingBytes := ch.TotalPendingAmount
	if !bytes.Equal(zeroAmtBytes, amtBytes) || !bytes.Equal(zeroAmtBytes, pendingBytes) {
		return common.ErrInvalidAmount
	}
	if len(ch.PendingPayIds.GetPayIds()) != 0 || len(ch.PendingPayIds.GetNextListHash()) != 0 {
		return common.ErrInvalidPendingPays
	}
	mysig, err := c.signer.SignEthMessage(ss.GetSimplexState())
	if err != nil {
		return err
	}
	if expFrom == c.EthAddress {
		ss.SigOfPeerFrom = mysig
		ss.SigOfPeerTo = nil
	} else {
		ss.SigOfPeerFrom = nil
		ss.SigOfPeerTo = mysig
	}
	return nil
}

// getOnChainBalanceForAuth query onchain if channel is opened or populate per init distribution if tcb opened
func (c *CNode) getOnChainBalanceForAuth(cid ctype.CidType, chState int, chanInitBytes []byte, ledgerAddr ctype.Addr) (*structs.OnChainBalance, error) {
	// onchain
	if chState == structs.ChanState_OPENED || chState == structs.ChanState_SETTLING {
		return ledgerview.GetOnChainChannelBalance(cid, ledgerAddr, c.nodeConfig)
	}
	// tcb, don't expect osp has ChanState_INSTANTIATING but add here for completeness
	if chState == structs.ChanState_TRUST_OPENED || chState == structs.ChanState_INSTANTIATING {
		var chanInit entity.PaymentChannelInitializer
		err := proto.Unmarshal(chanInitBytes, &chanInit)
		if err != nil {
			return nil, err
		}
		ret := new(structs.OnChainBalance)
		for _, aap := range chanInit.GetInitDistribution().GetDistribution() {
			addr := ctype.Bytes2Addr(aap.Account)
			amt := new(big.Int).SetBytes(aap.Amt)
			if addr == c.EthAddress {
				ret.MyDeposit = amt
			} else {
				ret.PeerDeposit = amt
			}
		}
		ret.MyWithdrawal = ctype.ZeroBigInt
		ret.PeerWithdrawal = ctype.ZeroBigInt
		return ret, nil
	}
	return nil, common.ErrInvalidChannelState
}

func insertPaymentFromAuth(tx *storage.DALTx, payBytes []byte, note *any.Any, cid ctype.CidType, state int, isIngress bool) error {
	pay := new(entity.ConditionalPay)
	err := proto.Unmarshal(payBytes, pay)
	if err != nil {
		return err
	}
	payID := ctype.Pay2PayID(pay)
	if isIngress { // use inCid and inState
		inCid, myState, found, err2 := tx.GetPayIngress(payID)
		if err2 != nil {
			return err2
		}
		if !found { // not found, just insert
			return tx.InsertPayment(payID, payBytes, pay, note, cid, state, ctype.ZeroCid, structs.PayState_NULL)
		}
		// if found, check data is valid and update if state is diff, otherwise no-op
		if inCid != cid {
			return common.ErrInvalidChannelID
		}
		if myState != state {
			return tx.UpdatePayIngressState(payID, state)
		}
		return nil
	}
	// outCid/outState
	outCid, myState, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return err
	}
	if !found {
		return tx.InsertPayment(payID, payBytes, pay, note, ctype.ZeroCid, structs.PayState_NULL, cid, state)
	}
	if outCid != cid {
		return common.ErrInvalidChannelID
	}
	if myState != state {
		return tx.UpdatePayEgressState(payID, state)
	}
	return nil
}

func getSeqAndToken(ss *rpc.SignedSimplexState) (uint64, *entity.TokenInfo, error) {
	simplexChan := new(entity.SimplexPaymentChannel)
	err := proto.Unmarshal(ss.SimplexState, simplexChan)
	if err != nil {
		return 0, nil, err
	}
	return simplexChan.SeqNum, simplexChan.GetTransferToPeer().GetToken(), nil
}

type msgOrErr struct {
	msg *rpc.CelerMsg
	err error
}

// waitRecvWithTimeout blocks until receive a msg from grpc stream or timeout
// currently this is only used for client waiting AuthAck
func waitRecvWithTimeout(s rpc.CelerStream, timeout time.Duration) (*rpc.CelerMsg, error) {
	ch := make(chan *msgOrErr, 1)
	t := time.NewTimer(timeout)
	go func() {
		m, e := s.Recv()
		ch <- &msgOrErr{
			msg: m,
			err: e,
		}
	}()
	select {
	case <-t.C:
		return nil, common.ErrRecvCelerMsgTimeout
	case recv := <-ch:
		t.Stop()
		return recv.msg, recv.err
	}
}

// util to return payID list from SignedSimplexState
func getPayIDs(simplex *rpc.SignedSimplexState) (payIDs []ctype.PayIDType) {
	simplexChan := new(entity.SimplexPaymentChannel)
	err := proto.Unmarshal(simplex.SimplexState, simplexChan)
	if err != nil {
		log.Warn(err)
		return
	}
	for _, payid := range simplexChan.GetPendingPayIds().GetPayIds() {
		payIDs = append(payIDs, ctype.Bytes2PayID(payid))
	}
	return
}

func isTimestampValid(addr ctype.Addr, ts uint64) bool {
	// we used to require client side ts is within certain window like osp, but some client time is really off
	// now := uint64(time.Now().Unix())
	// return ts >= now-config.AllowedTimeWindow && ts <= now+config.AllowedTimeWindow

	// now we just require ts is larger if we have this client in memory
	lastTs, ok := getTs(addr)
	if ok && ts <= lastTs {
		// found last entry and new req ts isn't larger
		return false
	}
	return true
}

func getTs(key ctype.Addr) (uint64, bool) {
	tsMapLock.RLock()
	t, ok := clientAuthTsMap[key]
	tsMapLock.RUnlock()
	return t, ok
}

func putTs(key ctype.Addr, ts uint64) {
	tsMapLock.Lock()
	clientAuthTsMap[key] = ts
	tsMapLock.Unlock()
}
