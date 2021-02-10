// Copyright 2019-2020 Celer Network
//
// Inner functions for the Data Access Layer to the SQL tables.
// They are used by the DAL and DALTx wrappers (see dal.go).

package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	listSep   = "," // separator for lists stored in a string column
	separator = "|" // reserved character for keys construction
)

var (
	ErrNoRows     = errors.New("No rows matched in the database")
	ErrTxConflict = errors.New("Transaction conflict")
	ErrTxInvalid  = errors.New("Invalid transaction")
)

type SqlStorage interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Return true if the error database error is not a business logic
// response (e.g. no rows found).
func IsDbError(err error) bool {
	if err == nil || errors.Is(err, ErrNoRows) {
		return false
	}
	return true
}

// Marshal the data into a byte-array if it is not one already.
func marshal(val interface{}) ([]byte, error) {
	switch v := val.(type) {
	case []byte:
		return v, nil
	default:
		return json.Marshal(val)
	}
}

func unmarshal(src []byte, dest interface{}) error {
	switch v := dest.(type) {
	case *[]byte:
		*v = append([]byte(nil), src...)
		return nil
	default:
		return json.Unmarshal(src, dest)
	}
}

// Check if the table and key parameters are valid.
func checkTableKey(table, key string) error {
	if table == "" || key == "" {
		return fmt.Errorf("table and key parameters must be specified")
	}

	// The separator character cannot be used in the table name.
	if strings.Contains(table, separator) {
		return fmt.Errorf("invalid table name: %s", table)
	}
	return nil
}

// storeKey returns the store's key for the given table and entry key.
func storeKey(table, key string) string {
	return table + separator + key
}

// tableKey returns the user visible (table, key) info from a store key.
func tableKey(skey []byte) (string, string) {
	parts := strings.SplitN(string(skey), separator, 2)
	return parts[0], parts[1]
}

func now() time.Time {
	return time.Now().UTC()
}

// Layouts for timestamp parsing to handle CockroachDB and SQLite.
// Ordered from most to least accurate formats.
var timeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05.999999999Z07:00",
	time.RFC3339,
}

func str2Time(ts string) (time.Time, error) {
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid timestamp: %s", ts)
}

func chkExec(res sql.Result, err error, want int64, caller string) error {
	var got int64
	if err == nil {
		got, err = res.RowsAffected()
		if err == nil && got != want {
			if got == 0 {
				// Wrap ErrNoRows with additional info.
				err = fmt.Errorf("%s: invalid SQL #rows: %d != %d: %w", caller, got, want, ErrNoRows)
			} else {
				err = fmt.Errorf("%s: invalid SQL #rows: %d != %d", caller, got, want)
			}
		}
	}
	return err
}

func chkQueryRow(err error) (bool, error) {
	found := false
	if err == nil {
		found = true
	} else if err == sql.ErrNoRows {
		err = nil
	} else {
		log.Debugln("chkQueryRow SQL error:", err)
	}
	return found, err
}

// Return the IN-clause of a SQL query based on the column name, the number
// of its values and their starting position, e.g. "status IN ($3, $4, $5)"
func inClause(column string, num, start int) string {
	if column == "" || num < 1 || start < 1 {
		return ""
	}

	params := make([]string, num)
	for i := 0; i < num; i++ {
		params[i] = fmt.Sprintf("$%d", start+i)
	}

	return fmt.Sprintf("%s IN (%s)", column, strings.Join(params, ", "))
}

// The "channels" table.
func insertChanWithTs(
	st SqlStorage,
	cid ctype.CidType,
	peer ctype.Addr,
	token *entity.TokenInfo,
	ledger ctype.Addr,
	state int,
	stateTs time.Time,
	openTs time.Time,
	openResp *rpc.OpenChannelResponse,
	onChainBalance *structs.OnChainBalance,
	baseSeqNum uint64,
	lastUsedSeqNum uint64,
	lastAckedSeqNum uint64,
	lastNackedSeqNum uint64,
	selfSimplex *rpc.SignedSimplexState,
	peerSimplex *rpc.SignedSimplexState) error {
	q := `INSERT INTO channels (cid, peer, token, ledger, state, statets, opents,
		openresp, onchainbalance, basesn, lastusedsn, lastackedsn,
		lastnackedsn, selfsimplex, peersimplex) VALUES ($1, $2, $3, $4,
		$5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	openRespBytes, err := marshal(openResp)
	if err != nil {
		return err
	}
	balance, err := marshalBalance(onChainBalance)
	if err != nil {
		return err
	}
	balanceBytes, err := marshal(balance)
	if err != nil {
		return err
	}
	selfSimplexBytes, err := marshal(selfSimplex)
	if err != nil {
		return err
	}
	peerSimplexBytes, err := marshal(peerSimplex)
	if err != nil {
		return err
	}
	res, err := st.Exec(q, ctype.Cid2Hex(cid), ctype.Addr2Hex(peer),
		utils.GetTokenAddrStr(token), ctype.Addr2Hex(ledger),
		state, stateTs, openTs, openRespBytes, balanceBytes,
		baseSeqNum, lastUsedSeqNum, lastAckedSeqNum, lastNackedSeqNum,
		selfSimplexBytes, peerSimplexBytes)
	return chkExec(res, err, 1, "insertChanWithTs")
}

func insertChan(
	st SqlStorage,
	cid ctype.CidType,
	peer ctype.Addr,
	token *entity.TokenInfo,
	ledger ctype.Addr,
	state int,
	openResp *rpc.OpenChannelResponse,
	onChainBalance *structs.OnChainBalance,
	baseSeqNum uint64,
	lastUsedSeqNum uint64,
	lastAckedSeqNum uint64,
	lastNackedSeqNum uint64,
	selfSimplex *rpc.SignedSimplexState,
	peerSimplex *rpc.SignedSimplexState) error {
	ts := now()
	return insertChanWithTs(st, cid, peer, token, ledger, state, ts, ts,
		openResp, onChainBalance, baseSeqNum, lastUsedSeqNum,
		lastAckedSeqNum, lastNackedSeqNum, selfSimplex, peerSimplex)
}

func deleteChan(st SqlStorage, cid ctype.CidType) error {
	q := `DELETE FROM channels WHERE cid = $1`
	res, err := st.Exec(q, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "deleteChan")
}

func unmarshalChainInitializer(openRespBytes []byte) (*entity.PaymentChannelInitializer, error) {
	var chaninit entity.PaymentChannelInitializer
	if len(openRespBytes) == 0 {
		return nil, nil
	}
	openResp := new(rpc.OpenChannelResponse)
	err := unmarshal(openRespBytes, openResp)
	if err != nil {
		return nil, err
	}
	if len(openResp.GetChannelInitializer()) == 0 {
		return nil, nil
	}
	err = proto.Unmarshal(openResp.GetChannelInitializer(), &chaninit)
	if err != nil {
		return nil, err
	}
	return &chaninit, nil
}

func getChanViewInfoByID(st SqlStorage, cid ctype.CidType) (
	int, *time.Time, *time.Time, *entity.PaymentChannelInitializer, *structs.OnChainBalance,
	*entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	var state int
	var stateTsStr, openTsStr string
	var openRespBytes, onChainBalanceBytes, selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT state, statets, opents, openresp, onchainbalance, selfsimplex, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&state, &stateTsStr, &openTsStr, &openRespBytes, &onChainBalanceBytes, &selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)

	var chaninit *entity.PaymentChannelInitializer
	var onChainBalance *structs.OnChainBalance
	var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
	var statets, opents time.Time
	if found {
		chaninit, err = unmarshalChainInitializer(openRespBytes)
		if err == nil {
			onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
		}
		if err == nil {
			selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
		}
		if err == nil {
			statets, err = str2Time(stateTsStr)
		}
		if err == nil {
			opents, err = str2Time(openTsStr)
		}
	}
	return state, &statets, &opents, chaninit, onChainBalance, selfSimplex, peerSimplex, found, err
}

func getAllChansByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int) (
	[]ctype.CidType, []ctype.Addr, []*entity.TokenInfo, []*time.Time, []*time.Time, []*structs.OnChainBalance,
	[]*entity.SimplexPaymentChannel, []*entity.SimplexPaymentChannel, error) {
	q := `SELECT cid, peer, token, statets, opents, onchainbalance, selfsimplex, peersimplex FROM channels WHERE token = $1 AND state = $2 ORDER BY statets`
	rows, err := st.Query(q, utils.GetTokenAddrStr(token), state)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()
	return getChanFromRows(rows)
}

func getInactiveChansByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int, stateTs time.Time) (
	[]ctype.CidType, []ctype.Addr, []*entity.TokenInfo, []*time.Time, []*time.Time, []*structs.OnChainBalance,
	[]*entity.SimplexPaymentChannel, []*entity.SimplexPaymentChannel, error) {
	q := `SELECT cid, peer, token, statets, opents, onchainbalance, selfsimplex, peersimplex FROM channels WHERE token = $1 AND state = $2 AND statets < $3 ORDER BY statets`
	rows, err := st.Query(q, utils.GetTokenAddrStr(token), state, stateTs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()
	return getChanFromRows(rows)
}

func getChanFromRows(rows *sql.Rows) (
	[]ctype.CidType, []ctype.Addr, []*entity.TokenInfo, []*time.Time, []*time.Time, []*structs.OnChainBalance,
	[]*entity.SimplexPaymentChannel, []*entity.SimplexPaymentChannel, error) {
	var cids []ctype.CidType
	var peers []ctype.Addr
	var tokens []*entity.TokenInfo
	var stateTses, openTses []*time.Time
	var balances []*structs.OnChainBalance
	var selfSimplexes, peerSimplexes []*entity.SimplexPaymentChannel

	for rows.Next() {
		var cidStr, peerStr, tokenStr, stateTsStr, openTsStr string
		var onChainBalanceBytes, selfSimplexBytes, peerSimplexBytes []byte
		err := rows.Scan(&cidStr, &peerStr, &tokenStr, &stateTsStr, &openTsStr, &onChainBalanceBytes, &selfSimplexBytes, &peerSimplexBytes)
		var onChainBalance *structs.OnChainBalance
		var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
		var statets, opents time.Time
		if err == nil {
			onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
		}
		if err == nil {
			selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
		}
		if err == nil {
			statets, err = str2Time(stateTsStr)
		}
		if err == nil {
			opents, err = str2Time(openTsStr)
		}
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		cids = append(cids, ctype.Hex2Cid(cidStr))
		peers = append(peers, ctype.Hex2Addr(peerStr))
		tokens = append(tokens, utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tokenStr)))
		stateTses = append(stateTses, &statets)
		openTses = append(openTses, &opents)
		balances = append(balances, onChainBalance)
		selfSimplexes = append(selfSimplexes, selfSimplex)
		peerSimplexes = append(peerSimplexes, peerSimplex)
	}

	return cids, peers, tokens, stateTses, openTses, balances, selfSimplexes, peerSimplexes, nil
}

func getCidsByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int) ([]ctype.CidType, error) {
	q := `SELECT cid FROM channels WHERE token = $1 AND state = $2`
	rows, err := st.Query(q, utils.GetTokenAddrStr(token), state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cids []ctype.CidType
	for rows.Next() {
		var cidStr string
		err = rows.Scan(&cidStr)
		if err != nil {
			return nil, err
		}
		cids = append(cids, ctype.Hex2Cid(cidStr))
	}
	return cids, nil
}

func countCidsByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int) (int, error) {
	q := `SELECT COUNT(*) FROM channels WHERE token = $1 AND state = $2`
	var count int
	err := st.QueryRow(q, utils.GetTokenAddrStr(token), state).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getInactiveCidsByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int, stateTs time.Time) ([]ctype.CidType, error) {
	q := `SELECT cid FROM channels WHERE token = $1 AND state = $2 AND statets < $3`
	rows, err := st.Query(q, utils.GetTokenAddrStr(token), state, stateTs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cids []ctype.CidType
	for rows.Next() {
		var cidStr string
		err = rows.Scan(&cidStr)
		if err != nil {
			return nil, err
		}
		cids = append(cids, ctype.Hex2Cid(cidStr))
	}
	return cids, nil
}

func countInactiveCidsByTokenAndState(st SqlStorage, token *entity.TokenInfo, state int, stateTs time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM channels WHERE token = $1 AND state = $2 AND statets < $3`
	var count int
	err := st.QueryRow(q, utils.GetTokenAddrStr(token), state, stateTs).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getChanState(st SqlStorage, cid ctype.CidType) (int, bool, error) {
	var state int
	q := `SELECT state FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&state)
	found, err := chkQueryRow(err)
	return state, found, err
}

func updateChanState(st SqlStorage, cid ctype.CidType, state int) error {
	q := `UPDATE channels SET state = $1, statets = $2 WHERE cid = $3`
	res, err := st.Exec(q, state, now(), ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanState")
}

func getChanOpenResp(st SqlStorage, cid ctype.CidType) (*rpc.OpenChannelResponse, bool, error) {
	var data []byte
	q := `SELECT openresp FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&data)
	found, err := chkQueryRow(err)
	var openResp rpc.OpenChannelResponse
	if found {
		err = unmarshal(data, &openResp)
	}
	return &openResp, found, err
}

