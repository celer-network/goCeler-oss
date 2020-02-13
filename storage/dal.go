// Copyright 2018-2019 Celer Network
//
// This is the Data Access Layer. It maps the server's data structures
// that need to be persisted to KVStore calls:
// * Construct table keys from object attribute(s).
// * Use the appropriate Go data types when fetching stored values.

package storage

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/openchannelts"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/route/graph"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	// peer-cid-token. immutable once channel created
	peerTokenCid    = "ptc" // peer:token -> cid
	peerLookUpTable = "plt" // cid -> peer
	tokenAddrTable  = "ta"  // cid -> token addr string

	// mutable channel info
	onChainBalance = "onbal" // cid -> onchainbalance struct
	channelState   = "cfsm"  // cid -> channel state string

	// routing?
	routingTable     = "rt"   // addr:token -> cid
	rtEdgeTable      = "rtet" // addr:token -> edge
	markedOspTable   = "mot"  // ospAddr -> bool
	servingOspsTable = "sosp" // client@token@ospAddr->bool

	// simplex channel state
	coSignedSimplexState = "csss" // owner@cid -> cosigned SignedSimplexState

	// pay related
	conditionalPayTable = "cp"   // payID -> pay bytes
	paymentState        = "pfsm" // payID:in/out -> {cid, pay state string, timestamp}
	secretRegistry      = "sr"   // hash string -> preimage string
	payNoteTable        = "pnt"  // payID -> Any paynote

	// monitor service
	logEventWatch   = "lew" // event name -> LogEventID
	eventMonitorBit = "emb" // event name -> bool

	// OSP only. to avoid handling concurrent openchan requests
	openChannelTs = "oct" // peer@token -> OpenChannelTs

	// channel message queue
	channelMessageQueue = "cmq"  // cid:seqnum -> msg
	channelSeqNums      = "csn"  // cid -> common.ChannelSeqNum
	peerActiveChannels  = "pacs" // peer -> map[cid]true

	// client side openchan request block number, peer:token->last request blkNum
	// used for api to return if has pending openchan request
	openChannelProgressTable = "ocpt"

	transactionalMaxRetry   = 10
	transactionalRetryDelay = 10 * time.Millisecond
)

type DAL struct {
	st KVStore
}

type DALTx struct {
	dal *DAL
	stx Transaction
}

type Storage interface {
	Put(table, key string, value interface{}) error
	Get(table, key string, value interface{}) error
	Delete(table, key string) error
	Has(table, key string) (bool, error)
	GetKeysByPrefix(table, prefix string) ([]string, error)
}

type TxFunc func(tx *DALTx, args ...interface{}) error

func NewDAL(store KVStore) *DAL {
	dal := &DAL{
		st: store,
	}
	return dal
}

func (d *DAL) OpenTransaction() (*DALTx, error) {
	var tx *DALTx
	stx, err := d.st.OpenTransaction()
	if err == nil {
		tx = &DALTx{
			dal: d,
			stx: stx,
		}
	}
	return tx, err
}

func (tx *DALTx) Discard() {
	if tx.stx != nil {
		tx.stx.Discard()
		tx.stx = nil
	}
	tx.dal = nil
}

func (tx *DALTx) Commit() error {
	if tx.stx == nil {
		return ErrTxInvalid
	}

	err := tx.stx.Commit()
	if err == nil {
		tx.stx = nil
	}
	return err
}

func (tx *DALTx) ConvertError(err error) error {
	return tx.stx.ConvertError(err)
}

func (d *DAL) Transactional(callback TxFunc, args ...interface{}) error {
	for i := 0; i < transactionalMaxRetry; i++ {
		tx, err := d.OpenTransaction()
		if err != nil {
			return err
		}

		err = callback(tx, args...)
		if err == nil {
			err = tx.Commit()
			if err == nil {
				return nil
			}
		}

		err = tx.ConvertError(err)
		tx.Discard()
		if err != ErrTxConflict {
			return err
		}

		log.Debugf("transactional: [%d] Tx conflict, retrying...", i)
		time.Sleep(transactionalRetryDelay)
	}

	err := fmt.Errorf("%d Tx commit retries", transactionalMaxRetry)
	log.Error(err)
	return err
}

// DAL for PeerLookUpTable table.

func getPeerLookUpTable(st Storage, cid ctype.CidType) (string, error) {
	var peer string
	err := st.Get(peerLookUpTable, ctype.Cid2Hex(cid), &peer)
	return peer, err
}

func putPeerLookUpTable(st Storage, cid ctype.CidType, peer string) error {
	return st.Put(peerLookUpTable, ctype.Cid2Hex(cid), peer)
}

func deletePeerLookUpTable(st Storage, cid ctype.CidType) error {
	return st.Delete(peerLookUpTable, ctype.Cid2Hex(cid))
}

func hasPeerLookUpTable(st Storage, cid ctype.CidType) (bool, error) {
	return st.Has(peerLookUpTable, ctype.Cid2Hex(cid))
}

func getAllPeerLookUpTableKeys(st Storage) ([]ctype.CidType, error) {
	keystrs, err := st.GetKeysByPrefix(peerLookUpTable, "")
	if err != nil {
		return nil, err
	}
	var keys []ctype.CidType
	for _, key := range keystrs {
		keys = append(keys, ctype.Hex2Cid(key))
	}
	return keys, nil
}

func (d *DAL) GetPeer(cid ctype.CidType) (string, error) {
	return getPeerLookUpTable(d.st, cid)
}

func (d *DAL) PutPeer(cid ctype.CidType, peer string) error {
	return putPeerLookUpTable(d.st, cid, peer)
}

func (d *DAL) DeletePeer(cid ctype.CidType) error {
	return deletePeerLookUpTable(d.st, cid)
}

func (d *DAL) HasPeer(cid ctype.CidType) (bool, error) {
	return hasPeerLookUpTable(d.st, cid)
}

func (d *DAL) GetAllPeerLookUpTableKeys() ([]ctype.CidType, error) {
	return getAllPeerLookUpTableKeys(d.st)
}

func (dtx *DALTx) GetPeer(cid ctype.CidType) (string, error) {
	return getPeerLookUpTable(dtx.stx, cid)
}

func (dtx *DALTx) PutPeer(cid ctype.CidType, peer string) error {
	return putPeerLookUpTable(dtx.stx, cid, peer)
}

func (dtx *DALTx) DeletePeer(cid ctype.CidType) error {
	return deletePeerLookUpTable(dtx.stx, cid)
}

func (dtx *DALTx) HasPeer(cid ctype.CidType) (bool, error) {
	return hasPeerLookUpTable(dtx.stx, cid)
}

func (dtx *DALTx) GetAllPeerLookUpTableKeys() ([]ctype.CidType, error) {
	return getAllPeerLookUpTableKeys(dtx.stx)
}

// DAL for SecretRegistry table.

func getSecretRegistry(st Storage, hash string) (string, error) {
	var preimage string
	err := st.Get(secretRegistry, hash, &preimage)
	return preimage, err
}

func putSecretRegistry(st Storage, hash, preimage string) error {
	return st.Put(secretRegistry, hash, preimage)
}

func deleteSecretRegistry(st Storage, hash string) error {
	return st.Delete(secretRegistry, hash)
}

