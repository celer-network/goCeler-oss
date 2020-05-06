// Copyright 2018-2020 Celer Network
//
// This is the Data Access Layer. It maps the server's data structures
// that need to be persisted to KVStore calls:
// * Construct table keys from object attribute(s).
// * Use the appropriate Go data types when fetching stored values.

package storage

import (
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/cnode/openchannelts"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	// OSP only. to avoid handling concurrent openchan requests
	openChannelTs = "oct" // peer@token -> OpenChannelTs

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

// ====================== DAL APIs for SQL schema ======================

// The "channels" table.
func (d *DAL) InsertChanWithTs(cid ctype.CidType, peer ctype.Addr, token *entity.TokenInfo, ledger ctype.Addr, state int, stateTs, openTs time.Time, openResp *rpc.OpenChannelResponse, onchainBalance *structs.OnChainBalance, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64, selfSimplex *rpc.SignedSimplexState, peerSimplex *rpc.SignedSimplexState) error {
	return insertChanWithTs(d.st, cid, peer, token, ledger, state, stateTs, openTs, openResp, onchainBalance, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum, selfSimplex, peerSimplex)
}

func (d *DAL) InsertChan(cid ctype.CidType, peer ctype.Addr, token *entity.TokenInfo, ledger ctype.Addr, state int, openResp *rpc.OpenChannelResponse, onchainBalance *structs.OnChainBalance, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64, selfSimplex *rpc.SignedSimplexState, peerSimplex *rpc.SignedSimplexState) error {
	return insertChan(d.st, cid, peer, token, ledger, state, openResp, onchainBalance, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum, selfSimplex, peerSimplex)
}

func (d *DAL) DeleteChan(cid ctype.CidType) error {
	return deleteChan(d.st, cid)
}

func (d *DAL) GetChanViewInfoByID(cid ctype.CidType) (int, *time.Time, *time.Time, *entity.PaymentChannelInitializer, *structs.OnChainBalance, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	return getChanViewInfoByID(d.st, cid)
}

func (d *DAL) GetAllChanInfoByToken(token *entity.TokenInfo) ([]ctype.CidType, []ctype.Addr, []*entity.TokenInfo, []int, []*time.Time, []*time.Time, []*structs.OnChainBalance, []*entity.SimplexPaymentChannel, []*entity.SimplexPaymentChannel, error) {
	return getAllChanInfoByToken(d.st, token)
}

func (d *DAL) GetInactiveChanInfo(token *entity.TokenInfo, stateTs time.Time) ([]ctype.CidType, []ctype.Addr, []*entity.TokenInfo, []int, []*time.Time, []*time.Time, []*structs.OnChainBalance, []*entity.SimplexPaymentChannel, []*entity.SimplexPaymentChannel, error) {
	return getInactiveChanInfo(d.st, token, stateTs)
}

func (d *DAL) GetChanState(cid ctype.CidType) (int, bool, error) {
	return getChanState(d.st, cid)
}

func (d *DAL) UpdateChanState(cid ctype.CidType, state int) error {
	return updateChanState(d.st, cid, state)
}

func (d *DAL) GetChanOpenResp(cid ctype.CidType) (*rpc.OpenChannelResponse, bool, error) {
	return getChanOpenResp(d.st, cid)
}

func (d *DAL) UpdateChanOpenResp(cid ctype.CidType, openResp *rpc.OpenChannelResponse) error {
	return updateChanOpenResp(d.st, cid, openResp)
}

func (d *DAL) GetChanPeer(cid ctype.CidType) (ctype.Addr, bool, error) {
	return getChanPeer(d.st, cid)
}

func (d *DAL) GetChanStateToken(cid ctype.CidType) (int, *entity.TokenInfo, bool, error) {
	return getChanStateToken(d.st, cid)
}

func (d *DAL) GetChanForDeposit(cid ctype.CidType) (int, *entity.TokenInfo, ctype.Addr, ctype.Addr, bool, error) {
	return getChanForDeposit(d.st, cid)
}

func (d *DAL) GetCidByPeerToken(peer ctype.Addr, token *entity.TokenInfo) (ctype.CidType, bool, error) {
	return getCidByPeerToken(d.st, peer, token)
}

func (d *DAL) GetCidStateByPeerToken(peer ctype.Addr, token *entity.TokenInfo) (ctype.CidType, int, bool, error) {
	return getCidStateByPeerToken(d.st, peer, token)
}

func (d *DAL) GetSelfSimplex(cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	return getSelfSimplex(d.st, cid)
}

func (d *DAL) GetPeerSimplex(cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	return getPeerSimplex(d.st, cid)
}

func (d *DAL) GetDuplexChannel(cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, *entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	return getDuplexChannel(d.st, cid)
}

func (d *DAL) GetChanSeqNums(cid ctype.CidType) (uint64, uint64, uint64, uint64, bool, error) {
	return getChanSeqNums(d.st, cid)
}

func (d *DAL) UpdateChanSeqNums(cid ctype.CidType, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64) error {
	return updateChanSeqNums(d.st, cid, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum)
}

func (d *DAL) GetOnChainBalance(cid ctype.CidType) (*structs.OnChainBalance, bool, error) {
	return getOnChainBalance(d.st, cid)
}

func (d *DAL) UpdateOnChainBalance(cid ctype.CidType, balance *structs.OnChainBalance) error {
	return updateOnChainBalance(d.st, cid, balance)
}

func (d *DAL) GetChannelsForAuthReq(peerAddr ctype.Addr) ([]*rpc.ChannelSummary, error) {
	return getChannelsForAuthReq(d.st, peerAddr)
}

func (d *DAL) GetChannelsForAuthAck(peerAddr ctype.Addr) ([]*chanForAuthAck, error) {
	return getChannelsForAuthAck(d.st, peerAddr)
}

func (d *DAL) GetCidTokensByPeer(peerAddr ctype.Addr) ([]ctype.CidType, []ctype.Addr, error) {
	return getCidTokensByPeer(d.st, peerAddr)
}

func (d *DAL) GetChanLedger(cid ctype.CidType) (ctype.Addr, bool, error) {
	return getChanLedger(d.st, cid)
}

func (d *DAL) GetCidsByTokenAndState(token *entity.TokenInfo, state int) ([]ctype.CidType, error) {
	return getCidsByTokenAndState(d.st, token, state)
}

func (d *DAL) CountCidsByTokenAndState(token *entity.TokenInfo, state int) (int, error) {
	return countCidsByTokenAndState(d.st, token, state)
}

func (d *DAL) GetInactiveCidsByTokenAndState(token *entity.TokenInfo, state int, stateTs time.Time) ([]ctype.CidType, error) {
	return getInactiveCidsByTokenAndState(d.st, token, state, stateTs)
}

func (d *DAL) CountInactiveCidsByTokenAndState(token *entity.TokenInfo, state int, stateTs time.Time) (int, error) {
	return countInactiveCidsByTokenAndState(d.st, token, state, stateTs)
}

func (d *DAL) UpdateChanLedger(cid ctype.CidType, ledger ctype.Addr) error {
	return updateChanLedger(d.st, cid, ledger)
}

func (d *DAL) GetChanForMigration(cid ctype.CidType) (int, ctype.Addr, bool, error) {
	return getChanForMigration(d.st, cid)
}

func (d *DAL) GetAllChanLedgers() ([]ctype.Addr, error) {
	return getAllChanLedgers(d.st)
}

func (dtx *DALTx) InsertChan(cid ctype.CidType, peer ctype.Addr, token *entity.TokenInfo, ledger ctype.Addr, state int, openResp *rpc.OpenChannelResponse, onchainBalance *structs.OnChainBalance, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64, selfSimplex *rpc.SignedSimplexState, peerSimplex *rpc.SignedSimplexState) error {
	return insertChan(dtx.stx, cid, peer, token, ledger, state, openResp, onchainBalance, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum, selfSimplex, peerSimplex)
}

func (dtx *DALTx) DeleteChan(cid ctype.CidType) error {
	return deleteChan(dtx.stx, cid)
}

func (dtx *DALTx) GetChanState(cid ctype.CidType) (int, bool, error) {
	return getChanState(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanState(cid ctype.CidType, state int) error {
	return updateChanState(dtx.stx, cid, state)
}

func (dtx *DALTx) GetChanPeer(cid ctype.CidType) (ctype.Addr, bool, error) {
	return getChanPeer(dtx.stx, cid)
}

func (dtx *DALTx) GetChanPeerState(cid ctype.CidType) (ctype.Addr, int, bool, error) {
	return getChanPeerState(dtx.stx, cid)
}

func (dtx *DALTx) GetChanForClose(cid ctype.CidType) (ctype.Addr, *entity.TokenInfo, time.Time, bool, error) {
	return getChanForClose(dtx.stx, cid)
}

func (dtx *DALTx) GetChanSeqNums(cid ctype.CidType) (uint64, uint64, uint64, uint64, bool, error) {
	return getChanSeqNums(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanSeqNums(cid ctype.CidType, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64) error {
	return updateChanSeqNums(dtx.stx, cid, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum)
}

func (dtx *DALTx) GetChanForBalance(cid ctype.CidType) (ctype.Addr, *structs.OnChainBalance, uint64, uint64, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	return getChanForBalance(dtx.stx, cid)
}

func (dtx *DALTx) GetOnChainBalance(cid ctype.CidType) (*structs.OnChainBalance, bool, error) {
	return getOnChainBalance(dtx.stx, cid)
}

func (dtx *DALTx) UpdateOnChainBalance(cid ctype.CidType, balance *structs.OnChainBalance) error {
	return updateOnChainBalance(dtx.stx, cid, balance)
}

func (dtx *DALTx) GetChanForSendCondPayRequest(cid ctype.CidType) (ctype.Addr, int, *structs.OnChainBalance, uint64, uint64, uint64, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	return getChanForSendCondPayRequest(dtx.stx, cid)
}

func (dtx *DALTx) GetChanForSendPaySettleRequest(cid ctype.CidType) (ctype.Addr, int, uint64, uint64, uint64, *entity.SimplexPaymentChannel, bool, error) {
	return getChanForSendPaySettleRequest(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanForSendRequest(cid ctype.CidType, baseSeqNum uint64, lastUsedSeqNum uint64) error {
	return updateChanForSendRequest(dtx.stx, cid, baseSeqNum, lastUsedSeqNum)
}

func (dtx *DALTx) GetChanForRecvPayRequest(cid ctype.CidType) (ctype.Addr, int, *structs.OnChainBalance, uint64, uint64, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	return getChanForRecvPayRequest(dtx.stx, cid)
}

func (dtx *DALTx) GetChanStateAndPeerSimplex(cid ctype.CidType) (int, *entity.SimplexPaymentChannel, bool, error) {
	return getChanStateAndPeerSimplex(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanForRecvRequest(cid ctype.CidType, peerSimplex *rpc.SignedSimplexState) error {
	return updateChanForRecvRequest(dtx.stx, cid, peerSimplex)
}

func (dtx *DALTx) GetChanForRecvResponse(cid ctype.CidType) (ctype.Addr, int, uint64, uint64, uint64, uint64, bool, error) {
	return getChanForRecvResponse(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanForRecvResponse(cid ctype.CidType, baseSeqNum uint64, lastUsedSeqNum uint64, lastAckedSeqNum uint64, lastNackedSeqNum uint64, selfSimplex *rpc.SignedSimplexState) error {
	return updateChanForRecvResponse(dtx.stx, cid, baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum, selfSimplex)
}

func (dtx *DALTx) GetChanForIntendSettle(cid ctype.CidType) (ctype.Addr, int, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	return getChanForIntendSettle(dtx.stx, cid)
}

func (dtx *DALTx) GetPeerSimplex(cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	return getPeerSimplex(dtx.stx, cid)
}

func (dtx *DALTx) GetChanLedger(cid ctype.CidType) (ctype.Addr, bool, error) {
	return getChanLedger(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanLedger(cid ctype.CidType, ledger ctype.Addr) error {
	return updateChanLedger(dtx.stx, cid, ledger)
}

func (dtx *DALTx) GetChanForMigration(cid ctype.CidType) (int, ctype.Addr, bool, error) {
	return getChanForMigration(dtx.stx, cid)
}

// The "messages" table.
func (d *DAL) InsertChanMessage(cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	return insertChanMessage(d.st, cid, seqnum, msg)
}

func (d *DAL) GetChanMessage(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, bool, error) {
	return getChanMessage(d.st, cid, seqnum)
}

func (d *DAL) GetAllChanMessages(cid ctype.CidType) ([]*rpc.CelerMsg, error) {
	return getAllChanMessages(d.st, cid)
}

func (d *DAL) DeleteChanMessage(cid ctype.CidType, seqnum uint64) error {
	return deleteChanMessage(d.st, cid, seqnum)
}

func (dtx *DALTx) InsertChanMessage(cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	return insertChanMessage(dtx.stx, cid, seqnum, msg)
}

func (dtx *DALTx) GetChanMessage(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, bool, error) {
	return getChanMessage(dtx.stx, cid, seqnum)
}

func (dtx *DALTx) DeleteChanMessage(cid ctype.CidType, seqnum uint64) error {
	return deleteChanMessage(dtx.stx, cid, seqnum)
}

// The "closedchannels" table.
func (d *DAL) InsertClosedChan(cid ctype.CidType, peer ctype.Addr, token *entity.TokenInfo, openTs time.Time, closeTs time.Time) error {
	return insertClosedChan(d.st, cid, peer, token, openTs, closeTs)
}

func (dtx *DALTx) InsertClosedChan(cid ctype.CidType, peer ctype.Addr, token *entity.TokenInfo, openTs time.Time, closeTs time.Time) error {
	return insertClosedChan(dtx.stx, cid, peer, token, openTs, closeTs)
}

func (d *DAL) GetClosedChan(cid ctype.CidType) (ctype.Addr, *entity.TokenInfo, *time.Time, *time.Time, bool, error) {
	return getClosedChan(d.st, cid)
}

// The "payments" table.
func (d *DAL) InsertPaymentWithTs(payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any, inCid ctype.CidType, inState int, outCid ctype.CidType, outState int, createTs time.Time) error {
	return insertPaymentWithTs(d.st, payID, payBytes, pay, note, inCid, inState, outCid, outState, createTs)
}

func (d *DAL) InsertPayment(payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any, inCid ctype.CidType, inState int, outCid ctype.CidType, outState int) error {
	return insertPayment(d.st, payID, payBytes, pay, note, inCid, inState, outCid, outState)
}

func (d *DAL) DeletePayment(payID ctype.PayIDType) error {
	return deletePayment(d.st, payID)
}

func (d *DAL) GetAllPayIDs() ([]ctype.PayIDType, error) {
	return getAllPayIDs(d.st)
}

func (d *DAL) GetPayment(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, bool, error) {
	return getPayment(d.st, payID)
}

func (d *DAL) GetPaymentInfo(payID ctype.PayIDType) (*entity.ConditionalPay, *any.Any, ctype.CidType, int, ctype.CidType, int, *time.Time, bool, error) {
	return getPaymentInfo(d.st, payID)
}

func (d *DAL) GetAllPaymentInfoByCid(cid ctype.CidType) ([]ctype.PayIDType, []*entity.ConditionalPay, []*any.Any, []ctype.CidType, []int, []ctype.CidType, []int, []*time.Time, error) {
	return getAllPaymentInfoByCid(d.st, cid)
}

func (d *DAL) GetPayStates(payID ctype.PayIDType) (int, int, bool, error) {
	return getPayStates(d.st, payID)
}

func (d *DAL) GetPayIngress(payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	return getPayIngress(d.st, payID)
}

func (d *DAL) GetPayIngressChannel(payID ctype.PayIDType) (ctype.CidType, bool, error) {
	return getPayIngressChannel(d.st, payID)
}

func (d *DAL) GetPayEgress(payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	return getPayEgress(d.st, payID)
}

func (d *DAL) GetPayEgressChannel(payID ctype.PayIDType) (ctype.CidType, bool, error) {
	return getPayEgressChannel(d.st, payID)
}

func (d *DAL) GetPayNote(payID ctype.PayIDType) (*any.Any, bool, error) {
	return getPayNote(d.st, payID)
}

// GetPayAndEgressState returns (pay, pay_bytes, egress_state, found, error)
func (d *DAL) GetPayAndEgressState(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, int, bool, error) {
	return getPayAndEgressState(d.st, payID)
}

// GetPayHistory returns payIDs, pay entities, pay in_states and creation time where peer is either src or dest of the pay.
// Pays are sorted in reverse-chronological order.
// smallestPayID: if query result contains multiple entries with creation timestamp equals to beforeTs, the pays returned from this func
// will only include ones that has higher payID than smallestPayID.
// Same dimensions of the four guaranteed.
func (d *DAL) GetPayHistory(
	peer ctype.Addr, beforeTs time.Time, smallestPayID ctype.PayIDType, maxResultSize int32) ([]ctype.PayIDType, []*entity.ConditionalPay, []int64, []int64, error) {
	return getPayHistory(d.st, peer, beforeTs, smallestPayID, maxResultSize)
}

func (d *DAL) GetPaysForAuthAck(payIDs []ctype.PayIDType, isOut bool) ([]*rpc.PayInAuthAck, error) {
	return getPaysForAuthAck(d.st, payIDs, isOut)
}

func (d *DAL) CountPayments() (int, error) {
	return countPayments(d.st)
}

func (dtx *DALTx) InsertPayment(payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any, inCid ctype.CidType, inState int, outCid ctype.CidType, outState int) error {
	return insertPayment(dtx.stx, payID, payBytes, pay, note, inCid, inState, outCid, outState)
}

func (dtx *DALTx) DeletePayment(payID ctype.PayIDType) error {
	return deletePayment(dtx.stx, payID)
}

func (dtx *DALTx) GetPayment(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, bool, error) {
	return getPayment(dtx.stx, payID)
}

func (dtx *DALTx) GetPayIngress(payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	return getPayIngress(dtx.stx, payID)
}

func (dtx *DALTx) GetPayIngressChannel(payID ctype.PayIDType) (ctype.CidType, bool, error) {
	return getPayIngressChannel(dtx.stx, payID)
}

func (dtx *DALTx) UpdatePayIngressState(payID ctype.PayIDType, state int) error {
	return updatePayIngressState(dtx.stx, payID, state)
}

func (dtx *DALTx) GetPayEgress(payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	return getPayEgress(dtx.stx, payID)
}

func (dtx *DALTx) UpdatePayEgress(payID ctype.PayIDType, cid ctype.CidType, state int) error {
	return updatePayEgress(dtx.stx, payID, cid, state)
}

func (dtx *DALTx) UpdatePayEgressState(payID ctype.PayIDType, state int) error {
	return updatePayEgressState(dtx.stx, payID, state)
}

func (dtx *DALTx) GetPayNote(payID ctype.PayIDType) (*any.Any, bool, error) {
	return getPayNote(dtx.stx, payID)
}

func (dtx *DALTx) GetPayForRecvSettle(payID ctype.PayIDType) (*entity.ConditionalPay, *any.Any, ctype.CidType, int, int, bool, error) {
	return getPayForRecvSettle(dtx.stx, payID)
}

func (dtx *DALTx) GetPayForRecvSecret(payID ctype.PayIDType) (*entity.ConditionalPay, *any.Any, int, bool, error) {
	return getPayForRecvSecret(dtx.stx, payID)
}

func (dtx *DALTx) GetPayAndEgressState(payID ctype.PayIDType) (*entity.ConditionalPay, []byte, int, bool, error) {
	return getPayAndEgressState(dtx.stx, payID)
}

// The "paydelegation" table.
func (d *DAL) InsertDelegatedPay(payID ctype.PayIDType, dest ctype.Addr, status int) error {
	return insertDelegatedPay(d.st, payID, dest, status)
}

func (d *DAL) DeleteDelegatedPay(payID ctype.PayIDType) error {
	return deleteDelegatedPay(d.st, payID)
}

func (d *DAL) GetDelegatedPayStatus(payID ctype.PayIDType) (int, bool, error) {
	return getDelegatedPayStatus(d.st, payID)
}

func (d *DAL) UpdateDelegatedPayStatus(payID ctype.PayIDType, status int) error {
	return updateDelegatedPayStatus(d.st, payID, status)
}

func (d *DAL) GetDelegatedPaysOnStatus(dest ctype.Addr, status int) (map[ctype.PayIDType]*entity.ConditionalPay, error) {
	return getDelegatedPaysOnStatus(d.st, dest, status)
}

func (dtx *DALTx) InsertDelegatedPay(payID ctype.PayIDType, dest ctype.Addr, status int) error {
	return insertDelegatedPay(dtx.stx, payID, dest, status)
}

func (dtx *DALTx) DeleteDelegatedPay(payID ctype.PayIDType) error {
	return deleteDelegatedPay(dtx.stx, payID)
}

func (dtx *DALTx) GetDelegatedPayStatus(payID ctype.PayIDType) (int, bool, error) {
	return getDelegatedPayStatus(dtx.stx, payID)
}

func (dtx *DALTx) UpdateDelegatedPayStatus(payID ctype.PayIDType, status int) error {
	return updateDelegatedPayStatus(dtx.stx, payID, status)
}

func (dtx *DALTx) UpdateSendingDelegatedPay(payID, payIDout ctype.PayIDType) error {
	return updateSendingDelegatedPay(dtx.stx, payID, payIDout)
}

func (d *DAL) InsertPayDelegator(payID ctype.PayIDType, dest, delegator ctype.Addr) error {
	return insertPayDelegator(d.st, payID, dest, delegator)
}

func (d *DAL) GetPayDelegator(payID ctype.PayIDType) (ctype.Addr, bool, error) {
	return getPayDelegator(d.st, payID)
}

func (dtx *DALTx) GetPayDelegator(payID ctype.PayIDType) (ctype.Addr, bool, error) {
	return getPayDelegator(dtx.stx, payID)
}

// The "secrets" table.
func (d *DAL) InsertSecret(hash, preImage string, payID ctype.PayIDType) error {
	return insertSecret(d.st, hash, preImage, payID)
}

func (d *DAL) GetSecret(hash string) (string, bool, error) {
	return getSecret(d.st, hash)
}

func (d *DAL) DeleteSecret(hash string) error {
	return deleteSecret(d.st, hash)
}

func (d *DAL) DeleteSecretByPayID(payID ctype.PayIDType) error {
	return deleteSecretByPayID(d.st, payID)
}

func (dtx *DALTx) InsertSecret(hash, preImage string, payID ctype.PayIDType) error {
	return insertSecret(dtx.stx, hash, preImage, payID)
}

func (dtx *DALTx) GetSecret(hash string) (string, bool, error) {
	return getSecret(dtx.stx, hash)
}

func (dtx *DALTx) DeleteSecretByPayID(payID ctype.PayIDType) error {
	return deleteSecretByPayID(dtx.stx, payID)
}

// The "tcb" table.
func (d *DAL) InsertTcb(addr ctype.Addr, token *entity.TokenInfo, deposit *big.Int) error {
	return insertTcb(d.st, addr, token, deposit)
}

func (d *DAL) GetTcbDeposit(addr ctype.Addr, token *entity.TokenInfo) (*big.Int, bool, error) {
	return getTcbDeposit(d.st, addr, token)
}

func (d *DAL) UpdateTcbDeposit(addr ctype.Addr, token *entity.TokenInfo, deposit *big.Int) error {
	return updateTcbDeposit(d.st, addr, token, deposit)
}

func (dtx *DALTx) InsertTcb(addr ctype.Addr, token *entity.TokenInfo, deposit *big.Int) error {
	return insertTcb(dtx.stx, addr, token, deposit)
}

func (dtx *DALTx) GetTcbDeposit(addr ctype.Addr, token *entity.TokenInfo) (*big.Int, bool, error) {
	return getTcbDeposit(dtx.stx, addr, token)
}

func (dtx *DALTx) UpdateTcbDeposit(addr ctype.Addr, token *entity.TokenInfo, deposit *big.Int) error {
	return updateTcbDeposit(dtx.stx, addr, token, deposit)
}

// The "monitor" table.
func (d *DAL) InsertMonitor(event string, blockNum uint64, blockIdx int64, restart bool) error {
	return insertMonitor(d.st, event, blockNum, blockIdx, restart)
}

func (d *DAL) GetMonitorBlock(event string) (uint64, int64, bool, error) {
	return getMonitorBlock(d.st, event)
}

func (d *DAL) GetMonitorRestart(event string) (bool, bool, error) {
	return getMonitorRestart(d.st, event)
}

func (d *DAL) GetMonitorAddrsByEventAndRestart(eventName string, restart bool) ([]ctype.Addr, error) {
	return getMonitorAddrsByEventAndRestart(d.st, eventName, restart)
}

func (d *DAL) UpdateMonitorBlock(event string, blockNum uint64, blockIdx int64) error {
	return updateMonitorBlock(d.st, event, blockNum, blockIdx)
}

func (d *DAL) UpsertMonitorBlock(event string, blockNum uint64, blockIdx int64, restart bool) error {
	return upsertMonitorBlock(d.st, event, blockNum, blockIdx, restart)
}

func (d *DAL) UpsertMonitorRestart(event string, restart bool) error {
	return upsertMonitorRestart(d.st, event, restart)
}

func (dtx *DALTx) InsertMonitor(event string, blockNum uint64, blockIdx int64, restart bool) error {
	return insertMonitor(dtx.stx, event, blockNum, blockIdx, restart)
}

func (dtx *DALTx) GetMonitorBlock(event string) (uint64, int64, bool, error) {
	return getMonitorBlock(dtx.stx, event)
}

func (dtx *DALTx) GetMonitorRestart(event string) (bool, bool, error) {
	return getMonitorRestart(dtx.stx, event)
}

func (dtx *DALTx) UpdateMonitorBlock(event string, blockNum uint64, blockIdx int64) error {
	return updateMonitorBlock(dtx.stx, event, blockNum, blockIdx)
}

func (dtx *DALTx) UpsertMonitorBlock(event string, blockNum uint64, blockIdx int64, restart bool) error {
	return upsertMonitorBlock(dtx.stx, event, blockNum, blockIdx, restart)
}

func (dtx *DALTx) UpsertMonitorRestart(event string, restart bool) error {
	return upsertMonitorRestart(dtx.stx, event, restart)
}

// The "routing" table.
func (d *DAL) UpsertRouting(dest ctype.Addr, token *entity.TokenInfo, cid ctype.CidType) error {
	return upsertRouting(d.st, dest, token, cid)
}

func (d *DAL) GetRoutingCid(dest ctype.Addr, token *entity.TokenInfo) (ctype.CidType, bool, error) {
	return getRoutingCid(d.st, dest, token)
}

func (d *DAL) GetAllRoutingCids() (map[ctype.Addr]map[ctype.Addr]ctype.CidType, error) {
	return getAllRoutingCids(d.st)
}

func (d *DAL) DeleteRouting(dest ctype.Addr, token *entity.TokenInfo) error {
	return deleteRouting(d.st, dest, token)
}

func (dtx *DALTx) DeleteRouting(dest ctype.Addr, token *entity.TokenInfo) error {
	return deleteRouting(dtx.stx, dest, token)
}

// The "edges" table.
func (d *DAL) InsertEdge(token *entity.TokenInfo, cid ctype.CidType, addr1, addr2 ctype.Addr) error {
	return insertEdge(d.st, token, cid, addr1, addr2)
}

func (d *DAL) GetAllEdges() ([]*structs.Edge, error) {
	return getAllEdges(d.st)
}

func (d *DAL) DeleteEdge(cid ctype.CidType) error {
	return deleteEdge(d.st, cid)
}

// The "peers" table.
func (d *DAL) InsertPeer(peer ctype.Addr, server string, cids []ctype.CidType) error {
	return insertPeer(d.st, peer, server, cids)
}

func (d *DAL) UpsertPeerServer(peer ctype.Addr, server string) error {
	return upsertPeerServer(d.st, peer, server)
}

func (d *DAL) UpdatePeerCids(peer ctype.Addr, cids []ctype.CidType) error {
	return updatePeerCids(d.st, peer, cids)
}

func (d *DAL) UpdatePeerCid(peer ctype.Addr, cid ctype.CidType, add bool) error {
	return updatePeerCid(d.st, peer, cid, add)
}

func (d *DAL) UpdatePeerDelegateProof(peer ctype.Addr, proof *rpc.DelegationProof) error {
	return updatePeerDelegateProof(d.st, peer, proof)
}

func (d *DAL) GetPeerServer(peer ctype.Addr) (string, bool, error) {
	return getPeerServer(d.st, peer)
}

func (d *DAL) GetAllPeerServers() ([]string, error) {
	return getAllPeerServers(d.st)
}

func (d *DAL) GetPeerCids(peer ctype.Addr) ([]ctype.CidType, bool, error) {
	return getPeerCids(d.st, peer)
}

func (d *DAL) GetPeerDelegateProof(peer ctype.Addr) (*rpc.DelegationProof, bool, error) {
	return getPeerDelegateProof(d.st, peer)
}

func (dtx *DALTx) InsertPeer(peer ctype.Addr, server string, cids []ctype.CidType) error {
	return insertPeer(dtx.stx, peer, server, cids)
}

func (dtx *DALTx) UpdatePeerCids(peer ctype.Addr, cids []ctype.CidType) error {
	return updatePeerCids(dtx.stx, peer, cids)
}

func (dtx *DALTx) UpdatePeerCid(peer ctype.Addr, cid ctype.CidType, add bool) error {
	return updatePeerCid(dtx.stx, peer, cid, add)
}

func (dtx *DALTx) GetPeerServer(peer ctype.Addr) (string, bool, error) {
	return getPeerServer(dtx.stx, peer)
}

func (dtx *DALTx) GetPeerCids(peer ctype.Addr) ([]ctype.CidType, bool, error) {
	return getPeerCids(dtx.stx, peer)
}

// The "desttokens" table.
func (d *DAL) InsertDestToken(dest ctype.Addr, token *entity.TokenInfo, osps []ctype.Addr, chanBlockNum uint64) error {
	return insertDestToken(d.st, dest, token, osps, chanBlockNum)
}

func (d *DAL) GetDestTokenOpenChanBlkNum(dest ctype.Addr, token *entity.TokenInfo) (uint64, bool, error) {
	return getDestTokenOpenChanBlkNum(d.st, dest, token)
}

func (d *DAL) UpsertDestTokenOpenChanBlkNum(dest ctype.Addr, token *entity.TokenInfo, chanBlockNum uint64) error {
	return upsertDestTokenOpenChanBlkNum(d.st, dest, token, chanBlockNum)
}

func (d *DAL) UpdateDestTokenOsps(dest ctype.Addr, token *entity.TokenInfo, osps []ctype.Addr) error {
	return updateDestTokenOsps(d.st, dest, token, osps)
}

func (d *DAL) DeleteDestToken(dest ctype.Addr, token *entity.TokenInfo) error {
	return deleteDestToken(d.st, dest, token)
}

func (d *DAL) GetDestTokenOsps(dest ctype.Addr, token *entity.TokenInfo) ([]ctype.Addr, error) {
	return getDestTokenOsps(d.st, dest, token)
}

func (d *DAL) GetAllDestTokenOsps() (map[ctype.Addr]map[ctype.Addr]map[ctype.Addr]bool, error) {
	return getAllDestTokenOsps(d.st)
}

// The "migration" table
func (d *DAL) UpsertChanMigration(cid ctype.CidType, toLedger ctype.Addr, deadline uint64, state int, onchainReq *chain.ChannelMigrationRequest) error {
	return upsertChanMigration(d.st, cid, toLedger, deadline, state, onchainReq)
}

func (d *DAL) GetChanMigration(cid ctype.CidType, toLedger ctype.Addr) (uint64, int, []byte, bool, error) {
	return getChanMigration(d.st, cid, toLedger)
}

func (d *DAL) DeleteChanMigration(cid ctype.CidType) error {
	return deleteChanMigration(d.st, cid)
}

func (d *DAL) UpdateChanMigrationState(cid ctype.CidType, toLedger ctype.Addr, state int) error {
	return updateChanMigrationState(d.st, cid, toLedger, state)
}

func (d *DAL) GetChanMigrationReqByLedgerAndStateWithLimit(toLedger ctype.Addr, state, limit int) (map[ctype.CidType][]byte, map[ctype.CidType]uint64, error) {
	return getChanMigrationReqByLedgerAndStateWithLimit(d.st, toLedger, state, limit)
}

func (dtx *DALTx) UpsertChanMigration(cid ctype.CidType, toLedger ctype.Addr, deadline uint64, state int, onchainReq *chain.ChannelMigrationRequest) error {
	return upsertChanMigration(dtx.stx, cid, toLedger, deadline, state, onchainReq)
}

func (dtx *DALTx) GetChanMigration(cid ctype.CidType, toLedger ctype.Addr) (uint64, int, []byte, bool, error) {
	return getChanMigration(dtx.stx, cid, toLedger)
}

func (dtx *DALTx) DeleteChanMigration(cid ctype.CidType) error {
	return deleteChanMigration(dtx.stx, cid)
}

func (dtx *DALTx) UpdateChanMigrationState(cid ctype.CidType, toLedger ctype.Addr, state int) error {
	return updateChanMigrationState(dtx.stx, cid, toLedger, state)
}

// The "deposit" table
func (d *DAL) InsertDeposit(uuid string, cid ctype.CidType, topeer bool, amount *big.Int, refill bool, deadline time.Time, state int, txhash string, errmsg string) error {
	return insertDeposit(d.st, uuid, cid, topeer, amount, refill, deadline, state, txhash, errmsg)
}

func (d *DAL) GetDeposit(uuid string) (ctype.CidType, bool, *big.Int, bool, time.Time, int, string, string, bool, error) {
	return getDeposit(d.st, uuid)
}

func (d *DAL) GetDepositState(uuid string) (int, string, bool, error) {
	return getDepositState(d.st, uuid)
}

func (d *DAL) GetDepositByTxHash(txhash string) (string, ctype.CidType, bool, *big.Int, bool, time.Time, int, string, bool, error) {
	return getDepositByTxHash(d.st, txhash)
}

func (d *DAL) GetDepositJob(uuid string) (*structs.DepositJob, bool, error) {
	return getDepositJob(d.st, uuid)
}

func (d *DAL) GetDepositJobByTxHash(txhash string) (*structs.DepositJob, bool, error) {
	return getDepositJobByTxHash(d.st, txhash)
}

func (d *DAL) GetAllDepositJobsByState(state int) ([]*structs.DepositJob, error) {
	return getAllDepositJobsByState(d.st, state)
}

func (d *DAL) GetAllDepositJobsByCid(cid ctype.CidType) ([]*structs.DepositJob, error) {
	return getAllDepositJobsByCid(d.st, cid)
}

func (d *DAL) GetAllRunningDepositJobs() ([]*structs.DepositJob, error) {
	return getAllRunningDepositJobs(d.st)
}

func (d *DAL) GetAllSubmittedDepositTxHashes() ([]string, error) {
	return getAllSubmittedDepositTxHashes(d.st)
}

func (d *DAL) UpdateDepositStateAndTxHash(uuid string, state int, txhash string) error {
	return updateDepositStateAndTxHash(d.st, uuid, state, txhash)
}

func (d *DAL) UpdateDepositsStateAndTxHash(uuids []string, state int, txhash string) error {
	return updateDepositsStateAndTxHash(d.st, uuids, state, txhash)
}

func (d *DAL) UpdateDepositErrMsg(uuid, errmsg string) error {
	return updateDepositErrMsg(d.st, uuid, errmsg)
}

func (d *DAL) UpdateDepositsErrMsg(uuids []string, errmsg string) error {
	return updateDepositsErrMsg(d.st, uuids, errmsg)
}

func (d *DAL) UpdateDepositErrMsgByTxHash(txhash, errmsg string) error {
	return updateDepositErrMsgByTxHash(d.st, txhash, errmsg)
}

func (d *DAL) HasDepositTxHash(txhash string) (bool, error) {
	return hasDepositTxHash(d.st, txhash)
}

func (d *DAL) DeleteDeposit(uuid string) error {
	return deleteDeposit(d.st, uuid)
}

func (dtx *DALTx) InsertDeposit(uuid string, cid ctype.CidType, topeer bool, amount *big.Int, refill bool, deadline time.Time, state int, txhash string, errmsg string) error {
	return insertDeposit(dtx.stx, uuid, cid, topeer, amount, refill, deadline, state, txhash, errmsg)
}

func (dtx *DALTx) HasDepositRefillPending(cid ctype.CidType) (bool, error) {
	return hasDepositRefillPending(dtx.stx, cid)
}

func (dtx *DALTx) UpdateDepositStatesByTxHashAndCid(txhash string, cid ctype.CidType, state int) error {
	return updateDepositStatesByTxHashAndCid(dtx.stx, txhash, cid, state)
}

// The "lease" table

func (d *DAL) UpdateLeaseTimestamp(id, owner string) error {
	return updateLeaseTimestamp(d.st, id, owner)
}

func (d *DAL) DeleteLeaseOwner(id, owner string) error {
	return deleteLeaseOwner(d.st, id, owner)
}

func (d *DAL) GetLeaseOwner(id string) (string, bool, error) {
	return getLeaseOwner(d.st, id)
}

func (dtx *DALTx) InsertLease(id, owner string) error {
	return insertLease(dtx.stx, id, owner)
}

func (dtx *DALTx) UpdateLeaseOwner(id, owner string) error {
	return updateLeaseOwner(dtx.stx, id, owner)
}

func (dtx *DALTx) UpdateLeaseTimestamp(id, owner string) error {
	return updateLeaseTimestamp(dtx.stx, id, owner)
}

func (dtx *DALTx) GetLease(id string) (string, time.Time, bool, error) {
	return getLease(dtx.stx, id)
}

func (dtx *DALTx) DeleteLease(id string) error {
	return deleteLease(dtx.stx, id)
}

// ====================== DAL APIs for K/V store ======================

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

func marshalBalance(balance *structs.OnChainBalance) ([]byte, error) {
	return marshal(balanceBigIntToBytes(balance))
}

func unmarshalBalance(balanceBytes []byte) (*structs.OnChainBalance, error) {
	var onChainBalance *structs.OnChainBalance
	var balance OnChainBalance
	err := unmarshal(balanceBytes, &balance)
	if err == nil {
		onChainBalance = balanceBytesToBigInt(&balance)
	}
	return onChainBalance, err
}

func balanceBytesToBigInt(balance *OnChainBalance) *structs.OnChainBalance {
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
	}
}

func balanceBigIntToBytes(balance *structs.OnChainBalance) *OnChainBalance {
	if balance == nil {
		return &OnChainBalance{}
	}
	var pendingWithdrawal PendingWithdrawal
	if balance.PendingWithdrawal != nil {
		pendingWithdrawal.Amount = balance.PendingWithdrawal.Amount.Bytes()
		pendingWithdrawal.Receiver = balance.PendingWithdrawal.Receiver
		pendingWithdrawal.Deadline = balance.PendingWithdrawal.Deadline
	}
	onchainBalance := &OnChainBalance{
		PendingWithdrawal: &pendingWithdrawal,
	}
	if balance.MyDeposit != nil {
		onchainBalance.MyDeposit = balance.MyDeposit.Bytes()
	}
	if balance.MyWithdrawal != nil {
		onchainBalance.MyWithdrawal = balance.MyWithdrawal.Bytes()
	}
	if balance.PeerDeposit != nil {
		onchainBalance.PeerDeposit = balance.PeerDeposit.Bytes()
	}
	if balance.PeerWithdrawal != nil {
		onchainBalance.PeerWithdrawal = balance.PeerWithdrawal.Bytes()
	}
	return onchainBalance
}