func updateChanOpenResp(st SqlStorage, cid ctype.CidType, openResp *rpc.OpenChannelResponse) error {
	q := `UPDATE channels SET openresp = $1 WHERE cid = $2`
	data, err := marshal(openResp)
	if err != nil {
		return err
	}
	res, err := st.Exec(q, data, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanOpenResp")
}

func getChanPeer(st SqlStorage, cid ctype.CidType) (ctype.Addr, bool, error) {
	var peer string
	q := `SELECT peer FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peer)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(peer), found, err
}

func getChanPeerState(st SqlStorage, cid ctype.CidType) (ctype.Addr, int, bool, error) {
	var peer string
	var state int
	q := `SELECT peer, state FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peer, &state)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(peer), state, found, err
}

func getChanSeqNums(st SqlStorage, cid ctype.CidType) (uint64, uint64, uint64, uint64, bool, error) {
	var base, lastUsed, lastAcked, lastNacked uint64
	q := `SELECT basesn, lastusedsn, lastackedsn, lastnackedsn FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&base, &lastUsed, &lastAcked, &lastNacked)
	found, err := chkQueryRow(err)
	return base, lastUsed, lastAcked, lastNacked, found, err
}

func updateChanSeqNums(
	st SqlStorage,
	cid ctype.CidType,
	baseSeqNum uint64,
	lastUsedSeqNum uint64,
	lastAckedSeqNum uint64,
	lastNackedSeqNum uint64) error {
	q := `UPDATE channels SET basesn = $1, lastusedsn = $2, lastackedsn = $3, lastnackedsn = $4 WHERE cid = $5`
	res, err := st.Exec(q, baseSeqNum, lastUsedSeqNum,
		lastAckedSeqNum, lastNackedSeqNum, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanSeqNums")
}

func getChanStateToken(st SqlStorage, cid ctype.CidType) (int, *entity.TokenInfo, bool, error) {
	var token string
	var state int
	q := `SELECT state, token FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&state, &token)
	found, err := chkQueryRow(err)
	return state, utils.GetTokenInfoFromAddress(ctype.Hex2Addr(token)), found, err
}

func getChanForDeposit(st SqlStorage, cid ctype.CidType) (int, *entity.TokenInfo, ctype.Addr, ctype.Addr, bool, error) {
	var token, peer, ledger string
	var state int
	q := `SELECT state, token, peer, ledger FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&state, &token, &peer, &ledger)
	found, err := chkQueryRow(err)
	return state, utils.GetTokenInfoFromAddress(ctype.Hex2Addr(token)), ctype.Hex2Addr(peer), ctype.Hex2Addr(ledger), found, err
}

func getCidByPeerToken(st SqlStorage, peer ctype.Addr, token *entity.TokenInfo) (ctype.CidType, bool, error) {
	var cid string
	q := `SELECT cid FROM channels WHERE peer = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(peer), utils.GetTokenAddrStr(token)).Scan(&cid)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), found, err
}

func getCidStateByPeerToken(st SqlStorage, peer ctype.Addr, token *entity.TokenInfo) (ctype.CidType, int, bool, error) {
	var cid string
	var state int
	q := `SELECT cid, state FROM channels WHERE peer = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(peer), utils.GetTokenAddrStr(token)).Scan(&cid, &state)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), state, found, err
}

func getSelfSimplex(st SqlStorage, cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	var selfSimplexBytes []byte
	q := `SELECT selfsimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&selfSimplexBytes)
	found, err := chkQueryRow(err)
	var selfSimplexState rpc.SignedSimplexState
	var selfSimplexChannel entity.SimplexPaymentChannel
	if found {
		err = unmarshal(selfSimplexBytes, &selfSimplexState)
		if err == nil {
			err = proto.Unmarshal(selfSimplexState.SimplexState, &selfSimplexChannel)
		}
	}
	return &selfSimplexChannel, &selfSimplexState, found, err
}

func getPeerSimplex(st SqlStorage, cid ctype.CidType) (*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	var peerSimplexBytes []byte
	q := `SELECT peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peerSimplexBytes)
	found, err := chkQueryRow(err)
	var peerSimplexState rpc.SignedSimplexState
	var peerSimplexChannel entity.SimplexPaymentChannel
	if found {
		err = unmarshal(peerSimplexBytes, &peerSimplexState)
		if err == nil {
			err = proto.Unmarshal(peerSimplexState.SimplexState, &peerSimplexChannel)
		}
	}
	return &peerSimplexChannel, &peerSimplexState, found, err
}

func getDuplexChannel(st SqlStorage, cid ctype.CidType) (
	*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, *entity.SimplexPaymentChannel, *rpc.SignedSimplexState, bool, error) {
	var selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT selfsimplex, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var selfSimplexChannel, peerSimplexChannel *entity.SimplexPaymentChannel
	var selfSimplexState, peerSimplexState *rpc.SignedSimplexState
	if found {
		selfSimplexChannel, selfSimplexState, peerSimplexChannel, peerSimplexState, err =
			unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
	}
	return selfSimplexChannel, selfSimplexState, peerSimplexChannel, peerSimplexState, found, err
}

func getChanForBalance(st SqlStorage, cid ctype.CidType) (
	ctype.Addr, *structs.OnChainBalance, uint64, uint64,
	*entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	var peer string
	var onChainBalanceBytes []byte
	var baseSeq, lastAckedSeq uint64
	var selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT peer, onchainbalance, basesn, lastackedsn, selfsimplex, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&peer, &onChainBalanceBytes, &baseSeq, &lastAckedSeq, &selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var onChainBalance *structs.OnChainBalance
	var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
	if found {
		onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
		if err == nil {
			selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
		}
	}

	return ctype.Hex2Addr(peer), onChainBalance, baseSeq, lastAckedSeq, selfSimplex, peerSimplex, found, err
}

func getOnChainBalance(st SqlStorage, cid ctype.CidType) (*structs.OnChainBalance, bool, error) {
	var onChainBalanceBytes []byte
	q := `SELECT onchainbalance FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&onChainBalanceBytes)
	found, err := chkQueryRow(err)
	var onChainBalance *structs.OnChainBalance
	if found {
		onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
	}
	return onChainBalance, found, err
}

func updateOnChainBalance(st SqlStorage, cid ctype.CidType, onChainBalance *structs.OnChainBalance) error {
	balance, err := marshalBalance(onChainBalance)
	if err != nil {
		return err
	}
	q := `UPDATE channels SET onchainbalance = $1 WHERE cid = $2`
	res, err := st.Exec(q, balance, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateOnChainBalance")
}

func getChanForSendCondPayRequest(st SqlStorage, cid ctype.CidType) (
	ctype.Addr, int, *structs.OnChainBalance, uint64, uint64, uint64,
	*entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	var peer string
	var state int
	var onChainBalanceBytes []byte
	var baseSeq, lastUsedSeq, lastAckedSeq uint64
	var selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT peer, state, onchainbalance, basesn, lastusedsn, lastackedsn, selfsimplex, peersimplex
		FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&peer, &state, &onChainBalanceBytes, &baseSeq, &lastUsedSeq, &lastAckedSeq, &selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var onChainBalance *structs.OnChainBalance
	var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
	if found {
		onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
		if err == nil {
			selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
		}
	}

	return ctype.Hex2Addr(peer), state, onChainBalance, baseSeq, lastUsedSeq, lastAckedSeq, selfSimplex, peerSimplex, found, err
}

func getChanForSendPaySettleRequest(st SqlStorage, cid ctype.CidType) (
	ctype.Addr, int, uint64, uint64, uint64, *entity.SimplexPaymentChannel, bool, error) {
	var peer string
	var state int
	var baseSeq, lastUsedSeq, lastAckedSeq uint64
	var selfSimplexBytes []byte
	q := `SELECT peer, state, basesn, lastusedsn, lastackedsn, selfsimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&peer, &state, &baseSeq, &lastUsedSeq, &lastAckedSeq, &selfSimplexBytes)
	found, err := chkQueryRow(err)
	var selfSimplexChannel *entity.SimplexPaymentChannel
	if found {
		selfSimplexChannel, err = simplexBytes2Channel(selfSimplexBytes)
	}

	return ctype.Hex2Addr(peer), state, baseSeq, lastUsedSeq, lastAckedSeq, selfSimplexChannel, found, err
}

func updateChanForSendRequest(st SqlStorage, cid ctype.CidType, baseSeqNum uint64, lastUsedSeqNum uint64) error {
	q := `UPDATE channels SET statets = $1, basesn = $2, lastusedsn = $3 WHERE cid = $4`
	res, err := st.Exec(q, now(), baseSeqNum, lastUsedSeqNum, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanForSendRequest")
}

func getChanForRecvPayRequest(st SqlStorage, cid ctype.CidType) (
	ctype.Addr, int, *structs.OnChainBalance, uint64, uint64,
	*entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	var peer string
	var state int
	var onChainBalanceBytes []byte
	var baseSeq, lastAckedSeq uint64
	var selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT peer, state, onchainbalance, basesn, lastackedsn, selfsimplex, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&peer, &state, &onChainBalanceBytes, &baseSeq, &lastAckedSeq, &selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var onChainBalance *structs.OnChainBalance
	var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
	if found {
		onChainBalance, err = unmarshalBalance(onChainBalanceBytes)
		if err == nil {
			selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
		}
	}
	return ctype.Hex2Addr(peer), state, onChainBalance, baseSeq, lastAckedSeq, selfSimplex, peerSimplex, found, err
}

func getChanStateAndPeerSimplex(st SqlStorage, cid ctype.CidType) (
	int, *entity.SimplexPaymentChannel, bool, error) {
	var state int
	var peerSimplexBytes []byte
	q := `SELECT state, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&state, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var peerSimplex *entity.SimplexPaymentChannel
	if found {
		peerSimplex, err = simplexBytes2Channel(peerSimplexBytes)
	}
	return state, peerSimplex, found, err
}

func updateChanForRecvRequest(
	st SqlStorage,
	cid ctype.CidType,
	peerSimplex *rpc.SignedSimplexState) error {
	peerSimplexBytes, err := marshal(peerSimplex)
	if err != nil {
		return err
	}
	q := `UPDATE channels SET statets = $1, peersimplex = $2 WHERE cid = $3`
	res, err := st.Exec(q, now(), peerSimplexBytes, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanForRecvRequest")
}

func getChanForRecvResponse(st SqlStorage, cid ctype.CidType) (ctype.Addr, int, uint64, uint64, uint64, uint64, bool, error) {
	var peer string
	var state int
	var baseSeq, lastUsedSeq, lastAckedSeq, lastNackedSeq uint64
	q := `SELECT peer, state, basesn, lastusedsn, lastackedsn, lastnackedsn FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(
		&peer, &state, &baseSeq, &lastUsedSeq, &lastAckedSeq, &lastNackedSeq)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(peer), state, baseSeq, lastUsedSeq, lastAckedSeq, lastNackedSeq, found, err
}

func updateChanForRecvResponse(
	st SqlStorage,
	cid ctype.CidType,
	baseSeqNum uint64,
	lastUsedSeqNum uint64,
	lastAckedSeqNum uint64,
	lastNackedSeqNum uint64,
	selfSimplex *rpc.SignedSimplexState) error {
	selfSimplexBytes, err := marshal(selfSimplex)
	if err != nil {
		return err
	}
	q := `UPDATE channels SET statets = $1, basesn = $2, lastusedsn = $3,
		lastackedsn = $4, lastnackedsn = $5, selfsimplex = $6 WHERE cid = $7`
	res, err := st.Exec(q, now(), baseSeqNum, lastUsedSeqNum, lastAckedSeqNum,
		lastNackedSeqNum, selfSimplexBytes, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanForRecvResponse")
}

func getChanForIntendSettle(st SqlStorage, cid ctype.CidType) (
	ctype.Addr, int, *entity.SimplexPaymentChannel, *entity.SimplexPaymentChannel, bool, error) {
	var peer string
	var state int
	var selfSimplexBytes, peerSimplexBytes []byte
	q := `SELECT peer, state, selfsimplex, peersimplex FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peer, &state, &selfSimplexBytes, &peerSimplexBytes)
	found, err := chkQueryRow(err)
	var selfSimplex, peerSimplex *entity.SimplexPaymentChannel
	if found {
		selfSimplex, _, peerSimplex, _, err = unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes)
	}

	return ctype.Hex2Addr(peer), state, selfSimplex, peerSimplex, found, err
}