func hasSecretRegistry(st Storage, hash string) (bool, error) {
	return st.Has(secretRegistry, hash)
}

func getAllSecretRegistryKeys(st Storage) ([]string, error) {
	return st.GetKeysByPrefix(secretRegistry, "")
}

func (d *DAL) GetSecretRegistry(hash string) (string, error) {
	return getSecretRegistry(d.st, hash)
}

func (d *DAL) PutSecretRegistry(hash, preimage string) error {
	return putSecretRegistry(d.st, hash, preimage)
}

func (d *DAL) DeleteSecretRegistry(hash string) error {
	return deleteSecretRegistry(d.st, hash)
}

func (d *DAL) HasSecretRegistry(hash string) (bool, error) {
	return hasSecretRegistry(d.st, hash)
}

func (d *DAL) GetAllSecretRegistryKeys() ([]string, error) {
	return getAllSecretRegistryKeys(d.st)
}

func (dtx *DALTx) GetSecretRegistry(hash string) (string, error) {
	return getSecretRegistry(dtx.stx, hash)
}

func (dtx *DALTx) PutSecretRegistry(hash, preimage string) error {
	return putSecretRegistry(dtx.stx, hash, preimage)
}

func (dtx *DALTx) DeleteSecretRegistry(hash string) error {
	return deleteSecretRegistry(dtx.stx, hash)
}

func (dtx *DALTx) HasSecretRegistry(hash string) (bool, error) {
	return hasSecretRegistry(dtx.stx, hash)
}

func (dtx *DALTx) GetAllSecretRegistryKeys() ([]string, error) {
	return getAllSecretRegistryKeys(dtx.stx)
}

// DAL for RoutingTable table.

func getRoute(st Storage, destAndTokenAddr string) (ctype.CidType, error) {
	var nextHop ctype.CidType
	err := st.Get(routingTable, destAndTokenAddr, &nextHop)
	return nextHop, err
}

func putRoute(st Storage, destAndTokenAddr string, nextHop ctype.CidType) error {
	return st.Put(routingTable, destAndTokenAddr, nextHop)
}

func deleteRoute(st Storage, destAndTokenAddr string) error {
	return st.Delete(routingTable, destAndTokenAddr)
}

func hasRoute(st Storage, destAndTokenAddr string) (bool, error) {
	return st.Has(routingTable, destAndTokenAddr)
}

func getAllRoutingTableKeys(st Storage) ([]string, error) {
	return st.GetKeysByPrefix(routingTable, "")
}

func getAllRoutingTableKeysToDest(st Storage, dest string) ([]string, error) {
	return st.GetKeysByPrefix(routingTable, dest)
}

func (d *DAL) GetRoute(dest, tokenAddr string) (ctype.CidType, error) {
	// Assemble dest and token key in form "dest@tokenAddr"
	destAndTokenAddr := dest + "@" + tokenAddr
	return getRoute(d.st, destAndTokenAddr)
}

func (d *DAL) PutRoute(dest, tokenAddr string, nextHop ctype.CidType) error {
	destAndTokenAddr := dest + "@" + tokenAddr
	return putRoute(d.st, destAndTokenAddr, nextHop)
}

func (d *DAL) DeleteRoute(dest, tokenAddr string) error {
	destAndTokenAddr := dest + "@" + tokenAddr
	return deleteRoute(d.st, destAndTokenAddr)
}

func (d *DAL) HasRoute(dest, tokenAddr string) (bool, error) {
	destAndTokenAddr := dest + "@" + tokenAddr
	return hasRoute(d.st, destAndTokenAddr)
}

func (d *DAL) GetAllRoutingTableKeys() ([]string, error) {
	return getAllRoutingTableKeys(d.st)
}

func (d *DAL) GetAllRoutingTableKeysToDest(dest string) ([]string, error) {
	return getAllRoutingTableKeysToDest(d.st, dest)
}

func (dtx *DALTx) GetRoute(dest, tokenAddr string) (ctype.CidType, error) {
	// Assemble dest and token key in form "dest@tokenAddr"
	destAndTokenAddr := dest + "@" + tokenAddr
	return getRoute(dtx.stx, destAndTokenAddr)
}

func (dtx *DALTx) PutRoute(dest, tokenAddr string, nextHop ctype.CidType) error {
	destAndTokenAddr := dest + "@" + tokenAddr
	return putRoute(dtx.stx, destAndTokenAddr, nextHop)
}

func (dtx *DALTx) DeleteRoute(dest, tokenAddr string) error {
	destAndTokenAddr := dest + "@" + tokenAddr
	return deleteRoute(dtx.stx, destAndTokenAddr)
}

func (dtx *DALTx) HasRoute(dest, tokenAddr string) (bool, error) {
	destAndTokenAddr := dest + "@" + tokenAddr
	return hasRoute(dtx.stx, destAndTokenAddr)
}

func (dtx *DALTx) GetAllRoutingTableKeys() ([]string, error) {
	return getAllRoutingTableKeys(dtx.stx)
}

func (dtx *DALTx) GetAllRoutingTableKeysToDest(dest string) ([]string, error) {
	return getAllRoutingTableKeysToDest(dtx.stx, dest)
}

// PendingOpenChannel
func hasOpenChannelTs(st Storage, peerAddr ctype.Addr, tokenAddr ctype.Addr) (bool, error) {
	return st.Has(openChannelTs, ctype.Addr2Hex(peerAddr)+"@"+ctype.Addr2Hex(tokenAddr))
}
func deleteOpenChannelTs(st Storage, peerAddr ctype.Addr, tokenAddr ctype.Addr) error {
	return st.Delete(openChannelTs, ctype.Addr2Hex(peerAddr)+"@"+ctype.Addr2Hex(tokenAddr))
}
func getOpenChannelTs(st Storage, peerAddr ctype.Addr, tokenAddr ctype.Addr) (*openchannelts.OpenChannelTs, error) {
	var ts openchannelts.OpenChannelTs
	err := st.Get(openChannelTs, ctype.Addr2Hex(peerAddr)+"@"+ctype.Addr2Hex(tokenAddr), &ts)
	return &ts, err
}
func putOpenChannelTs(st Storage, peerAddr ctype.Addr, tokenAddr ctype.Addr, ts *openchannelts.OpenChannelTs) error {
	return st.Put(openChannelTs, ctype.Addr2Hex(peerAddr)+"@"+ctype.Addr2Hex(tokenAddr), ts)
}
func (dtx *DALTx) HasOpenChannelTs(peerAddr ctype.Addr, tokenAddr ctype.Addr) (bool, error) {
	return hasOpenChannelTs(dtx.stx, peerAddr, tokenAddr)
}
func (dtx *DALTx) DeleteOpenChannelTs(peerAddr ctype.Addr, tokenAddr ctype.Addr) error {
	return deleteOpenChannelTs(dtx.stx, peerAddr, tokenAddr)
}
func (dtx *DALTx) GetOpenChannelTs(peerAddr ctype.Addr, tokenAddr ctype.Addr) (*openchannelts.OpenChannelTs, error) {
	return getOpenChannelTs(dtx.stx, peerAddr, tokenAddr)
}
func (dtx *DALTx) PutOpenChannelTs(peerAddr ctype.Addr, tokenAddr ctype.Addr, openChannelTs *openchannelts.OpenChannelTs) error {
	return putOpenChannelTs(dtx.stx, peerAddr, tokenAddr, openChannelTs)
}
func getSimplexPaymentChannel(st Storage, ownerAndCid string) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, error) {
	var simplexChannel entity.SimplexPaymentChannel
	var simplexState rpc.SignedSimplexState
	err := st.Get(coSignedSimplexState, ownerAndCid, &simplexState)
	if err != nil {
		err = fmt.Errorf("%s NO_SIMPLEX_STATE", ownerAndCid)
		return nil, nil, err
	}
	err = proto.Unmarshal(simplexState.SimplexState, &simplexChannel)
	if err != nil {
		return nil, &simplexState, errors.New("CANNOT_PARSE_SIMPLEX_CHANNEL")
	}
	return &simplexChannel, &simplexState, err
}