func unmarshalDuplexChannel(selfSimplexBytes, peerSimplexBytes []byte) (
	*entity.SimplexPaymentChannel, *rpc.SignedSimplexState,
	*entity.SimplexPaymentChannel, *rpc.SignedSimplexState, error) {
	var selfSimplexChannel, peerSimplexChannel entity.SimplexPaymentChannel
	var selfSimplexState, peerSimplexState rpc.SignedSimplexState
	err := unmarshal(selfSimplexBytes, &selfSimplexState)
	if err == nil {
		err = unmarshal(peerSimplexBytes, &peerSimplexState)
	}
	if err == nil {
		err = proto.Unmarshal(selfSimplexState.SimplexState, &selfSimplexChannel)
	}
	if err == nil {
		err = proto.Unmarshal(peerSimplexState.SimplexState, &peerSimplexChannel)
	}

	return &selfSimplexChannel, &selfSimplexState, &peerSimplexChannel, &peerSimplexState, err
}

func getChannelsForAuthReq(st SqlStorage, peerAddr ctype.Addr) ([]*rpc.ChannelSummary, error) {
	q := `SELECT cid, lastackedsn, peersimplex FROM channels WHERE peer = $1`
	rows, err := st.Query(q, ctype.Addr2Hex(peerAddr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []*rpc.ChannelSummary
	var cid string
	var lastAcked uint64
	for rows.Next() {
		var peerSimplexBytes []byte
		err = rows.Scan(&cid, &lastAcked, &peerSimplexBytes)
		if err != nil {
			log.Error(err)
			// best effort try including more channel summary so move to next row instead of err return
			continue
		}
		peerSimplexChannel, err2 := simplexBytes2Channel(peerSimplexBytes)
		if err2 != nil {
			log.Error(err2)
			continue
		}
		ret = append(ret, &rpc.ChannelSummary{
			ChannelId:  ctype.Hex2Bytes(cid),
			MySeqNum:   lastAcked,
			PeerSeqNum: peerSimplexChannel.SeqNum,
		})
	}
	return ret, nil
}

func getCidTokensByPeer(st SqlStorage, peerAddr ctype.Addr) ([]ctype.CidType, []ctype.Addr, error) {
	q := `SELECT cid, token FROM channels WHERE peer = $1`
	rows, err := st.Query(q, ctype.Addr2Hex(peerAddr))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var cids []ctype.CidType
	var tokens []ctype.Addr
	for rows.Next() {
		var cid, token string
		err = rows.Scan(&cid, &token)
		if err != nil {
			return nil, nil, err
		}
		cids = append(cids, ctype.Hex2Cid(cid))
		tokens = append(tokens, ctype.Hex2Addr(token))
	}
	return cids, tokens, nil
}

func getChanForClose(st SqlStorage, cid ctype.CidType) (ctype.Addr, *entity.TokenInfo, time.Time, bool, error) {
	var peer, token string
	var openTsStr string
	q := `SELECT peer, token, opents FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peer, &token, &openTsStr)
	found, err := chkQueryRow(err)
	var opents time.Time
	if found && err == nil {
		opents, err = str2Time(openTsStr)
	}
	return ctype.Hex2Addr(peer), utils.GetTokenInfoFromAddress(ctype.Hex2Addr(token)), opents, found, err
}

// intermediate data format for easier handling in auth
type chanForAuthAck struct {
	Cid                  string
	State                int64
	OpenChanResp         *rpc.OpenChannelResponse
	MySigned, PeerSigned *rpc.SignedSimplexState
	MySeq, PeerSeq       uint64
	LedgerAddr           ctype.Addr
}

func getChannelsForAuthAck(st SqlStorage, peerAddr ctype.Addr) ([]*chanForAuthAck, error) {
	q := `SELECT cid, state, openresp, lastackedsn, selfsimplex, peersimplex, ledger FROM channels WHERE peer = $1`
	rows, err := st.Query(q, ctype.Addr2Hex(peerAddr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []*chanForAuthAck
	var cid string
	var state int64
	var myseq uint64
	for rows.Next() {
		var openresp, selfsimplex, peersimplex []byte
		var ledgerAddrStr string
		err = rows.Scan(&cid, &state, &openresp, &myseq, &selfsimplex, &peersimplex, &ledgerAddrStr)
		if err != nil {
			log.Error(err)
			continue
		}
		newchan := &chanForAuthAck{
			Cid:        cid,
			State:      state,
			MySeq:      myseq,
			LedgerAddr: ctype.Hex2Addr(ledgerAddrStr),
		}
		if len(openresp) > 0 { // openresp could be null
			openrespMsg := new(rpc.OpenChannelResponse)
			err = unmarshal(openresp, openrespMsg)
			if err == nil {
				newchan.OpenChanResp = openrespMsg
			}
		}
		selfSigned, peerSigned := new(rpc.SignedSimplexState), new(rpc.SignedSimplexState)
		err = unmarshal(selfsimplex, selfSigned)
		if err == nil {
			newchan.MySigned = selfSigned
		}
		err = unmarshal(peersimplex, peerSigned)
		if err == nil {
			newchan.PeerSigned = peerSigned
		}
		peerChannel := new(entity.SimplexPaymentChannel)
		err = proto.Unmarshal(peerSigned.SimplexState, peerChannel)
		if err == nil {
			newchan.PeerSeq = peerChannel.SeqNum
		}
		ret = append(ret, newchan)
	}
	return ret, nil
}

// parse simplex bytes from channels table and return SimplexPaymentChannel
func simplexBytes2Channel(simplexBytes []byte) (*entity.SimplexPaymentChannel, error) {
	signedSimplexState := new(rpc.SignedSimplexState)
	err := unmarshal(simplexBytes, signedSimplexState)
	if err != nil {
		return nil, err
	}
	ret := new(entity.SimplexPaymentChannel)
	err = proto.Unmarshal(signedSimplexState.SimplexState, ret)
	return ret, err
}

func getChanLedger(st SqlStorage, cid ctype.CidType) (ctype.Addr, bool, error) {
	var ledger string
	q := `SELECT ledger FROM channels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&ledger)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(ledger), found, err
}

func updateChanLedger(st SqlStorage, cid ctype.CidType, ledger ctype.Addr) error {
	q := `UPDATE channels SET ledger = $1 WHERE cid = $2`
	res, err := st.Exec(q, ctype.Addr2Hex(ledger), ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "updateChanLedger")
}

func getChanForMigration(st SqlStorage, cid ctype.CidType) (int, ctype.Addr, bool, error) {
	q := `SELECT state, ledger FROM channels WHERE cid = $1`
	var state int
	var ledger string
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&state, &ledger)
	found, err := chkQueryRow(err)

	return state, ctype.Hex2Addr(ledger), found, err
}

func getAllChanLedgers(st SqlStorage) ([]ctype.Addr, error) {
	q := `SELECT DISTINCT ledger FROM channels`

	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ledgers []ctype.Addr
	for rows.Next() {
		var ledger string
		err = rows.Scan(&ledger)
		if err != nil {
			return nil, err
		}
		ledgers = append(ledgers, ctype.Hex2Addr(ledger))
	}

	return ledgers, nil
}