func putSimplexState(st Storage, ownerAndCid string, simplexState *rpc.SignedSimplexState) error {
	return st.Put(coSignedSimplexState, ownerAndCid, simplexState)
}

func deleteSimplexState(st Storage, ownerAndCid string) error {
	return st.Delete(coSignedSimplexState, ownerAndCid)
}

func (d *DAL) GetSimplexPaymentChannel(cid ctype.CidType, owner string) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, error) {
	return getSimplexPaymentChannel(d.st, owner+"@"+ctype.Cid2Hex(cid))
}

func (d *DAL) PutSimplexState(cid ctype.CidType, owner string, simplexState *rpc.SignedSimplexState) error {
	return putSimplexState(d.st, owner+"@"+ctype.Cid2Hex(cid), simplexState)
}

func (d *DAL) DeleteSimplexState(cid ctype.CidType, owner string) error {
	return deleteSimplexState(d.st, owner+"@"+ctype.Cid2Hex(cid))
}

func (dtx *DALTx) GetSimplexPaymentChannel(cid ctype.CidType, owner string) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, error) {
	return getSimplexPaymentChannel(dtx.stx, owner+"@"+ctype.Cid2Hex(cid))
}

func (dtx *DALTx) PutSimplexState(cid ctype.CidType, owner string, simplexState *rpc.SignedSimplexState) error {
	return putSimplexState(dtx.stx, owner+"@"+ctype.Cid2Hex(cid), simplexState)
}

func (dtx *DALTx) DeleteSimplexState(cid ctype.CidType, owner string) error {
	return deleteSimplexState(dtx.stx, owner+"@"+ctype.Cid2Hex(cid))
}

func putPayNote(st Storage, condPay *entity.ConditionalPay, note *any.Any) error {
	payID := ctype.Pay2PayID(condPay)
	if exist, _ := st.Has(payNoteTable, ctype.PayID2Hex(payID)); exist {
		return errors.New("PAY_NOTE_EXIST")
	}
	return st.Put(payNoteTable, ctype.PayID2Hex(payID), note)
}