// The "messages" table.
func insertChanMessage(
	st SqlStorage,
	cid ctype.CidType,
	seqnum uint64,
	msg *rpc.CelerMsg) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	q := `INSERT INTO chanmessages (cid, seqnum, msg) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, ctype.Cid2Hex(cid), seqnum, msgBytes)
	return chkExec(res, err, 1, "insertMessage")
}

func getChanMessage(
	st SqlStorage,
	cid ctype.CidType,
	seqnum uint64) (*rpc.CelerMsg, bool, error) {
	var msgBytes []byte
	q := `SELECT msg FROM chanmessages WHERE cid = $1 AND seqnum = $2`
	err := st.QueryRow(q, ctype.Cid2Hex(cid), seqnum).Scan(&msgBytes)
	found, err := chkQueryRow(err)
	var msg rpc.CelerMsg
	if found {
		err = proto.Unmarshal(msgBytes, &msg)
	}
	return &msg, found, err
}

func getAllChanMessages(st SqlStorage, cid ctype.CidType) ([]*rpc.CelerMsg, error) {
	q := `SELECT msg FROM chanmessages WHERE cid = $1 ORDER BY seqnum ASC`
	rows, err := st.Query(q, ctype.Cid2Hex(cid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*rpc.CelerMsg
	for rows.Next() {
		var msgBytes []byte
		err = rows.Scan(&msgBytes)
		if err != nil {
			return nil, err
		}
		var msg rpc.CelerMsg
		err = proto.Unmarshal(msgBytes, &msg)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func deleteChanMessage(st SqlStorage, cid ctype.CidType, seqnum uint64) error {
	q := `DELETE FROM chanmessages WHERE cid = $1 AND seqnum = $2`
	res, err := st.Exec(q, ctype.Cid2Hex(cid), seqnum)
	return chkExec(res, err, 1, "deleteMessage")
}

// The "closedchannels" table.
func insertClosedChan(
	st SqlStorage,
	cid ctype.CidType,
	peer ctype.Addr,
	token *entity.TokenInfo,
	openTs time.Time,
	closeTs time.Time) error {
	q := `INSERT INTO closedchannels (cid, peer, token, opents, closets)
		VALUES ($1, $2, $3, $4, $5)`
	res, err := st.Exec(q, ctype.Cid2Hex(cid), ctype.Addr2Hex(peer),
		utils.GetTokenAddrStr(token), openTs, closeTs)
	return chkExec(res, err, 1, "insertClosedChan")
}

func getClosedChan(st SqlStorage, cid ctype.CidType) (ctype.Addr, *entity.TokenInfo, *time.Time, *time.Time, bool, error) {
	var peer, token string
	var openTsStr, closeTsStr string
	q := `SELECT peer, token, opents, closets FROM closedchannels WHERE cid = $1`
	err := st.QueryRow(q, ctype.Cid2Hex(cid)).Scan(&peer, &token, &openTsStr, &closeTsStr)
	found, err := chkQueryRow(err)
	var opents, closets time.Time
	if found {
		opents, err = str2Time(openTsStr)
		if err == nil {
			closets, err = str2Time(closeTsStr)
		}
	}
	return ctype.Hex2Addr(peer), utils.GetTokenInfoFromAddress(ctype.Hex2Addr(token)), &opents, &closets, found, err
}

// The "payments" table.
func insertPaymentWithTs(st SqlStorage, payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any, inCid ctype.CidType, inState int, outCid ctype.CidType, outState int, createTs time.Time) error {
	noteBytes, err := marshal(note)
	if err != nil {
		return err
	}
	var inCidStr, outCidStr string
	if inCid != ctype.ZeroCid {
		inCidStr = ctype.Cid2Hex(inCid)
	}
	if outCid != ctype.ZeroCid {
		outCidStr = ctype.Cid2Hex(outCid)
	}
	q := `INSERT INTO payments (payid, pay, paynote, incid, instate, outcid, outstate, src, dest, createts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	res, err := st.Exec(
		q, ctype.PayID2Hex(payID), payBytes, noteBytes, inCidStr, inState, outCidStr, outState,
		ctype.Bytes2Hex(pay.GetSrc()), ctype.Bytes2Hex(pay.GetDest()), createTs)
	return chkExec(res, err, 1, "insertPaymentWithTs")
}

func insertPayment(st SqlStorage, payID ctype.PayIDType, payBytes []byte, pay *entity.ConditionalPay, note *any.Any, inCid ctype.CidType, inState int, outCid ctype.CidType, outState int) error {
	return insertPaymentWithTs(st, payID, payBytes, pay, note, inCid, inState, outCid, outState, now())
}

func deletePayment(st SqlStorage, payID ctype.PayIDType) error {
	q := `DELETE FROM payments WHERE payid = $1`
	res, err := st.Exec(q, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "deletePayment")
}

func getAllPayIDs(st SqlStorage) ([]ctype.PayIDType, error) {
	q := `SELECT payid FROM payments ORDER BY createts DESC`
	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payids []ctype.PayIDType
	var data string
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		payids = append(payids, ctype.Hex2PayID(data))
	}

	return payids, nil
}

func getPayment(st SqlStorage, payID ctype.PayIDType) (*entity.ConditionalPay, []byte, bool, error) {
	var payBytes []byte
	q := `SELECT pay FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&payBytes)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	if found {
		err = proto.Unmarshal(payBytes, &pay)
	}
	return &pay, payBytes, found, err
}

func getPayNote(st SqlStorage, payID ctype.PayIDType) (*any.Any, bool, error) {
	var noteBytes []byte
	q := `SELECT paynote FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&noteBytes)
	found, err := chkQueryRow(err)
	var note any.Any
	if found {
		err = unmarshal(noteBytes, &note)
	}
	return &note, found, err
}

func getPaymentInfo(st SqlStorage, payID ctype.PayIDType) (
	*entity.ConditionalPay, *any.Any, ctype.CidType, int, ctype.CidType, int, *time.Time, bool, error) {
	var payBytes, noteBytes []byte
	var inCid, outCid, createTsStr string
	var inState, outState int
	var createTs time.Time
	q := `SELECT pay, paynote, incid, instate, outcid, outstate, createts FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(
		&payBytes, &noteBytes, &inCid, &inState, &outCid, &outState, &createTsStr)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	var note any.Any
	if found {
		err = proto.Unmarshal(payBytes, &pay)
		if err == nil {
			err = unmarshal(noteBytes, &note)
		}
		if err == nil {
			createTs, err = str2Time(createTsStr)
		}
	}
	return &pay, &note, ctype.Hex2Cid(inCid), inState, ctype.Hex2Cid(outCid), outState, &createTs, found, err
}

func getAllPaymentInfoByCid(st SqlStorage, cid ctype.CidType) (
	[]ctype.PayIDType, []*entity.ConditionalPay, []*any.Any, []ctype.CidType, []int, []ctype.CidType, []int, []*time.Time, error) {
	q := `SELECT payid, pay, paynote, incid, instate, outcid, outstate, createts FROM payments WHERE incid = $1 OR outcid = $1 ORDER BY createts DESC`
	rows, err := st.Query(q, ctype.Cid2Hex(cid))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	var payIDs []ctype.PayIDType
	var pays []*entity.ConditionalPay
	var notes []*any.Any
	var incids, outcids []ctype.CidType
	var instates, outstates []int
	var createTses []*time.Time

	for rows.Next() {
		var payBytes, noteBytes []byte
		var payID, inCid, outCid, createTsStr string
		var inState, outState int
		var createTs time.Time
		var pay entity.ConditionalPay
		var note any.Any
		err = rows.Scan(&payID, &payBytes, &noteBytes, &inCid, &inState, &outCid, &outState, &createTsStr)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		err = proto.Unmarshal(payBytes, &pay)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		err = unmarshal(noteBytes, &note)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		createTs, err = str2Time(createTsStr)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		payIDs = append(payIDs, ctype.Hex2PayID(payID))
		pays = append(pays, &pay)
		notes = append(notes, &note)
		incids = append(incids, ctype.Hex2Cid(inCid))
		instates = append(instates, inState)
		outcids = append(outcids, ctype.Hex2Cid(outCid))
		outstates = append(outstates, outState)
		createTses = append(createTses, &createTs)
	}
	return payIDs, pays, notes, incids, instates, outcids, outstates, createTses, nil
}

func getPayStates(st SqlStorage, payID ctype.PayIDType) (int, int, bool, error) {
	var inState, outState int
	q := `SELECT instate, outstate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&inState, &outState)
	found, err := chkQueryRow(err)
	return inState, outState, found, err
}

func getPayForRecvSettleReq(st SqlStorage, payID ctype.PayIDType) (*entity.ConditionalPay, *any.Any, ctype.CidType, int, int, bool, error) {
	var payBytes, noteBytes []byte
	var inState, outState int
	var inCid string
	q := `SELECT pay, paynote, incid, instate, outstate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&payBytes, &noteBytes, &inCid, &inState, &outState)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	var note any.Any
	if found {
		err = proto.Unmarshal(payBytes, &pay)
		if err == nil {
			err = unmarshal(noteBytes, &note)
		}
	}
	return &pay, &note, ctype.Hex2Cid(inCid), inState, outState, found, err
}

func getPayForRecvSettleProof(st SqlStorage, payID ctype.PayIDType) (*entity.ConditionalPay, ctype.Addr, bool, error) {
	var payBytes []byte
	var peer string
	q := `
		SELECT p.pay, c.peer
		FROM payments AS p
		JOIN channels AS c ON p.outcid = c.cid
		WHERE payid = $1
	`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&payBytes, &peer)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	if found {
		err = proto.Unmarshal(payBytes, &pay)
	}
	return &pay, ctype.Hex2Addr(peer), found, err
}

func getPayIngressPeer(st SqlStorage, payID ctype.PayIDType) (ctype.Addr, bool, error) {
	var peer string
	q := `
		SELECT c.peer
		FROM payments AS p
		JOIN channels AS c ON p.incid = c.cid
		WHERE payid = $1
	`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&peer)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(peer), found, err
}

func getPayForRecvSecret(st SqlStorage, payID ctype.PayIDType) (*entity.ConditionalPay, *any.Any, int, bool, error) {
	var payBytes, noteBytes []byte
	var state int
	q := `SELECT pay, paynote, instate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&payBytes, &noteBytes, &state)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	var note any.Any
	if found {
		err = proto.Unmarshal(payBytes, &pay)
		if err == nil {
			err = unmarshal(noteBytes, &note)
		}
	}
	return &pay, &note, state, found, err
}

func getPayAndEgressState(st SqlStorage, payID ctype.PayIDType) (*entity.ConditionalPay, []byte, int, bool, error) {
	var payBytes []byte
	var state int
	q := `SELECT pay, outstate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&payBytes, &state)
	found, err := chkQueryRow(err)
	var pay entity.ConditionalPay
	if found {
		err = proto.Unmarshal(payBytes, &pay)
	}
	return &pay, payBytes, state, found, err
}

func getPayIngress(st SqlStorage, payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	var state int
	var cid string
	q := `SELECT incid, instate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&cid, &state)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), state, found, err
}

func getPayIngressChannel(st SqlStorage, payID ctype.PayIDType) (ctype.CidType, bool, error) {
	var cid string
	q := `SELECT incid FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&cid)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), found, err
}

func getPayHistory(
	st SqlStorage, peer ctype.Addr, beforeTs time.Time, smallestPayID ctype.PayIDType, maxResultSize int32) ([]ctype.PayIDType, []*entity.ConditionalPay, []int64, []int64, error) {
	smallestPayIDHex := ctype.PayID2Hex(smallestPayID)
	peerHex := ctype.Addr2Hex(peer)
	q := `SELECT payid, pay, instate, createts FROM payments WHERE (src=$1 OR dest=$2) AND (createts<$3 OR (createts=$4 AND payid>$5)) ORDER BY createts DESC, payid ASC LIMIT $6`
	// To be sqllite-compatible, $x can't be used as variable. They are simply placeholders like %s in printf.
	rows, err := st.Query(q, peerHex, peerHex, beforeTs, beforeTs, smallestPayIDHex, maxResultSize)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer rows.Close()

	var payIDs []ctype.PayIDType
	var pays []*entity.ConditionalPay
	var instates []int64
	var createTses []int64
	for rows.Next() {
		var payIDStr string
		var payBytes []byte
		var instate int64
		var createTsTime time.Time
		var createTsStr string
		err = rows.Scan(&payIDStr, &payBytes, &instate, &createTsStr)
		if err != nil {
			// Rather errors than returning incomplete data
			return nil, nil, nil, nil, err
		}
		createTsTime, err = str2Time(createTsStr)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		pay := &entity.ConditionalPay{}
		err = proto.Unmarshal(payBytes, pay)
		if err != nil {
			// Rather errors than returning incomplete data
			return nil, nil, nil, nil, err
		}
		payIDs = append(payIDs, ctype.Hex2PayID(payIDStr))
		pays = append(pays, pay)
		instates = append(instates, instate)
		createTses = append(createTses, createTsTime.Unix())
	}
	return payIDs, pays, instates, createTses, nil
}

func getPaysForAuthAck(st SqlStorage, payIDs []ctype.PayIDType, isOut bool) ([]*rpc.PayInAuthAck, error) {
	if len(payIDs) < 1 {
		return nil, nil
	}
	stateColumn := "instate"
	if isOut {
		stateColumn = "outstate"
	}
	q := fmt.Sprintf("SELECT pay, paynote, %s FROM payments WHERE %s", stateColumn, inClause("payid", len(payIDs), 1))

	var args []interface{}
	for _, payid := range payIDs {
		args = append(args, ctype.PayID2Hex(payid))
	}
	rows, err := st.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []*rpc.PayInAuthAck
	for rows.Next() {
		var payByte, payNoteByte []byte
		var state int64
		err = rows.Scan(&payByte, &payNoteByte, &state)
		if err != nil {
			continue
		}
		pay := &rpc.PayInAuthAck{
			Pay:   payByte,
			State: state,
		}
		if len(payNoteByte) > 0 {
			var note any.Any
			err = unmarshal(payNoteByte, &note)
			if err != nil {
				log.Warn(err)
			} else {
				pay.Note = &note
			}
		}
		ret = append(ret, pay)
	}
	return ret, nil
}

func updatePayIngressState(st SqlStorage, payID ctype.PayIDType, state int) error {
	q := `UPDATE payments SET instate = $1 WHERE payid = $2`
	res, err := st.Exec(q, state, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "updatePayIngressState")
}

func getPayEgress(st SqlStorage, payID ctype.PayIDType) (ctype.CidType, int, bool, error) {
	var state int
	var cid string
	q := `SELECT outcid, outstate FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&cid, &state)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), state, found, err
}