func getPayNote(st Storage, payID ctype.PayIDType) (*any.Any, error) {
	note := &any.Any{}
	err := st.Get(payNoteTable, ctype.PayID2Hex(payID), &note)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (d *DAL) PutPayNote(condPay *entity.ConditionalPay, note *any.Any) error {
	return putPayNote(d.st, condPay, note)
}

func (dtx *DALTx) PutPayNote(condPay *entity.ConditionalPay, note *any.Any) error {
	return putPayNote(dtx.stx, condPay, note)
}

func (d *DAL) GetPayNote(payID ctype.PayIDType) (*any.Any, error) {
	return getPayNote(d.st, payID)
}

func (dtx *DALTx) GetPayNote(payID ctype.PayIDType) (*any.Any, error) {
	return getPayNote(dtx.stx, payID)
}

// save payBytes, autogen key from hash, skip if already exists
func putConditionalPay(st Storage, payBytes []byte) error {
	var pay entity.ConditionalPay
	err := proto.Unmarshal(payBytes, &pay)
	if err != nil {
		return err
	}
	payID := ctype.Pay2PayID(&pay)
	k := ctype.PayID2Hex(payID) // has 0x prefix
	if exist, _ := st.Has(conditionalPayTable, k); exist {
		return nil
	}
	return st.Put(conditionalPayTable, k, payBytes)
}

// save payBytes, autogen key from hash, skip if already exists
func (d *DAL) PutConditionalPay(payBytes []byte) error {
	return putConditionalPay(d.st, payBytes)
}

func (dtx *DALTx) PutConditionalPay(payBytes []byte) error {
	return putConditionalPay(dtx.stx, payBytes)
}

func getConditionalPay(st Storage, payID ctype.PayIDType) (*entity.ConditionalPay, []byte, error) {
	var condPayBytes []byte
	err := st.Get(conditionalPayTable, ctype.PayID2Hex(payID), &condPayBytes)
	if err != nil {
		return nil, nil, err
	}
	var condPay entity.ConditionalPay
	err = proto.Unmarshal(condPayBytes, &condPay)
	if err != nil {
		return nil, nil, err
	}
	return &condPay, condPayBytes, nil
}

func (d *DAL) GetConditionalPay(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, error) {
	return getConditionalPay(d.st, payID)
}

func (dtx *DALTx) GetConditionalPay(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, error) {
	return getConditionalPay(dtx.stx, payID)
}

func deleteConditionalPay(st Storage, payID ctype.PayIDType) error {
	return st.Delete(conditionalPayTable, ctype.PayID2Hex(payID))
}

func (d *DAL) DeleteConditionalPay(payID ctype.PayIDType) error {
	return deleteConditionalPay(d.st, payID)
}

func (dtx *DALTx) DeleteConditionalPay(payID ctype.PayIDType) error {
	return deleteConditionalPay(dtx.stx, payID)
}

func getAllConditionalPays(st Storage) ([]*entity.ConditionalPay, error) {
	keys, err := st.GetKeysByPrefix(conditionalPayTable, "")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	pays := make([]*entity.ConditionalPay, 0, len(keys))
	for _, payIDStr := range keys {
		// var payID ctype.PayIDType
		payID := ctype.Hex2PayID(payIDStr)
		pay, _, err := getConditionalPay(st, payID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		pays = append(pays, pay)
	}
	return pays, nil
}

func (d *DAL) GetAllConditionalPays() ([]*entity.ConditionalPay, error) {
	return getAllConditionalPays(d.st)
}

func (dtx *DALTx) GetAllConditionalPays() ([]*entity.ConditionalPay, error) {
	return getAllConditionalPays(dtx.stx)
}

// DAL for payment state machine

type PayState struct {
	Cid       ctype.CidType
	Status    string
	Timestamp int64
}

func payStateKey(payID ctype.PayIDType, direction string) string {
	return ctype.PayID2Hex(payID) + ":" + direction
}

// return channel ID, payment status, update timestamp, err message
func getPaymentState(st Storage, payID ctype.PayIDType, direction string) (ctype.CidType, string, int64, error) {
	var ps PayState
	err := st.Get(paymentState, payStateKey(payID, direction), &ps)
	return ps.Cid, ps.Status, ps.Timestamp, err
}

func putPaymentState(st Storage, payID ctype.PayIDType, direction string, cid ctype.CidType, status string) error {
	ps := &PayState{
		Cid:       cid,
		Status:    status,
		Timestamp: time.Now().Unix(),
	}
	return st.Put(paymentState, payStateKey(payID, direction), ps)
}

func deletePaymentState(st Storage, payID ctype.PayIDType, direction string) error {
	return st.Delete(paymentState, payStateKey(payID, direction))
}

func hasPaymentState(st Storage, payID ctype.PayIDType, direction string) (bool, error) {
	return st.Has(paymentState, payStateKey(payID, direction))
}

func getAllPaymentStateKeys(st Storage) ([]ctype.PayIDType, []ctype.PayIDType, error) {
	keystrs, err := st.GetKeysByPrefix(paymentState, "")
	if err != nil {
		return nil, nil, err
	}
	var ingress []ctype.PayIDType
	var egress []ctype.PayIDType
	for _, key := range keystrs {
		s := strings.Split(key, ":")
		if len(s) != 2 {
			continue
		}
		payID := ctype.Hex2PayID(s[0])
		dir := s[1]
		if dir == "in" {
			ingress = append(ingress, payID)
		} else if dir == "out" {
			egress = append(egress, payID)
		}

	}
	return ingress, egress, nil
}

// GetPayIngressState returns pay ingress cid, status, timestamp, err for a given payID
func (d *DAL) GetPayIngressState(payID ctype.PayIDType) (ctype.CidType, string, int64, error) {
	return getPaymentState(d.st, payID, "in")
}

// GetPayEgressState returns pay egress cid, status, timestamp, err for a given payID
func (d *DAL) GetPayEgressState(payID ctype.PayIDType) (ctype.CidType, string, int64, error) {
	return getPaymentState(d.st, payID, "out")
}

// PutPayIngressState write pay ingress cid, status for a given payID
func (d *DAL) PutPayIngressState(payID ctype.PayIDType, cid ctype.CidType, status string) error {
	return putPaymentState(d.st, payID, "in", cid, status)
}

// PutPayEgressState write pay egress cid, status for a given payID
func (d *DAL) PutPayEgressState(payID ctype.PayIDType, cid ctype.CidType, status string) error {
	return putPaymentState(d.st, payID, "out", cid, status)
}

func (d *DAL) HasPayIngressState(payID ctype.PayIDType) (bool, error) {
	return hasPaymentState(d.st, payID, "in")
}

func (d *DAL) HasPayEgressState(payID ctype.PayIDType) (bool, error) {
	return hasPaymentState(d.st, payID, "out")
}

func (d *DAL) DeletePayIngressState(payID ctype.PayIDType) error {
	return deletePaymentState(d.st, payID, "in")
}

func (d *DAL) DeletePayEgressState(payID ctype.PayIDType) error {
	return deletePaymentState(d.st, payID, "out")
}

func (d *DAL) GetAllPaymentStateKeys() ([]ctype.PayIDType, []ctype.PayIDType, error) {
	return getAllPaymentStateKeys(d.st)
}

// GetPayIngressState returns pay ingress cid, status, timestamp, err for a given payID
func (dtx *DALTx) GetPayIngressState(payID ctype.PayIDType) (ctype.CidType, string, int64, error) {
	return getPaymentState(dtx.stx, payID, "in")
}

// GetPayEgressState returns pay egress cid, status, timestamp, err for a given payID
func (dtx *DALTx) GetPayEgressState(payID ctype.PayIDType) (ctype.CidType, string, int64, error) {
	return getPaymentState(dtx.stx, payID, "out")
}

// PutPayIngressState write pay ingress cid, status for a given payID
func (dtx *DALTx) PutPayIngressState(payID ctype.PayIDType, cid ctype.CidType, status string) error {
	return putPaymentState(dtx.stx, payID, "in", cid, status)
}

// PutPayEgressState write pay egress cid, status for a given payID
func (dtx *DALTx) PutPayEgressState(payID ctype.PayIDType, cid ctype.CidType, status string) error {
	return putPaymentState(dtx.stx, payID, "out", cid, status)
}

func (dtx *DALTx) HasPayIngressState(payID ctype.PayIDType) (bool, error) {
	return hasPaymentState(dtx.stx, payID, "in")
}

func (dtx *DALTx) HasPayEgressState(payID ctype.PayIDType) (bool, error) {
	return hasPaymentState(dtx.stx, payID, "out")
}

func (dtx *DALTx) DeletePayIngressState(payID ctype.PayIDType) error {
	return deletePaymentState(dtx.stx, payID, "in")
}

func (dtx *DALTx) DeletePayEgressState(payID ctype.PayIDType) error {
	return deletePaymentState(dtx.stx, payID, "out")
}

func (dtx *DALTx) GetAllPaymentStateKeys() ([]ctype.PayIDType, []ctype.PayIDType, error) {
	return getAllPaymentStateKeys(dtx.stx)
}

// DAL for Log Event Watch table.

func getLogEventWatch(st Storage, name string) (*structs.LogEventID, error) {
	var id structs.LogEventID
	err := st.Get(logEventWatch, name, &id)
	return &id, err
}

func putLogEventWatch(st Storage, name string, id *structs.LogEventID) error {
	return st.Put(logEventWatch, name, id)
}

func deleteLogEventWatch(st Storage, name string) error {
	return st.Delete(logEventWatch, name)
}

func hasLogEventWatch(st Storage, name string) (bool, error) {
	return st.Has(logEventWatch, name)
}

func getAllLogEventWatchKeys(st Storage) ([]string, error) {
	return st.GetKeysByPrefix(logEventWatch, "")
}

func (d *DAL) GetLogEventWatch(name string) (*structs.LogEventID, error) {
	return getLogEventWatch(d.st, name)
}

func (d *DAL) PutLogEventWatch(name string, id *structs.LogEventID) error {
	return putLogEventWatch(d.st, name, id)
}

func (d *DAL) DeleteLogEventWatch(name string) error {
	return deleteLogEventWatch(d.st, name)
}

func (d *DAL) HasLogEventWatch(name string) (bool, error) {
	return hasLogEventWatch(d.st, name)
}

func (d *DAL) GetAllLogEventWatchKeys() ([]string, error) {
	return getAllLogEventWatchKeys(d.st)
}

func (dtx *DALTx) GetLogEventWatch(name string) (*structs.LogEventID, error) {
	return getLogEventWatch(dtx.stx, name)
}

func (dtx *DALTx) PutLogEventWatch(name string, id *structs.LogEventID) error {
	return putLogEventWatch(dtx.stx, name, id)
}

func (dtx *DALTx) DeleteLogEventWatch(name string) error {
	return deleteLogEventWatch(dtx.stx, name)
}

func (dtx *DALTx) HasLogEventWatch(name string) (bool, error) {
	return hasLogEventWatch(dtx.stx, name)
}

func (dtx *DALTx) GetAllLogEventWatchKeys() ([]string, error) {
	return getAllLogEventWatchKeys(dtx.stx)
}

// DAL for event monitor bit.

func hasEventMonitorBit(st Storage, eventName string) (bool, error) {
	return st.Has(eventMonitorBit, eventName)
}

func putEventMonitorBit(st Storage, eventName string) error {
	return st.Put(eventMonitorBit, eventName, true)
}

func deleteEventMonitorBit(st Storage, eventName string) error {
	return st.Delete(eventMonitorBit, eventName)
}

func (d *DAL) HasEventMonitorBit(eventName string) (bool, error) {
	return hasEventMonitorBit(d.st, eventName)
}

func (d *DAL) PutEventMonitorBit(eventName string) error {
	return putEventMonitorBit(d.st, eventName)
}

func (d *DAL) DeleteEventMonitorBit(eventName string) error {
	return deleteEventMonitorBit(d.st, eventName)
}

func (dtx *DALTx) HasEventMonitorBit(eventName string) (bool, error) {
	return hasEventMonitorBit(dtx.stx, eventName)
}

func (dtx *DALTx) PutEventMonitorBit(eventName string) error {
	return putEventMonitorBit(dtx.stx, eventName)
}

func (dtx *DALTx) DeleteEventMonitorBit(eventName string) error {
	return deleteEventMonitorBit(dtx.stx, eventName)
}

// DAL for payment channel state machine

type ChanState struct {
	State     string
	Timestamp int64
}

// return channel state, timestamp, err
func getChannelState(st Storage, cid ctype.CidType) (string, int64, error) {
	var cs ChanState
	err := st.Get(channelState, ctype.Cid2Hex(cid), &cs)
	if err != nil {
		return "", 0, err
	}
	return cs.State, cs.Timestamp, err
}

func putChannelState(st Storage, cid ctype.CidType, state string) error {
	cs := &ChanState{
		State:     state,
		Timestamp: time.Now().Unix(),
	}
	return st.Put(channelState, ctype.Cid2Hex(cid), cs)
}

func deleteChannelState(st Storage, cid ctype.CidType) error {
	return st.Delete(channelState, ctype.Cid2Hex(cid))
}

func hasChannelState(st Storage, cid ctype.CidType) (bool, error) {
	return st.Has(channelState, ctype.Cid2Hex(cid))
}

func getAllChannelStateKeys(st Storage) ([]ctype.CidType, error) {
	keystrs, err := st.GetKeysByPrefix(channelState, "")
	if err != nil {
		return nil, err
	}
	var keys []ctype.CidType
	for _, key := range keystrs {
		keys = append(keys, ctype.Hex2Cid(key))
	}
	return keys, nil
}

// GetChannelState returns channel status, last update timestamp, and err msg
func (d *DAL) GetChannelState(cid ctype.CidType) (string, int64, error) {
	return getChannelState(d.st, cid)
}

func (d *DAL) PutChannelState(cid ctype.CidType, state string) error {
	return putChannelState(d.st, cid, state)
}

func (d *DAL) HasChannelState(cid ctype.CidType) (bool, error) {
	return hasChannelState(d.st, cid)
}

func (d *DAL) DeleteChannelState(cid ctype.CidType) error {
	return deleteChannelState(d.st, cid)
}

func (d *DAL) GetAllChannelStateKeys() ([]ctype.CidType, error) {
	return getAllChannelStateKeys(d.st)
}

// GetChannelState returns channel status, last update timestamp, and err msg
func (dtx *DALTx) GetChannelState(cid ctype.CidType) (string, int64, error) {
	return getChannelState(dtx.stx, cid)
}

func (dtx *DALTx) PutChannelState(cid ctype.CidType, state string) error {
	return putChannelState(dtx.stx, cid, state)
}

func (dtx *DALTx) HasChannelState(cid ctype.CidType) (bool, error) {
	return hasChannelState(dtx.stx, cid)
}

func (dtx *DALTx) DeleteChannelState(cid ctype.CidType) error {
	return deleteChannelState(dtx.stx, cid)
}

func (dtx *DALTx) GetAllChannelStateKeys() ([]ctype.CidType, error) {
	return getAllChannelStateKeys(dtx.stx)
}

// DAL for Token Contract Address table.

func getTokenContractAddr(st Storage, cid ctype.CidType) (string, error) {
	var addr string
	err := st.Get(tokenAddrTable, ctype.Cid2Hex(cid), &addr)
	return addr, err
}

func putTokenContractAddr(st Storage, cid ctype.CidType, addr string) error {
	return st.Put(tokenAddrTable, ctype.Cid2Hex(cid), addr)
}

func (d *DAL) GetTokenContractAddr(cid ctype.CidType) (string, error) {
	return getTokenContractAddr(d.st, cid)
}

func (d *DAL) PutTokenContractAddr(cid ctype.CidType, addr string) error {
	return putTokenContractAddr(d.st, cid, addr)
}

func (dtx *DALTx) GetTokenContractAddr(cid ctype.CidType) (string, error) {
	return getTokenContractAddr(dtx.stx, cid)
}

func (dtx *DALTx) PutTokenContractAddr(cid ctype.CidType, addr string) error {
	return putTokenContractAddr(dtx.stx, cid, addr)
}

// DAL for client openchannel block number
// openchannelprogress peer:token -> blkNum
func getLastOpenChanReqBlkNum(st Storage, peer []byte, token *entity.TokenInfo) (int64, error) {
	var ret int64
	err := st.Get(openChannelProgressTable, ctype.Bytes2Hex(peer)+":"+utils.GetTokenAddrStr(token), &ret)
	return ret, err
}

func putLastOpenChanReqBlkNum(st Storage, peer []byte, token *entity.TokenInfo, blkNum int64) error {
	return st.Put(openChannelProgressTable, ctype.Bytes2Hex(peer)+":"+utils.GetTokenAddrStr(token), blkNum)
}
func (d *DAL) GetLastOpenChanReqBlkNum(peer []byte, token *entity.TokenInfo) (int64, error) {
	return getLastOpenChanReqBlkNum(d.st, peer, token)
}
func (dtx *DALTx) GetLastOpenChanReqBlkNum(peer []byte, token *entity.TokenInfo) (int64, error) {
	return getLastOpenChanReqBlkNum(dtx.stx, peer, token)
}
func (d *DAL) PutLastOpenChanReqBlkNum(peer []byte, token *entity.TokenInfo, blkNum int64) error {
	return putLastOpenChanReqBlkNum(d.st, peer, token, blkNum)
}
func (dtx *DALTx) PutLastOpenChanReqBlkNum(peer []byte, token *entity.TokenInfo, blkNum int64) error {
	return putLastOpenChanReqBlkNum(dtx.stx, peer, token, blkNum)
}

// DAL for peer:token -> cid

// putCidForPeerAndToken writes to DB an entry peer:token -> cid
func putCidForPeerAndToken(st Storage, peer []byte, token *entity.TokenInfo, cid ctype.CidType) error {
	return st.Put(peerTokenCid, ctype.Bytes2Hex(peer)+":"+utils.GetTokenAddrStr(token), cid)
}

// getCidByPeerAndToken get the cid from peer and token
// if cid is 0, means error/not found, b/c cid is never 0
func getCidByPeerAndToken(st Storage, peer []byte, token *entity.TokenInfo) (cid ctype.CidType) {
	st.Get(peerTokenCid, ctype.Bytes2Hex(peer)+":"+utils.GetTokenAddrStr(token), &cid)
	return cid
}

// scanCidsByPeer calls GetKeysByPrefix which is expensive
// we only expect this to be called when no token info (first ever client conn or recovery case)
// return empty list if no cid exists or error
func scanCidsByPeer(st Storage, peer []byte) (cids []ctype.CidType) {
	keys, err := st.GetKeysByPrefix(peerTokenCid, ctype.Bytes2Hex(peer)+":")
	if err != nil {
		return cids
	}
	var cid ctype.CidType
	for _, k := range keys {
		err = st.Get(peerTokenCid, k, &cid)
		if err != nil {
			log.Error("scancidsbypeer err: ", err)
		}
		if cid != ctype.ZeroCid {
			cids = append(cids, cid)
		}
	}
	return cids
}

func (d *DAL) GetCidByPeerAndTokenWithErr(peer []byte, token *entity.TokenInfo) (ctype.CidType, bool, error) {
	var cid ctype.CidType
	key := ctype.Bytes2Hex(peer) + ":" + utils.GetTokenAddrStr(token)
	has, err := d.st.Has(peerTokenCid, key)
	if err != nil || !has {
		return cid, false, err
	}
	err = d.st.Get(peerTokenCid, key, &cid)
	return cid, true, err
}

// Ideally we should enforce peer type as well, use []byte is easy for value from proto msg
func (d *DAL) GetCidByPeerAndToken(peer []byte, token *entity.TokenInfo) ctype.CidType {
	return getCidByPeerAndToken(d.st, peer, token)
}

func (d *DAL) PutCidForPeerAndToken(peer []byte, token *entity.TokenInfo, cid ctype.CidType) error {
	return putCidForPeerAndToken(d.st, peer, token, cid)
}

// ScanAllCidsByPeer does key scan and is expensive!
// only hello, sync_db and messenger/queue should call this to proper handle old db schema transition to new one that has peerActiveChannels table
// new code should only use peerActiveChannels
func (d *DAL) ScanAllCidsByPeer(peer []byte) []ctype.CidType {
	return scanCidsByPeer(d.st, peer)
}

func (dtx *DALTx) GetCidByPeerAndToken(peer []byte, token *entity.TokenInfo) ctype.CidType {
	return getCidByPeerAndToken(dtx.stx, peer, token)
}

func (dtx *DALTx) PutCidForPeerAndToken(peer []byte, token *entity.TokenInfo, cid ctype.CidType) error {
	return putCidForPeerAndToken(dtx.stx, peer, token, cid)
}

func (dtx *DALTx) ScanAllCidsByPeer(peer []byte) []ctype.CidType {
	return scanCidsByPeer(dtx.stx, peer)
}

// DAL for on chain balances
type OnChainBalance struct {
	MyDeposit         []byte
	MyWithdrawal      []byte
	PeerDeposit       []byte
	PeerWithdrawal    []byte
	PendingWithdrawal *PendingWithdrawal
}

type PendingWithdrawal struct {
	Amount   []byte
	Receiver ctype.Addr
	Deadline uint64
}

func getOnChainBalance(st Storage, cid ctype.CidType) (*structs.OnChainBalance, error) {
	var balance OnChainBalance
	err := st.Get(onChainBalance, ctype.Cid2Hex(cid), &balance)
	if err != nil {
		return nil, err
	}
	pendingWithdrawal := &structs.PendingWithdrawal{Amount: ctype.ZeroBigInt}
	if balance.PendingWithdrawal != nil {
		pendingWithdrawal = &structs.PendingWithdrawal{
			Amount:   new(big.Int).SetBytes(balance.PendingWithdrawal.Amount),
			Receiver: balance.PendingWithdrawal.Receiver,
			Deadline: balance.PendingWithdrawal.Deadline,
		}
	}
	return &structs.OnChainBalance{
		MyDeposit:         new(big.Int).SetBytes(balance.MyDeposit),
		MyWithdrawal:      new(big.Int).SetBytes(balance.MyWithdrawal),
		PeerDeposit:       new(big.Int).SetBytes(balance.PeerDeposit),
		PeerWithdrawal:    new(big.Int).SetBytes(balance.PeerWithdrawal),
		PendingWithdrawal: pendingWithdrawal,
	}, nil
}

func putOnChainBalance(st Storage, cid ctype.CidType, balance *structs.OnChainBalance) error {
	var pendingWithdrawal PendingWithdrawal
	if balance.PendingWithdrawal != nil {
		pendingWithdrawal.Amount = balance.PendingWithdrawal.Amount.Bytes()
		pendingWithdrawal.Receiver = balance.PendingWithdrawal.Receiver
		pendingWithdrawal.Deadline = balance.PendingWithdrawal.Deadline
	}
	bal := &OnChainBalance{
		MyDeposit:         balance.MyDeposit.Bytes(),
		MyWithdrawal:      balance.MyWithdrawal.Bytes(),
		PeerDeposit:       balance.PeerDeposit.Bytes(),
		PeerWithdrawal:    balance.PeerWithdrawal.Bytes(),
		PendingWithdrawal: &pendingWithdrawal,
	}
	return st.Put(onChainBalance, ctype.Cid2Hex(cid), bal)
}

func deleteOnChainBalance(st Storage, cid ctype.CidType) error {
	return st.Delete(onChainBalance, ctype.Cid2Hex(cid))
}

func hasOnChainBalance(st Storage, cid ctype.CidType) (bool, error) {
	return st.Has(onChainBalance, ctype.Cid2Hex(cid))
}

func getAllOnChainBalanceKeys(st Storage) ([]ctype.CidType, error) {
	keystrs, err := st.GetKeysByPrefix(onChainBalance, "")
	if err != nil {
		return nil, err
	}
	var keys []ctype.CidType
	for _, key := range keystrs {
		keys = append(keys, ctype.Hex2Cid(key))
	}
	return keys, nil
}

func (d *DAL) GetOnChainBalance(cid ctype.CidType) (*structs.OnChainBalance, error) {
	return getOnChainBalance(d.st, cid)
}

func (d *DAL) PutOnChainBalance(cid ctype.CidType, balance *structs.OnChainBalance) error {
	return putOnChainBalance(d.st, cid, balance)
}

func (d *DAL) HasOnChainBalance(cid ctype.CidType) (bool, error) {
	return hasOnChainBalance(d.st, cid)
}

func (d *DAL) DeleteOnChainBalance(cid ctype.CidType) error {
	return deleteOnChainBalance(d.st, cid)
}

func (d *DAL) GetAllOnChainBalanceKeys() ([]ctype.CidType, error) {
	return getAllOnChainBalanceKeys(d.st)
}

func (dtx *DALTx) GetOnChainBalance(cid ctype.CidType) (*structs.OnChainBalance, error) {
	return getOnChainBalance(dtx.stx, cid)
}

func (dtx *DALTx) PutOnChainBalance(cid ctype.CidType, balance *structs.OnChainBalance) error {
	return putOnChainBalance(dtx.stx, cid, balance)
}

func (dtx *DALTx) HasOnChainBalance(cid ctype.CidType) (bool, error) {
	return hasOnChainBalance(dtx.stx, cid)
}

func (dtx *DALTx) DeleteOnChainBalance(cid ctype.CidType) error {
	return deleteOnChainBalance(dtx.stx, cid)
}

func (dtx *DALTx) GetAllOnChainBalanceKeys() ([]ctype.CidType, error) {
	return getAllOnChainBalanceKeys(dtx.stx)
}

func putEdge(st Storage, token ctype.Addr, cid ctype.CidType, edge *graph.Edge) error {
	key := fmt.Sprintf("%x:%x", token.Bytes(), cid.Bytes())
	return st.Put(rtEdgeTable, key, edge)
}
func getEdges(st Storage, token ctype.Addr) (map[ctype.CidType]*graph.Edge, error) {
	keyPrefix := fmt.Sprintf("%x:", token.Bytes())
	keys, err := st.GetKeysByPrefix(rtEdgeTable, keyPrefix)
	if err != nil {
		return nil, err
	}
	edges := make(map[ctype.CidType]*graph.Edge)
	for _, key := range keys {
		cidStr := strings.TrimPrefix(key, keyPrefix)
		cid := ctype.Hex2Cid(cidStr)
		edge := &graph.Edge{}
		getErr := st.Get(rtEdgeTable, key, edge)
		if getErr != nil {
			return nil, getErr
		}
		edges[cid] = edge
	}
	return edges, nil
}

func (d *DAL) PutEdge(token ctype.Addr, cid ctype.CidType, edge *graph.Edge) error {
	return putEdge(d.st, token, cid, edge)
}
func (d *DAL) GetEdges(token ctype.Addr) (map[ctype.CidType]*graph.Edge, error) {
	return getEdges(d.st, token)
}
func (d *DAL) DeleteEdge(token ctype.Addr, cid ctype.CidType) error {
	key := fmt.Sprintf("%x:%x", token.Bytes(), cid.Bytes())
	return d.st.Delete(rtEdgeTable, key)
}
func (d *DAL) GetAllEdgeTokens() (map[ctype.Addr]bool, error) {
	keys, err := d.st.GetKeysByPrefix(rtEdgeTable, "")
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	tks := make(map[ctype.Addr]bool)
	for _, key := range keys {
		// tokenAddr and cid are split by ":" in key
		tokenCid := strings.Split(key, ":")
		if len(tokenCid) != 2 {
			return nil, errors.New("Wrong key " + key)
		}
		tokenAddr := ctype.Hex2Addr(tokenCid[0])
		tks[tokenAddr] = true
	}
	return tks, nil
}
func (d *DAL) GetAllRoutes() (map[ctype.Addr]map[ctype.Addr]ctype.CidType, error) {
	dstsOnTokens, err := d.st.GetKeysByPrefix(routingTable, "")
	if err != nil {
		return nil, err
	}
	routes := make(map[ctype.Addr]map[ctype.Addr]ctype.CidType)
	for _, dstOnToken := range dstsOnTokens {
		dstToken := strings.Split(dstOnToken, "@")
		if len(dstToken) != 2 {
			return nil, errors.New("Wrong route key " + dstOnToken)
		}
		dst := ctype.Hex2Addr(dstToken[0])
		token := ctype.Hex2Addr(dstToken[1])
		cid, getRouteErr := getRoute(d.st, dstOnToken)
		if getRouteErr != nil {
			log.Errorln(getRouteErr)
			continue
		}
		if _, ok := routes[token]; !ok {
			routes[token] = make(map[ctype.Addr]ctype.CidType)
		}
		routes[token][dst] = cid
	}
	return routes, nil
}

type servingOspMap = map[ctype.Addr]bool

// GetAllServingOsps returns (tokenAddr,client)->map[ospaddr]bool
func (d *DAL) GetAllServingOsps() (map[ctype.Addr]map[ctype.Addr]servingOspMap, error) {
	// key is clientAddr@tokenAddr@ospAddr
	clientsOnTokensOnOsps, err := d.st.GetKeysByPrefix(servingOspsTable, "")
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	allServingOsps := make(map[ctype.Addr]map[ctype.Addr]servingOspMap)
	for _, clientOnTokenOnOsp := range clientsOnTokensOnOsps {
		clientTokenOsp := strings.Split(clientOnTokenOnOsp, "@")
		if len(clientTokenOsp) != 3 {
			return nil, errors.New("Wrong serivng osps key " + clientOnTokenOnOsp)
		}
		client := ctype.Hex2Addr(clientTokenOsp[0])
		token := ctype.Hex2Addr(clientTokenOsp[1])
		osp := ctype.Hex2Addr(clientTokenOsp[2])
		if _, ok := allServingOsps[token]; !ok {
			allServingOsps[token] = make(map[ctype.Addr]servingOspMap)
		}
		if _, ok := allServingOsps[token][client]; !ok {
			allServingOsps[token][client] = make(servingOspMap)
		}
		allServingOsps[token][client][osp] = true
	}
	return allServingOsps, nil
}

func (d *DAL) GetServingOsps(clientAddr ctype.Addr, tokenAddr ctype.Addr) (servingOspMap, error) {
	key := strings.Join([]string{ctype.Addr2Hex(clientAddr), ctype.Addr2Hex(tokenAddr)}, "@")
	clientsOnTokenOnOsps, err := d.st.GetKeysByPrefix(servingOspsTable, key)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	osps := make(map[ctype.Addr]bool)
	for _, clientOnTokenOnOsp := range clientsOnTokenOnOsps {
		clientTokenOsp := strings.Split(clientOnTokenOnOsp, "@")
		if len(clientTokenOsp) != 3 {
			log.Errorln("wrong serving osps key", clientOnTokenOnOsp)
			continue
		}
		osp := clientTokenOsp[2]
		osps[ctype.Hex2Addr(osp)] = true
	}
	return osps, nil
}
func (d *DAL) PutServingOsp(clientAddr ctype.Addr, tokenAddr ctype.Addr, ospAddr ctype.Addr) error {
	key := strings.Join([]string{ctype.Addr2Hex(clientAddr), ctype.Addr2Hex(tokenAddr), ctype.Addr2Hex(ospAddr)}, "@")
	return d.st.Put(servingOspsTable, key, true)
}
func (d *DAL) DeleteServingOsp(clientAddr ctype.Addr, tokenAddr ctype.Addr, ospAddr ctype.Addr) error {
	key := strings.Join([]string{ctype.Addr2Hex(clientAddr), ctype.Addr2Hex(tokenAddr), ctype.Addr2Hex(ospAddr)}, "@")
	return d.st.Delete(servingOspsTable, key)
}
func (d *DAL) MarkOsp(osp ctype.Addr) {
	d.st.Put(markedOspTable, ctype.Addr2Hex(osp), true)
}
func (d *DAL) GetAllMarkedOsp() (map[ctype.Addr]bool, error) {
	osps, err := d.st.GetKeysByPrefix(markedOspTable, "")
	if err != nil {
		return nil, err
	}
	ospSet := make(map[ctype.Addr]bool)
	for _, osp := range osps {
		ospSet[ctype.Hex2Addr(osp)] = true
	}
	return ospSet, nil
}
func (d *DAL) UnmarkOsp(osp ctype.Addr) {
	d.st.Delete(markedOspTable, ctype.Addr2Hex(osp))
}

// DAL for peer message queue

func channelMsgKey(cid ctype.CidType, seqnum uint64) string {
	return fmt.Sprintf("%s:%d", ctype.Cid2Hex(cid), seqnum)
}

func getChannelMessage(st Storage, cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, error) {
	var msgBytes []byte
	key := channelMsgKey(cid, seqnum)
	err := st.Get(channelMessageQueue, key, &msgBytes)
	if err != nil {
		return nil, err
	}
	var msg rpc.CelerMsg
	err = proto.Unmarshal(msgBytes, &msg)
	return &msg, err
}

func putChannelMessage(st Storage, cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	key := channelMsgKey(cid, seqnum)
	return st.Put(channelMessageQueue, key, msgBytes)
}

func deleteChannelMessage(st Storage, cid ctype.CidType, seqnum uint64) error {
	key := channelMsgKey(cid, seqnum)
	return st.Delete(channelMessageQueue, key)
}

func hasChannelMessage(st Storage, cid ctype.CidType, seqnum uint64) (bool, error) {
	key := channelMsgKey(cid, seqnum)
	return st.Has(channelMessageQueue, key)
}

func getAllChannelMessageSeqnums(st Storage, cid ctype.CidType) ([]uint64, error) {
	keys, err := st.GetKeysByPrefix(channelMessageQueue, ctype.Cid2Hex(cid)+":")
	if err != nil {
		return nil, err
	}

	ret := make([]uint64, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid peer message key: %s", key)
		}
		seqnum, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid peer message key: %s", key)
		}
		ret = append(ret, seqnum)
	}

	return ret, nil
}

func (d *DAL) GetChannelMessage(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, error) {
	return getChannelMessage(d.st, cid, seqnum)
}

func (d *DAL) PutChannelMessage(cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	return putChannelMessage(d.st, cid, seqnum, msg)
}

func (d *DAL) DeleteChannelMessage(cid ctype.CidType, seqnum uint64) error {
	return deleteChannelMessage(d.st, cid, seqnum)
}

func (d *DAL) HasChannelMessage(cid ctype.CidType, seqnum uint64) (bool, error) {
	return hasChannelMessage(d.st, cid, seqnum)
}

func (d *DAL) GetAllChannelMessageSeqnums(cid ctype.CidType) ([]uint64, error) {
	return getAllChannelMessageSeqnums(d.st, cid)
}

func (dtx *DALTx) GetChannelMessage(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, error) {
	return getChannelMessage(dtx.stx, cid, seqnum)
}

func (dtx *DALTx) PutChannelMessage(cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	return putChannelMessage(dtx.stx, cid, seqnum, msg)
}

func (dtx *DALTx) DeleteChannelMessage(cid ctype.CidType, seqnum uint64) error {
	return deleteChannelMessage(dtx.stx, cid, seqnum)
}

func (dtx *DALTx) HasChannelMessage(cid ctype.CidType, seqnum uint64) (bool, error) {
	return hasChannelMessage(dtx.stx, cid, seqnum)
}