func getPayEgressChannel(st SqlStorage, payID ctype.PayIDType) (ctype.CidType, bool, error) {
	var cid string
	q := `SELECT outcid FROM payments WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&cid)
	found, err := chkQueryRow(err)
	return ctype.Hex2Cid(cid), found, err
}

func updatePayEgress(st SqlStorage, payID ctype.PayIDType, cid ctype.CidType, state int) error {
	q := `UPDATE payments SET outcid = $1, outstate = $2 WHERE payid = $3`
	res, err := st.Exec(q, ctype.Cid2Hex(cid), state, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "updatePayEgress")
}

func updatePayEgressState(st SqlStorage, payID ctype.PayIDType, state int) error {
	q := `UPDATE payments SET outstate = $1 WHERE payid = $2`
	res, err := st.Exec(q, state, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "updatePayEgressState")
}

func countPayments(st SqlStorage) (int, error) {
	var count int
	q := `SELECT COUNT(*) FROM payments`
	err := st.QueryRow(q).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// The "paydelegation" table.
func insertDelegatedPay(
	st SqlStorage,
	payID ctype.PayIDType,
	dest ctype.Addr,
	status int) error {
	q := `INSERT INTO paydelegation (payid, dest, status) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, ctype.PayID2Hex(payID), ctype.Addr2Hex(dest), status)
	return chkExec(res, err, 1, "insertDelegatedPay")
}

func deleteDelegatedPay(st SqlStorage, payID ctype.PayIDType) error {
	q := `DELETE FROM paydelegation WHERE payid = $1`
	res, err := st.Exec(q, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "deletetDelegatedPay")
}

func getDelegatedPayStatus(st SqlStorage, payID ctype.PayIDType) (int, bool, error) {
	var status int
	q := `SELECT status FROM paydelegation WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&status)
	found, err := chkQueryRow(err)
	return status, found, err
}

func updateDelegatedPayStatus(st SqlStorage, payID ctype.PayIDType, status int) error {
	q := `UPDATE paydelegation SET status = $1 WHERE payid = $2`
	res, err := st.Exec(q, status, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "updateDelegatedPayStatus")
}

func updateSendingDelegatedPay(st SqlStorage, payID, payIDout ctype.PayIDType) error {
	var payidOutStr sql.NullString
	payidOutStr.String = ctype.PayID2Hex(payIDout)
	payidOutStr.Valid = true
	q := `UPDATE paydelegation SET status = $1, payidout = $2 WHERE payid = $3 AND status = $4`
	res, err := st.Exec(
		q, structs.DelegatedPayStatus_SENDING, payidOutStr, ctype.PayID2Hex(payID), structs.DelegatedPayStatus_RECVD)
	return chkExec(res, err, 1, "updateSendingDelegatedPay")
}

func getDelegatedPaysOnStatus(st SqlStorage, dest ctype.Addr, status int) (map[ctype.PayIDType]*entity.ConditionalPay, error) {
	q := ` 
		SELECT d.payid, p.pay
		FROM paydelegation AS d
		JOIN payments AS p ON d.payid = p.payid
		WHERE d.dest = $1 AND d.status = $2
	`
	rows, err := st.Query(q, ctype.Addr2Hex(dest), status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pays := make(map[ctype.PayIDType]*entity.ConditionalPay)
	for rows.Next() {
		var payID string
		var payBytes []byte
		err = rows.Scan(&payID, &payBytes)
		if err != nil {
			return nil, err
		}
		var pay entity.ConditionalPay
		err = proto.Unmarshal(payBytes, &pay)
		if err != nil {
			return nil, err
		}
		pays[ctype.Hex2PayID(payID)] = &pay
	}

	return pays, nil
}

func insertPayDelegator(
	st SqlStorage,
	payID ctype.PayIDType,
	dest ctype.Addr,
	delegator ctype.Addr) error {
	var delegatorStr sql.NullString
	delegatorStr.String = ctype.Addr2Hex(delegator)
	delegatorStr.Valid = true
	q := `INSERT INTO paydelegation (payid, dest, status, delegator) VALUES ($1, $2, $3, $4)`
	res, err := st.Exec(q, ctype.PayID2Hex(payID), ctype.Addr2Hex(dest), 0, delegatorStr)
	return chkExec(res, err, 1, "insertPayDelegator")
}

func getPayDelegator(st SqlStorage, payID ctype.PayIDType) (ctype.Addr, bool, error) {
	var delegatorStr sql.NullString
	q := `SELECT delegator FROM paydelegation WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&delegatorStr)
	found, err := chkQueryRow(err)
	var delegator ctype.Addr
	if found {
		found = delegatorStr.Valid
		delegator = ctype.Hex2Addr(delegatorStr.String)
	}
	return delegator, found, err
}

// The "crossnetpays" table.
func insertCrossNetPay(
	st SqlStorage,
	payID, originalPayID ctype.PayIDType,
	originalPay []byte, state int,
	srcNetId, DstNetId uint64,
	bridgeAddr ctype.Addr, bridgeNetId uint64) error {
	q := `INSERT INTO crossnetpays (payid, originalpayid, originalpay, state, srcnetid, dstnetid, bridgeaddr, bridgenetid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	res, err := st.Exec(
		q, ctype.PayID2Hex(payID), ctype.PayID2Hex(originalPayID), originalPay, state,
		srcNetId, DstNetId, ctype.Addr2Hex(bridgeAddr), bridgeNetId)
	return chkExec(res, err, 1, "insertCrossNetPay")
}

func getCrossNetInfoByPayID(st SqlStorage, payID ctype.PayIDType) (ctype.PayIDType, int, ctype.Addr, bool, error) {
	var originalPayID string
	var state int
	var bridgeAddr string
	q := `SELECT originalpayid, state, bridgeaddr FROM crossnetpays WHERE payid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(payID)).Scan(&originalPayID, &state, &bridgeAddr)
	found, err := chkQueryRow(err)
	return ctype.Hex2PayID(originalPayID), state, ctype.Hex2Addr(bridgeAddr), found, err
}

func getCrossNetInfoByOrignalPayID(st SqlStorage, originalPayID ctype.PayIDType) (ctype.PayIDType, int, ctype.Addr, bool, error) {
	var payID string
	var state int
	var bridgeAddr string
	q := `SELECT payid, state, bridgeaddr FROM crossnetpays WHERE originalpayid = $1`
	err := st.QueryRow(q, ctype.PayID2Hex(originalPayID)).Scan(&payID, &state, &bridgeAddr)
	found, err := chkQueryRow(err)
	return ctype.Hex2PayID(payID), state, ctype.Hex2Addr(bridgeAddr), found, err
}

func deleteCrossNetPay(st SqlStorage, payID ctype.PayIDType) error {
	q := `DELETE FROM crossnetpays WHERE payid = $1`
	res, err := st.Exec(q, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "deleteCrossNetPay")
}

// The "secrets" table.
func insertSecret(st SqlStorage, hash, preImage string, payID ctype.PayIDType) error {
	q := `INSERT INTO secrets (hash, preimage, payid) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, hash, preImage, ctype.PayID2Hex(payID))
	return chkExec(res, err, 1, "insertSecret")
}

func getSecret(st SqlStorage, hash string) (string, bool, error) {
	var preImage string
	q := `SELECT preimage FROM secrets WHERE hash = $1`
	err := st.QueryRow(q, hash).Scan(&preImage)
	found, err := chkQueryRow(err)
	return preImage, found, err
}

func deleteSecret(st SqlStorage, hash string) error {
	q := `DELETE FROM secrets WHERE hash = $1`
	res, err := st.Exec(q, hash)
	return chkExec(res, err, 1, "deleteSecret")
}

func deleteSecretByPayID(st SqlStorage, payID ctype.PayIDType) error {
	q := `DELETE FROM secrets WHERE payid = $1`
	_, err := st.Exec(q, ctype.PayID2Hex(payID))
	return err
}

// The "tcb" table.
func insertTcb(
	st SqlStorage,
	addr ctype.Addr,
	token *entity.TokenInfo,
	deposit *big.Int) error {
	q := `INSERT INTO tcb (addr, token, deposit) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, ctype.Addr2Hex(addr),
		utils.GetTokenAddrStr(token), deposit.String())
	return chkExec(res, err, 1, "insertTcb")
}

func getTcbDeposit(
	st SqlStorage,
	addr ctype.Addr,
	token *entity.TokenInfo) (*big.Int, bool, error) {
	var data string
	q := `SELECT deposit FROM tcb WHERE addr = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(addr),
		utils.GetTokenAddrStr(token)).Scan(&data)
	found, err := chkQueryRow(err)
	if !found {
		return nil, false, nil
	}

	deposit, ok := new(big.Int).SetString(data, 10)
	if !ok {
		return nil, false, fmt.Errorf("invalid deposit value: %s", data)
	}

	return deposit, true, nil
}

func updateTcbDeposit(st SqlStorage, addr ctype.Addr, token *entity.TokenInfo, deposit *big.Int) error {
	q := `UPDATE tcb SET deposit = $1 WHERE addr = $2 AND token = $3`
	res, err := st.Exec(q, deposit.String(), ctype.Addr2Hex(addr),
		utils.GetTokenAddrStr(token))
	return chkExec(res, err, 1, "updateTcbDeposit")
}

// The "monitor" table.
func insertMonitor(
	st SqlStorage,
	event string,
	blockNum uint64,
	blockIdx int64,
	restart bool) error {
	q := `INSERT INTO monitor (event, blocknum, blockidx, restart)
		VALUES ($1, $2, $3, $4)`
	res, err := st.Exec(q, event, blockNum, blockIdx, restart)
	return chkExec(res, err, 1, "insertMonitor")
}

func getMonitorBlock(st SqlStorage, event string) (uint64, int64, bool, error) {
	var blockNum uint64
	var blockIdx int64
	q := `SELECT blocknum, blockidx FROM monitor WHERE event = $1`
	err := st.QueryRow(q, event).Scan(&blockNum, &blockIdx)
	found, err := chkQueryRow(err)
	return blockNum, blockIdx, found, err
}

func getMonitorRestart(st SqlStorage, event string) (bool, bool, error) {
	var restart bool
	q := `SELECT restart FROM monitor WHERE event = $1`
	err := st.QueryRow(q, event).Scan(&restart)
	found, err := chkQueryRow(err)
	return restart, found, err
}

func getMonitorAddrsByEventAndRestart(st SqlStorage, eventName string, restart bool) ([]ctype.Addr, error) {
	q := `SELECT event FROM monitor WHERE event LIKE $1 AND restart = $2`

	rows, err := st.Query(q, "%-"+eventName, restart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []ctype.Addr
	for rows.Next() {
		var event string
		err = rows.Scan(&event)
		if err != nil {
			return nil, err
		}

		// split addr from event
		strs := strings.SplitN(event, "-", 2)
		if len(strs) < 2 || strs[1] != eventName {
			continue
		}
		addrs = append(addrs, ctype.Hex2Addr(strs[0]))
	}

	return addrs, nil
}

func updateMonitorBlock(
	st SqlStorage,
	event string,
	blockNum uint64,
	blockIdx int64) error {
	q := `UPDATE monitor SET blocknum = $1, blockidx = $2 WHERE event = $3`
	res, err := st.Exec(q, blockNum, blockIdx, event)
	return chkExec(res, err, 1, "updateMonitorBlock")
}

func upsertMonitorBlock(
	st SqlStorage,
	event string,
	blockNum uint64,
	blockIdx int64,
	restart bool) error {
	q := `INSERT INTO monitor (event, blocknum, blockidx, restart)
		VALUES ($1, $2, $3, $4) ON CONFLICT (event) DO UPDATE
		SET blocknum = excluded.blocknum, blockidx = excluded.blockidx`
	res, err := st.Exec(q, event, blockNum, blockIdx, restart)
	return chkExec(res, err, 1, "upsertMonitorBlock")
}

func upsertMonitorRestart(st SqlStorage, event string, restart bool) error {
	q := `INSERT INTO monitor (event, blocknum, blockidx, restart)
		VALUES ($1, $2, $3, $4) ON CONFLICT (event) DO UPDATE
		SET restart = excluded.restart`
	res, err := st.Exec(q, event, uint64(0), int64(0), restart)
	return chkExec(res, err, 1, "upsertMonitorRestart")
}

// The "routing" table.
func upsertRouting(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo,
	cid ctype.CidType) error {
	q := `INSERT INTO routing (dest, token, cid) VALUES ($1, $2, $3)
		ON CONFLICT (dest, token) DO UPDATE SET cid = excluded.cid`
	res, err := st.Exec(q, ctype.Addr2Hex(dest),
		utils.GetTokenAddrStr(token), ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "upsertRouting")
}

func getRoutingCid(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo) (ctype.CidType, bool, error) {
	var data string
	var cid ctype.CidType
	q := `SELECT cid FROM routing WHERE dest = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(dest),
		utils.GetTokenAddrStr(token)).Scan(&data)
	found, err := chkQueryRow(err)
	if found {
		cid = ctype.Hex2Cid(data)
	}
	return cid, found, err
}

// Return a nested map of token-addr -> dest-addr -> cid.
func getAllRoutingCids(st SqlStorage) (map[ctype.Addr]map[ctype.Addr]ctype.CidType, error) {
	q := `SELECT dest, token, cid FROM routing`
	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routeMap := make(map[ctype.Addr]map[ctype.Addr]ctype.CidType)
	var destStr, tokenStr, cidStr string
	for rows.Next() {
		err = rows.Scan(&destStr, &tokenStr, &cidStr)
		if err != nil {
			return nil, err
		}

		dest := ctype.Hex2Addr(destStr)
		token := ctype.Hex2Addr(tokenStr)
		cid := ctype.Hex2Cid(cidStr)
		if _, ok := routeMap[token]; !ok {
			routeMap[token] = make(map[ctype.Addr]ctype.CidType)
		}
		routeMap[token][dest] = cid
	}

	return routeMap, nil
}

func deleteRouting(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo) error {
	q := `DELETE FROM routing WHERE dest = $1 AND token = $2`
	res, err := st.Exec(q, ctype.Addr2Hex(dest),
		utils.GetTokenAddrStr(token))
	return chkExec(res, err, 1, "deleteRouting")
}

// The "edges" table.
func insertEdge(
	st SqlStorage,
	token *entity.TokenInfo,
	cid ctype.CidType,
	addr1 ctype.Addr,
	addr2 ctype.Addr) error {
	q := `INSERT INTO edges (token, cid, addr1, addr2) VALUES ($1, $2, $3, $4)`
	res, err := st.Exec(q, utils.GetTokenAddrStr(token), ctype.Cid2Hex(cid),
		ctype.Addr2Hex(addr1), ctype.Addr2Hex(addr2))
	return chkExec(res, err, 1, "insertEdge")
}

func getAllEdges(st SqlStorage) ([]*structs.Edge, error) {
	q := `SELECT token, cid, addr1, addr2 FROM edges`
	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*structs.Edge
	var token, cid, addr1, addr2 string
	for rows.Next() {
		err = rows.Scan(&token, &cid, &addr1, &addr2)
		if err != nil {
			return nil, err
		}

		e := &structs.Edge{
			P1:    ctype.Hex2Addr(addr1),
			P2:    ctype.Hex2Addr(addr2),
			Cid:   ctype.Hex2Cid(cid),
			Token: ctype.Hex2Addr(token),
		}
		edges = append(edges, e)
	}

	return edges, nil
}

func deleteEdge(st SqlStorage, cid ctype.CidType) error {
	q := `DELETE FROM edges WHERE cid = $1`
	res, err := st.Exec(q, ctype.Cid2Hex(cid))
	return chkExec(res, err, 1, "deleteEdge")
}

// The "netbridge" table
func upsertNetBridge(st SqlStorage, bridgeAddr ctype.Addr, bridgeNetId uint64) error {
	q := `INSERT INTO netbridge (bridgeaddr, bridgenetid) VALUES ($1, $2)
		ON CONFLICT (bridgeaddr) DO UPDATE SET bridgenetid = excluded.bridgenetid`
	res, err := st.Exec(q, ctype.Addr2Hex(bridgeAddr), bridgeNetId)
	return chkExec(res, err, 1, "upsertNetBridge")
}

func getNetBridge(st SqlStorage, bridgeAddr ctype.Addr) (uint64, bool, error) {
	var bridgeNetId uint64
	q := `SELECT bridgenetid FROM netbridge WHERE bridgeaddr = $1`
	err := st.QueryRow(q, ctype.Addr2Hex(bridgeAddr)).Scan(&bridgeNetId)
	found, err := chkQueryRow(err)
	return bridgeNetId, found, err
}

func deleteNetBridge(st SqlStorage, bridgeAddr ctype.Addr) error {
	q := `DELETE FROM netbridge WHERE bridgeaddr = $1`
	res, err := st.Exec(q, ctype.Addr2Hex(bridgeAddr))
	return chkExec(res, err, 1, "deleteNetBridge")
}

// The "bridgerouting" table
func upsertBridgeRouting(st SqlStorage, destNetId uint64, bridgeAddr ctype.Addr) error {
	q := `INSERT INTO bridgerouting (destnetid, bridgeaddr) VALUES ($1, $2)
		ON CONFLICT (destnetid) DO UPDATE SET bridgeaddr = excluded.bridgeaddr`
	res, err := st.Exec(q, destNetId, ctype.Addr2Hex(bridgeAddr))
	return chkExec(res, err, 1, "upsertBridgeRouting")
}

func getBridgeRouting(st SqlStorage, destNetId uint64) (ctype.Addr, uint64, bool, error) {
	var bridgeAddr string
	var bridgeNetId uint64
	q := `
		SELECT r.bridgeaddr, b.bridgenetid
		FROM bridgerouting AS r
		JOIN netbridge AS b ON r.bridgeaddr = b.bridgeaddr
		WHERE destnetid = $1
	`
	err := st.QueryRow(q, destNetId).Scan(&bridgeAddr, &bridgeNetId)
	found, err := chkQueryRow(err)
	return ctype.Hex2Addr(bridgeAddr), bridgeNetId, found, err
}

func deleteBridgeRouting(st SqlStorage, destNetId uint64) error {
	q := `DELETE FROM bridgerouting WHERE destnetid = $1`
	res, err := st.Exec(q, destNetId)
	return chkExec(res, err, 1, "deleteBridgeRouting")
}

// The "nettokens" table
func upsertNetToken(st SqlStorage, netId uint64, netToken *entity.TokenInfo, localToken *entity.TokenInfo) error {
	q := `INSERT INTO nettokens (netid, nettoken, localtoken, rate) VALUES ($1, $2, $3, $4)
		ON CONFLICT (netid, nettoken) DO UPDATE SET localtoken = excluded.localtoken`
	res, err := st.Exec(q, netId, utils.GetTokenAddrStr(netToken), utils.GetTokenAddrStr(localToken), 1.0)
	return chkExec(res, err, 1, "upsertNetToken")
}

func getLocalToken(st SqlStorage, netId uint64, netToken *entity.TokenInfo) (*entity.TokenInfo, bool, error) {
	var localTokenAddr string
	var localToken *entity.TokenInfo
	q := `SELECT localtoken FROM nettokens WHERE netid = $1 AND nettoken = $2`
	err := st.QueryRow(q, netId, utils.GetTokenAddrStr(netToken)).Scan(&localTokenAddr)
	found, err := chkQueryRow(err)
	if found {
		localToken = utils.GetTokenInfoFromAddress(ctype.Hex2Addr(localTokenAddr))
	}
	return localToken, found, err
}

func deleteNetToken(st SqlStorage, netId uint64, netToken *entity.TokenInfo) error {
	q := `DELETE FROM nettokens WHERE netid = $1 AND nettoken = $2`
	res, err := st.Exec(q, netId, utils.GetTokenAddrStr(netToken))
	return chkExec(res, err, 1, "deleteNetToken")
}