func (dtx *DALTx) GetAllChannelMessageSeqnums(cid ctype.CidType) ([]uint64, error) {
	return getAllChannelMessageSeqnums(dtx.stx, cid)
}

func getChannelSeqNums(st Storage, cid ctype.CidType) (*common.ChannelSeqNums, error) {
	var seqnums common.ChannelSeqNums
	err := st.Get(channelSeqNums, ctype.Cid2Hex(cid), &seqnums)
	return &seqnums, err
}

func putChannelSeqNums(st Storage, cid ctype.CidType, seqnums *common.ChannelSeqNums) error {
	return st.Put(channelSeqNums, ctype.Cid2Hex(cid), seqnums)
}

func deleteChannelSeqNums(st Storage, cid ctype.CidType) error {
	return st.Delete(channelSeqNums, ctype.Cid2Hex(cid))
}

func hasChannelSeqNums(st Storage, cid ctype.CidType) (bool, error) {
	return st.Has(channelSeqNums, ctype.Cid2Hex(cid))
}

func (d *DAL) GetChannelSeqNums(cid ctype.CidType) (*common.ChannelSeqNums, error) {
	return getChannelSeqNums(d.st, cid)
}

func (d *DAL) PutChannelSeqNums(cid ctype.CidType, seqnums *common.ChannelSeqNums) error {
	return putChannelSeqNums(d.st, cid, seqnums)
}

func (d *DAL) DeleteChannelSeqNums(cid ctype.CidType) error {
	return deleteChannelSeqNums(d.st, cid)
}

func (d *DAL) HasChannelSeqNums(cid ctype.CidType) (bool, error) {
	return hasChannelSeqNums(d.st, cid)
}

func (dtx *DALTx) GetChannelSeqNums(cid ctype.CidType) (*common.ChannelSeqNums, error) {
	return getChannelSeqNums(dtx.stx, cid)
}

func (dtx *DALTx) PutChannelSeqNums(cid ctype.CidType, seqnums *common.ChannelSeqNums) error {
	return putChannelSeqNums(dtx.stx, cid, seqnums)
}

func (dtx *DALTx) DeleteChannelSeqNums(cid ctype.CidType) error {
	return deleteChannelSeqNums(dtx.stx, cid)
}

func (dtx *DALTx) HasChannelSeqNums(cid ctype.CidType) (bool, error) {
	return hasChannelSeqNums(dtx.stx, cid)
}

func getPeerActiveChannels(st Storage, peer string) (map[ctype.CidType]bool, error) {
	var cids map[ctype.CidType]bool
	err := st.Get(peerActiveChannels, peer, &cids)
	return cids, err
}

func putPeerActiveChannels(st Storage, peer string, cids map[ctype.CidType]bool) error {
	return st.Put(peerActiveChannels, peer, cids)
}