// The "peers" table.
func insertPeer(st SqlStorage, peer ctype.Addr, server string, cids []ctype.CidType) error {
	s := make([]string, 0, len(cids))
	for _, c := range cids {
		s = append(s, ctype.Cid2Hex(c))
	}

	q := `INSERT INTO peers (peer, server, activecids) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, ctype.Addr2Hex(peer), server, strings.Join(s, listSep))
	return chkExec(res, err, 1, "insertPeer")
}

func upsertPeerServer(st SqlStorage, peer ctype.Addr, server string) error {
	q := `INSERT INTO peers (peer, server, activecids) VALUES ($1, $2, $3)
		ON CONFLICT (peer) DO UPDATE SET server = excluded.server`
	res, err := st.Exec(q, ctype.Addr2Hex(peer), server, "")
	return chkExec(res, err, 1, "upsertPeerServer")
}

func updatePeerCids(st SqlStorage, peer ctype.Addr, cids []ctype.CidType) error {
	s := make([]string, 0, len(cids))
	for _, c := range cids {
		s = append(s, ctype.Cid2Hex(c))
	}

	q := `UPDATE peers SET activecids = $1 WHERE peer = $2`
	res, err := st.Exec(q, strings.Join(s, listSep), ctype.Addr2Hex(peer))
	return chkExec(res, err, 1, "updatePeerCids")
}

func updatePeerCid(st SqlStorage, peer ctype.Addr, cid ctype.CidType, add bool) error {
	cids, found, err := getPeerCids(st, peer)
	if err != nil {
		return err
	} else if !found {
		if !add {
			return nil
		}
		return insertPeer(st, peer, "", []ctype.CidType{cid})
	}

	if add {
		for _, c := range cids {
			if c == cid {
				return nil // already there
			}
		}
		cids = append(cids, cid)
	} else {
		newCids := make([]ctype.CidType, 0, len(cids))
		found = false
		for _, c := range cids {
			if c != cid {
				newCids = append(newCids, c)
				found = true
			}
		}
		if !found {
			return nil // nothing to remove
		}
		cids = newCids
	}

	return updatePeerCids(st, peer, cids)
}

func updatePeerDelegateProof(st SqlStorage, peer ctype.Addr, proof *rpc.DelegationProof) error {
	q := `UPDATE peers SET delegateproof = $1 WHERE peer = $2`
	data, err := marshal(proof)
	if err != nil {
		return err
	}
	res, err := st.Exec(q, data, ctype.Addr2Hex(peer))
	return chkExec(res, err, 1, "updatePeerDelegateProof")
}

func getPeerServer(st SqlStorage, peer ctype.Addr) (string, bool, error) {
	var server string // a host:port string
	q := `SELECT server FROM peers WHERE peer = $1`
	err := st.QueryRow(q, ctype.Addr2Hex(peer)).Scan(&server)
	found, err := chkQueryRow(err)
	return server, found, err
}

func getAllPeerServers(st SqlStorage) ([]string, error) {
	q := `SELECT DISTINCT server FROM peers`
	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var server string
		err = rows.Scan(&server)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func getPeerCids(st SqlStorage, peer ctype.Addr) ([]ctype.CidType, bool, error) {
	var data string
	q := `SELECT activecids FROM peers WHERE peer = $1`
	err := st.QueryRow(q, ctype.Addr2Hex(peer)).Scan(&data)
	found, err := chkQueryRow(err)
	if !found {
		return nil, false, err
	}
	if data == "" {
		return nil, true, nil // special case
	}

	s := strings.Split(data, listSep)
	cids := make([]ctype.CidType, 0, len(s))
	for _, c := range s {
		cids = append(cids, ctype.Hex2Cid(c))
	}

	return cids, true, nil
}

func getPeerDelegateProof(st SqlStorage, peer ctype.Addr) (*rpc.DelegationProof, bool, error) {
	var data []byte
	q := `SELECT delegateproof FROM peers WHERE peer = $1`
	err := st.QueryRow(q, ctype.Addr2Hex(peer)).Scan(&data)
	found, err := chkQueryRow(err)
	if found && data != nil {
		var proof rpc.DelegationProof
		err = unmarshal(data, &proof)
		return &proof, found, err
	}
	return nil, found, err
}

// The "desttokens" table.
func insertDestToken(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo,
	osps []ctype.Addr,
	chanBlockNum uint64) error {
	s := make([]string, 0, len(osps))
	for _, o := range osps {
		s = append(s, ctype.Addr2Hex(o))
	}

	q := `INSERT INTO desttokens (dest, token, osps, openchanblknum)
		VALUES ($1, $2, $3, $4)`
	res, err := st.Exec(q, ctype.Addr2Hex(dest), utils.GetTokenAddrStr(token),
		strings.Join(s, listSep), chanBlockNum)
	return chkExec(res, err, 1, "insertDestToken")
}

func getDestTokenOpenChanBlkNum(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo) (uint64, bool, error) {
	var blkNum uint64
	q := `SELECT openchanblknum FROM desttokens WHERE dest = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(dest),
		utils.GetTokenAddrStr(token)).Scan(&blkNum)
	found, err := chkQueryRow(err)
	return blkNum, found, err
}

func upsertDestTokenOpenChanBlkNum(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo,
	chanBlockNum uint64) error {
	q := `INSERT INTO desttokens (dest, token, osps, openchanblknum)
		VALUES ($1, $2, $3, $4) ON CONFLICT (dest, token)
		DO UPDATE SET openchanblknum = excluded.openchanblknum`
	res, err := st.Exec(q, ctype.Addr2Hex(dest),
		utils.GetTokenAddrStr(token), "", chanBlockNum)
	return chkExec(res, err, 1, "upsertDestTokenOpenChanBlkNum")
}

func updateDestTokenOsps(
	st SqlStorage,
	dest ctype.Addr,
	token *entity.TokenInfo,
	osps []ctype.Addr) error {
	s := make([]string, 0, len(osps))
	for _, o := range osps {
		s = append(s, ctype.Addr2Hex(o))
	}

	q := `UPDATE desttokens SET osps = $1 WHERE dest = $2 AND token = $3`
	res, err := st.Exec(q, strings.Join(s, listSep),
		ctype.Addr2Hex(dest), utils.GetTokenAddrStr(token))
	return chkExec(res, err, 1, "updateDestTokenOsps")
}

func deleteDestToken(st SqlStorage, dest ctype.Addr, token *entity.TokenInfo) error {
	q := `DELETE FROM desttokens WHERE dest = $1 AND token = $2`
	res, err := st.Exec(q, ctype.Addr2Hex(dest), utils.GetTokenAddrStr(token))
	return chkExec(res, err, 1, "deleteDestToken")
}

func getAllDestTokenOsps(st SqlStorage) (map[ctype.Addr]map[ctype.Addr]map[ctype.Addr]bool, error) {
	q := `SELECT dest, token, osps FROM desttokens`
	rows, err := st.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dtoMap := make(map[ctype.Addr]map[ctype.Addr]map[ctype.Addr]bool)
	var dest, token, osps string
	for rows.Next() {
		err = rows.Scan(&dest, &token, &osps)
		if err != nil {
			return nil, err
		}

		destAddr := ctype.Hex2Addr(dest)
		tokenAddr := ctype.Hex2Addr(token)

		if _, ok := dtoMap[tokenAddr]; !ok {
			dtoMap[tokenAddr] = make(map[ctype.Addr]map[ctype.Addr]bool)
		}
		ospMap := make(map[ctype.Addr]bool)
		dtoMap[tokenAddr][destAddr] = ospMap

		if osps != "" {
			s := strings.Split(osps, listSep)
			for _, o := range s {
				ospMap[ctype.Hex2Addr(o)] = true
			}
		}
	}

	return dtoMap, nil
}

func getDestTokenOsps(st SqlStorage, dest ctype.Addr, token *entity.TokenInfo) ([]ctype.Addr, error) {
	var data string
	q := `SELECT osps FROM desttokens WHERE dest = $1 AND token = $2`
	err := st.QueryRow(q, ctype.Addr2Hex(dest), utils.GetTokenAddrStr(token)).Scan(&data)
	found, err := chkQueryRow(err)
	if !found {
		return nil, err
	}
	if data == "" {
		return nil, nil // special case
	}

	s := strings.Split(data, listSep)
	osps := make([]ctype.Addr, 0, len(s))
	for _, o := range s {
		osps = append(osps, ctype.Hex2Addr(o))
	}

	return osps, nil
}

// The "migration" table.
func upsertChanMigration(
	st SqlStorage,
	cid ctype.CidType,
	toLedger ctype.Addr,
	deadline uint64,
	state int,
	onchainReq *chain.ChannelMigrationRequest) error {
	q := `INSERT INTO chanmigration (cid, toledger, deadline, onchainreq, state, ts) 
		VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (cid, toledger) DO UPDATE
		SET deadline = EXCLUDED.deadline, onchainreq = EXCLUDED.onchainreq,
		state = EXCLUDED.state, ts = EXCLUDED.ts`

	reqBytes, err := proto.Marshal(onchainReq)
	if err != nil {
		return err
	}

	ts := now()
	res, err := st.Exec(q, ctype.Cid2Hex(cid), ctype.Addr2Hex(toLedger), deadline,
		reqBytes, state, ts)

	return chkExec(res, err, 1, "upsertChanMigration")
}

func deleteChanMigration(st SqlStorage, cid ctype.CidType) error {
	q := `DELETE FROM chanmigration WHERE cid = $1`
	_, err := st.Exec(q, ctype.Cid2Hex(cid))
	// don't check db result here, might delete more than 1 row
	return err
}

func updateChanMigrationState(st SqlStorage, cid ctype.CidType, toLedger ctype.Addr, state int) error {
	q := `UPDATE chanmigration SET state = $1 WHERE cid = $2 AND toledger = $3`
	res, err := st.Exec(q, state, ctype.Cid2Hex(cid), ctype.Addr2Hex(toLedger))
	return chkExec(res, err, 1, "updateChanMigrationState")
}

func getChanMigration(st SqlStorage, cid ctype.CidType, toLedger ctype.Addr) (uint64, int, []byte, bool, error) {
	q := `SELECT deadline, state, onchainreq FROM chanmigration WHERE cid = $1 AND toledger = $2`

	var onchainReq []byte
	var state int
	var deadline uint64
	err := st.QueryRow(q, ctype.Cid2Hex(cid), ctype.Addr2Hex(toLedger)).Scan(&deadline, &state, &onchainReq)
	found, err := chkQueryRow(err)

	return deadline, state, onchainReq, found, err
}

func getChanMigrationReqByLedgerAndStateWithLimit(st SqlStorage, toLedger ctype.Addr, state, limit int) (map[ctype.CidType][]byte, map[ctype.CidType]uint64, error) {
	q := `SELECT cid, deadline, onchainreq FROM chanmigration WHERE toledger = $1 AND state = $2 ORDER BY deadline LIMIT $3`

	rows, err := st.Query(q, ctype.Addr2Hex(toLedger), state, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	reqs := make(map[ctype.CidType][]byte)
	deadlines := make(map[ctype.CidType]uint64)
	for rows.Next() {
		var cidStr string
		var onchainReq []byte
		var deadline uint64
		if err := rows.Scan(&cidStr, &deadline, &onchainReq); err != nil {
			return nil, nil, err
		}

		cid := ctype.Hex2Cid(cidStr)
		reqs[cid] = onchainReq
		deadlines[cid] = deadline
	}

	return reqs, deadlines, nil
}

// The "deposit" table
func insertDeposit(
	st SqlStorage,
	uuid string,
	cid ctype.CidType,
	topeer bool,
	amount *big.Int,
	refill bool,
	deadline time.Time,
	state int,
	txhash string,
	errmsg string) error {
	q := `INSERT INTO deposit (uuid, cid, topeer, amount, refill, deadline, state, txhash, errmsg) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	res, err := st.Exec(q, uuid, ctype.Cid2Hex(cid), topeer, amount.String(), refill, deadline, state, txhash, errmsg)
	return chkExec(res, err, 1, "insertDeposit")
}

func getDeposit(st SqlStorage, uuid string) (
	ctype.CidType, bool, *big.Int, bool, time.Time, int, string, string, bool, error) {
	var cid, amount, deadlineStr, txhash, errmsg string
	var state int
	var topeer, refill bool
	q := `SELECT cid, topeer, amount, refill, deadline, state, txhash, errmsg FROM deposit WHERE uuid = $1`
	err := st.QueryRow(q, uuid).Scan(&cid, &topeer, &amount, &refill, &deadlineStr, &state, &txhash, &errmsg)
	found, err := chkQueryRow(err)
	var deadline time.Time
	if found && err == nil {
		deadline, err = str2Time(deadlineStr)
	}
	amtInt := new(big.Int)
	amtInt.SetString(amount, 10)
	return ctype.Hex2Cid(cid), topeer, amtInt, refill, deadline, state, txhash, errmsg, found, err
}

func getDepositJob(st SqlStorage, uuid string) (*structs.DepositJob, bool, error) {
	cid, topeer, amount, refill, deadline, state, txhash, errmsg, found, err := getDeposit(st, uuid)
	if err != nil || !found {
		return nil, found, err
	}
	job := &structs.DepositJob{
		UUID:     uuid,
		Cid:      cid,
		ToPeer:   topeer,
		Amount:   amount,
		Refill:   refill,
		Deadline: deadline,
		State:    state,
		TxHash:   txhash,
		ErrMsg:   errmsg,
	}
	return job, found, err
}

func getDepositState(st SqlStorage, uuid string) (int, string, bool, error) {
	var state int
	var errmsg string
	q := `SELECT state, errmsg FROM deposit WHERE uuid = $1`
	err := st.QueryRow(q, uuid).Scan(&state, &errmsg)
	found, err := chkQueryRow(err)
	return state, errmsg, found, err
}

func getDepositByTxHash(st SqlStorage, txhash string) (
	string, ctype.CidType, bool, *big.Int, bool, time.Time, int, string, bool, error) {
	var uuid, cid, amount, deadlineStr, errmsg string
	var state int
	var topeer, refill bool
	q := `SELECT uuid, cid, topeer, amount, refill, deadline, state, errmsg FROM deposit WHERE txhash = $1`
	err := st.QueryRow(q, txhash).Scan(&uuid, &cid, &topeer, &amount, &refill, &deadlineStr, &state, &errmsg)
	found, err := chkQueryRow(err)
	var deadline time.Time
	if found && err == nil {
		deadline, err = str2Time(deadlineStr)
	}
	amtInt := new(big.Int)
	amtInt.SetString(amount, 10)
	return uuid, ctype.Hex2Cid(cid), topeer, amtInt, refill, deadline, state, errmsg, found, err
}

func getDepositJobByTxHash(st SqlStorage, txhash string) (*structs.DepositJob, bool, error) {
	uuid, cid, topeer, amount, refill, deadline, state, errmsg, found, err := getDepositByTxHash(st, txhash)
	if err != nil || !found {
		return nil, found, err
	}
	job := &structs.DepositJob{
		UUID:     uuid,
		Cid:      cid,
		ToPeer:   topeer,
		Amount:   amount,
		Refill:   refill,
		Deadline: deadline,
		State:    state,
		TxHash:   txhash,
		ErrMsg:   errmsg,
	}
	return job, found, err
}

func getAllDepositJobsByState(st SqlStorage, state int) ([]*structs.DepositJob, error) {
	q := `SELECT uuid, cid, topeer, amount, refill, deadline, txhash, errmsg FROM deposit WHERE state = $1`
	rows, err := st.Query(q, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uuid, cid, amount, deadlineStr, txhash, errmsg string
	var deadline time.Time
	var topeer, refill bool

	var jobs []*structs.DepositJob
	for rows.Next() {
		err := rows.Scan(&uuid, &cid, &topeer, &amount, &refill, &deadlineStr, &txhash, &errmsg)
		if err != nil {
			return nil, err
		}
		deadline, err = str2Time(deadlineStr)
		if err != nil {
			return nil, err
		}
		amtInt := new(big.Int)
		amtInt.SetString(amount, 10)
		job := &structs.DepositJob{
			UUID:     uuid,
			Cid:      ctype.Hex2Cid(cid),
			ToPeer:   topeer,
			Amount:   amtInt,
			Refill:   refill,
			Deadline: deadline,
			State:    state,
			TxHash:   txhash,
			ErrMsg:   errmsg,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func getAllDepositJobsByCid(st SqlStorage, cid ctype.CidType) ([]*structs.DepositJob, error) {
	q := `SELECT uuid, topeer, amount, refill, deadline, state, txhash, errmsg FROM deposit WHERE cid = $1`
	rows, err := st.Query(q, ctype.Cid2Hex(cid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uuid, amount, deadlineStr, txhash, errmsg string
	var deadline time.Time
	var state int
	var topeer, refill bool

	var jobs []*structs.DepositJob
	for rows.Next() {
		err := rows.Scan(&uuid, &topeer, &amount, &refill, &deadlineStr, &state, &txhash, &errmsg)
		if err != nil {
			return nil, err
		}
		deadline, err = str2Time(deadlineStr)
		if err != nil {
			return nil, err
		}
		amtInt := new(big.Int)
		amtInt.SetString(amount, 10)
		job := &structs.DepositJob{
			UUID:     uuid,
			Cid:      cid,
			ToPeer:   topeer,
			Amount:   amtInt,
			Refill:   refill,
			Deadline: deadline,
			State:    state,
			TxHash:   txhash,
			ErrMsg:   errmsg,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func getAllRunningDepositJobs(st SqlStorage) ([]*structs.DepositJob, error) {
	q := `SELECT uuid, cid, topeer, amount, refill, deadline, state, txhash, errmsg FROM deposit
		WHERE state = $1 OR state = $2`
	rows, err := st.Query(q, structs.DepositState_APPROVING_ERC20, structs.DepositState_TX_SUBMITTED)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uuid, cid, amount, deadlineStr, txhash, errmsg string
	var deadline time.Time
	var state int
	var topeer, refill bool

	var jobs []*structs.DepositJob
	for rows.Next() {
		err := rows.Scan(&uuid, &cid, &topeer, &amount, &refill, &deadlineStr, &state, &txhash, &errmsg)
		if err != nil {
			return nil, err
		}
		deadline, err = str2Time(deadlineStr)
		if err != nil {
			return nil, err
		}
		amtInt := new(big.Int)
		amtInt.SetString(amount, 10)
		job := &structs.DepositJob{
			UUID:     uuid,
			Cid:      ctype.Hex2Cid(cid),
			ToPeer:   topeer,
			Amount:   amtInt,
			Refill:   refill,
			Deadline: deadline,
			State:    state,
			TxHash:   txhash,
			ErrMsg:   errmsg,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func getAllSubmittedDepositTxHashes(st SqlStorage) ([]string, error) {
	q := `SELECT DISTINCT txhash FROM deposit WHERE state = $1`
	rows, err := st.Query(q, structs.DepositState_TX_SUBMITTED)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txHashes []string
	for rows.Next() {
		var txHash string
		err = rows.Scan(&txHash)
		if err != nil {
			return nil, err
		}
		txHashes = append(txHashes, txHash)
	}
	return txHashes, nil
}

func hasDepositRefillPending(st SqlStorage, cid ctype.CidType) (bool, error) {
	var data int
	q := `SELECT 1 FROM deposit WHERE cid = $1 AND refill = $2 AND (state = $3 OR state = $4 OR state = $5)`
	err := st.QueryRow(q, ctype.Cid2Hex(cid), true,
		structs.DepositState_QUEUED, structs.DepositState_TX_SUBMITTING, structs.DepositState_TX_SUBMITTED).Scan(&data)
	return chkQueryRow(err)
}

func hasDepositTxHash(st SqlStorage, txhash string) (bool, error) {
	var data int
	q := `SELECT 1 FROM deposit WHERE txhash = $1`
	err := st.QueryRow(q, txhash).Scan(&data)
	return chkQueryRow(err)
}

func updateDepositStateAndTxHash(st SqlStorage, uuid string, state int, txhash string) error {
	q := `UPDATE deposit SET state = $1, txhash = $2 WHERE uuid = $3`
	res, err := st.Exec(q, state, txhash, uuid)
	return chkExec(res, err, 1, "updateDepositStateAndTxHash")
}

func updateDepositsStateAndTxHash(st SqlStorage, uuids []string, state int, txhash string) error {
	ulen := len(uuids)
	if ulen == 0 {
		return nil
	}
	q := fmt.Sprintf("UPDATE deposit SET state = $1, txhash = $2 WHERE %s", inClause("uuid", ulen, 3))
	var args []interface{}
	args = append(args, state)
	args = append(args, txhash)
	for _, uuid := range uuids {
		args = append(args, uuid)
	}
	res, err := st.Exec(q, args...)
	return chkExec(res, err, int64(ulen), "updateDepositsStateAndTxHash")
}

func updateDepositErrMsg(st SqlStorage, uuid, errmsg string) error {
	q := `UPDATE deposit SET state = $1, errmsg = $2 WHERE uuid = $3`
	res, err := st.Exec(q, structs.DepositState_FAILED, errmsg, uuid)
	return chkExec(res, err, 1, "updateDepositErrMsg")
}

func updateDepositsErrMsg(st SqlStorage, uuids []string, errmsg string) error {
	ulen := len(uuids)
	if ulen == 0 {
		return nil
	}
	q := fmt.Sprintf("UPDATE deposit SET state = $1, errmsg = $2 WHERE %s", inClause("uuid", ulen, 3))
	var args []interface{}
	args = append(args, structs.DepositState_FAILED)
	args = append(args, errmsg)
	for _, uuid := range uuids {
		args = append(args, uuid)
	}
	res, err := st.Exec(q, args...)
	return chkExec(res, err, int64(ulen), "updateDepositsErrMsg")
}

func updateDepositStatesByTxHashAndCid(st SqlStorage, txhash string, cid ctype.CidType, state int) error {
	q := `UPDATE deposit SET state = $1 WHERE txhash = $2 AND cid = $3`
	_, err := st.Exec(q, structs.DepositState_SUCCEEDED, txhash, ctype.Cid2Hex(cid))
	return err
}

func updateDepositErrMsgByTxHash(st SqlStorage, txhash, errmsg string) error {
	q := `UPDATE deposit SET state = $1, errmsg = $2 WHERE txhash = $3`
	_, err := st.Exec(q, structs.DepositState_FAILED, errmsg, txhash)
	return err
}

func deleteDeposit(st SqlStorage, uuid string) error {
	q := `DELETE FROM deposit WHERE uuid = $1`
	res, err := st.Exec(q, uuid)
	return chkExec(res, err, 1, "deleteDeposit")
}

// The "lease" table
func insertLease(st SqlStorage, id, owner string) error {
	q := `INSERT INTO lease (id, owner, updatets) VALUES ($1, $2, $3)`
	res, err := st.Exec(q, id, owner, now())
	return chkExec(res, err, 1, "insertLease")
}

func updateLeaseOwner(st SqlStorage, id, owner string) error {
	q := `UPDATE lease SET owner = $1, updatets = $2 WHERE id = $3`
	res, err := st.Exec(q, owner, now(), id)
	return chkExec(res, err, 1, "updateLeaseOwner")
}

func updateLeaseTimestamp(st SqlStorage, id, owner string) error {
	q := `UPDATE lease SET updatets = $1 WHERE id = $2 AND owner = $3`
	res, err := st.Exec(q, now(), id, owner)
	return chkExec(res, err, 1, "updateLeaseTimestamp")
}

func getLease(st SqlStorage, id string) (string, time.Time, bool, error) {
	var owner, updateTsStr string
	q := `SELECT owner, updatets FROM lease WHERE id = $1`
	err := st.QueryRow(q, id).Scan(&owner, &updateTsStr)
	found, err := chkQueryRow(err)
	var updateTs time.Time
	if found && err == nil {
		updateTs, err = str2Time(updateTsStr)
	}
	return owner, updateTs, found, err
}

func getLeaseOwner(st SqlStorage, id string) (string, bool, error) {
	var owner string
	q := `SELECT owner FROM lease WHERE id = $1`
	err := st.QueryRow(q, id).Scan(&owner)
	found, err := chkQueryRow(err)
	return owner, found, err
}

func deleteLeaseOwner(st SqlStorage, id, owner string) error {
	q := `DELETE FROM lease WHERE id = $1 AND owner = $2`
	res, err := st.Exec(q, id, owner)
	return chkExec(res, err, 1, "deleteLeaseOwner")
}

func deleteLease(st SqlStorage, id string) error {
	q := `DELETE FROM lease WHERE id = $1`
	res, err := st.Exec(q, id)
	return chkExec(res, err, 1, "deleteLease")
}