func deletePeerActiveChannels(st Storage, peer string) error {
	return st.Delete(peerActiveChannels, peer)
}

func hasPeerActiveChannels(st Storage, peer string) (bool, error) {
	return st.Has(peerActiveChannels, peer)
}

func (d *DAL) GetPeerActiveChannels(peer string) (map[ctype.CidType]bool, error) {
	return getPeerActiveChannels(d.st, peer)
}

func (d *DAL) PutPeerActiveChannels(peer string, cids map[ctype.CidType]bool) error {
	return putPeerActiveChannels(d.st, peer, cids)
}

func (d *DAL) DeletePeerActiveChannels(peer string) error {
	return deletePeerActiveChannels(d.st, peer)
}

func (d *DAL) HasPeerActiveChannels(peer string) (bool, error) {
	return hasPeerActiveChannels(d.st, peer)
}

func (dtx *DALTx) GetPeerActiveChannels(peer string) (map[ctype.CidType]bool, error) {
	return getPeerActiveChannels(dtx.stx, peer)
}

func (dtx *DALTx) PutPeerActiveChannels(peer string, cids map[ctype.CidType]bool) error {
	return putPeerActiveChannels(dtx.stx, peer, cids)
}

func (dtx *DALTx) DeletePeerActiveChannels(peer string) error {
	return deletePeerActiveChannels(dtx.stx, peer)
}

func (dtx *DALTx) HasPeerActiveChannels(peer string) (bool, error) {
	return hasPeerActiveChannels(dtx.stx, peer)
}
