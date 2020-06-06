// Copyright 2018-2020 Celer Network

package watcher

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
	ethereum "github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	perBlock = 3         // Number of log events per block
	bigIndex = 999999999 // Testing a visible "Removed" log
	errProb  = 10        // Random on-chain error probability 1/N
	stDriver = "sqlite3"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Return a temporary store file for testing without creating the file.
func tempStoreFile() string {
	user, _ := user.Current()
	ts := time.Now().UnixNano() / 1000
	dir := fmt.Sprintf("sql-store-%s-%d.db", user.Username, ts)
	return filepath.Join(os.TempDir(), dir)
}

func fakeLog(pos int) types.Log {
	log := types.Log{
		Data:        []byte(fmt.Sprintf("dummy data %d", pos)),
		BlockNumber: uint64(pos / perBlock),
		Index:       uint(pos % perBlock),
	}
	return log
}

// Fake watch Ethereum client to fetch event logs.
type fakeClient struct {
	randErr  bool
	blkSleep time.Duration
	quit     chan bool
	blkChan  chan int64
	blkNum   int64
	noLog    bool // won't return any fakeLog
}

// NewFakeClient creates a fake watch client that return increasing
// block numbers and the requested event logs.
func NewFakeClient(blkSleep time.Duration, randErr bool) *fakeClient {
	fc := &fakeClient{
		randErr:  randErr,
		blkNum:   0,
		blkSleep: blkSleep,
		quit:     make(chan bool),
		blkChan:  make(chan int64),
	}
	go fc.blkTick()
	return fc
}

func (fc *fakeClient) fakeError() bool {
	return fc.randErr && rand.Intn(errProb) == 0
}

func (fc *fakeClient) blkTick() {
	ticker := time.NewTicker(fc.blkSleep)
	defer ticker.Stop()

	for {
		select {
		case <-fc.quit:
			return

		case <-ticker.C:
			fc.blkNum++

		case fc.blkChan <- fc.blkNum:
		}
	}
}

func (fc *fakeClient) Close() {
	close(fc.quit)
}

func (fc *fakeClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if fc.fakeError() {
		return nil, fmt.Errorf("Random fake error")
	}
	blkNum := <-fc.blkChan
	head := &types.Header{
		Number: big.NewInt(blkNum),
	}
	return head, nil
}

func (fc *fakeClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	var logs []types.Log

	if fc.fakeError() {
		return nil, fmt.Errorf("Random fake error")
	}

	if q.BlockHash != nil {
		hash := q.BlockHash.Big().Int64()
		if hash == 404 {
			return logs, nil // empty
		}
		return nil, fmt.Errorf("Invalid BlockHash: %v", q.BlockHash)
	}

	if q.FromBlock == nil {
		return nil, fmt.Errorf("FromBlock not specified")
	}

	blkNum := <-fc.blkChan

	from := q.FromBlock.Int64()
	if from > blkNum {
		return nil, fmt.Errorf("FromBlock %d > blk num %d", from, blkNum)
	}

	to := blkNum
	if q.ToBlock != nil {
		to = q.ToBlock.Int64()
		if to > blkNum {
			return nil, fmt.Errorf("ToBlock %d > blk num %d", to, blkNum)
		}
	}

	if from > to {
		return nil, fmt.Errorf("FromBlock %d > ToBlock %d", from, to)
	}

	if !fc.noLog {
		start := int(from) * perBlock
		end := int(to+1) * perBlock
		for i := start; i < end; i++ {
			logs = append(logs, fakeLog(i))
		}
	}
	return logs, nil
}

func consumeSome(t *testing.T, w *Watch, num, start int) {
	for i := 0; i < num; i++ {
		log, err := w.Next()
		if err != nil {
			t.Fatalf("Cannot get next log %d: %v", start, err)
		}

		exp := fakeLog(start)
		if !reflect.DeepEqual(log, exp) {
			t.Errorf("Got wrong log data %d: %v != %v", start, log, exp)
		}

		err = w.Ack()
		if err != nil {
			t.Errorf("ACK failed %d: %v", start, err)
		}
		start++
	}
}

func TestWatcher(t *testing.T) {
	stFile := tempStoreFile()
	defer os.Remove(stFile)

	st, _ := storage.NewKVStoreSQL(stDriver, stFile)
	defer st.Close()
	dal := storage.NewDAL(st)

	polling := 20 // msec
	blkSleep := time.Duration(polling) * time.Millisecond
	client := NewFakeClient(blkSleep, false)
	defer client.Close()

	ws := makeWatchService(client, dal, uint64(polling))
	if ws == nil {
		t.Fatalf("Cannot create watch service")
	}
	defer ws.Close()

	bkn1 := ws.GetBlockNumber()
	time.Sleep(2 * blkSleep)
	bkn2 := ws.GetBlockNumber()

	if bkn2 <= bkn1 {
		t.Errorf("Block number did not increase: %d vs %d", bkn1, bkn2)
	}

	time.Sleep(2 * blkSleep)
	blockNumber := ws.GetCurrentBlockNumber()
	if blockNumber == nil {
		t.Errorf("Block number (big.Int) returned nil")
	}
	if blockNumber.Uint64() <= bkn2 {
		t.Errorf("Block number (big.Int) did not increase: %d vs %d", bkn2, blockNumber.Uint64())
	}

	query := ethereum.FilterQuery{}

	w, err := ws.NewWatch("foo", query, 2, 1, false)
	if err != nil {
		t.Fatalf("Cannot create watcher: %v", err)
	}

	consumeSome(t, w, 10, 0)

	// Manually close the watcher (not defer) and give it some time
	// before exiting the test to increase code coverage of the
	// watcher shutdown code (goroutines exiting).
	w.Close()
	// test noLog case
	client.noLog = true
	w2, err := ws.NewWatch("foo", query, 2, 1, true) // reset fromBlock to 0
	if w2.fromBlock != 0 {
		t.Error("fromBlock isn't 0")
	}
	time.Sleep(2 * blkSleep) // so fetchLogEvents has run
	if w2.fromBlock == 0 {
		t.Error("fromBlock is still 0")
	}
	w2.Close()
	time.Sleep(100 * time.Millisecond)
}

func TestBadWatcher(t *testing.T) {
	stFile := tempStoreFile()
	defer os.Remove(stFile)

	st, _ := storage.NewKVStoreSQL(stDriver, stFile)
	defer st.Close()
	dal := storage.NewDAL(st)

	ws := NewWatchService(nil, dal, 1)
	if ws != nil {
		ws.Close()
		t.Errorf("Watch service did not error on NIL client")
	}

	blkSleep := 20 * time.Millisecond
	client := NewFakeClient(blkSleep, true)
	defer client.Close()

	ws = NewWatchService(client, nil, 1)
	if ws != nil {
		ws.Close()
		t.Errorf("Watch service did not error on NIL DAL")
	}

	ws = NewWatchService(client, dal, 0)
	if ws != nil {
		ws.Close()
		t.Errorf("Watch service did not error on zero polling")
	}

	ws = NewWatchService(client, dal, 5)
	if ws == nil {
		t.Fatalf("Cannot create watch service")
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(3),
	}

	w, err := ws.NewWatch("", query, 0, 1, false)
	if err == nil {
		w.Close()
		t.Errorf("Watcher did not error on empty name")
	}

	w, err = ws.NewWatch("foo", query, 0, 1, false)
	if err != nil {
		t.Errorf("Cannot create watcher: %v", err)
	}

	w2, err := ws.NewWatch("foo", query, 0, 1, false)
	if err == nil {
		w2.Close()
		t.Errorf("Duplicate watcher did not fail")
	}

	if _, err = w.Next(); err != nil {
		t.Errorf("Cannot get 1st log: %v", err)
	}

	if _, err = w.Next(); err == nil {
		t.Errorf("Get next log without ACK did not fail")
	}

	if err = w.Ack(); err != nil {
		t.Errorf("Cannot ACK: %v", err)
	}

	if err = w.Ack(); err == nil {
		t.Errorf("Double ACK did not fail")
	}

	if _, err = w.Next(); err != nil {
		t.Errorf("Cannot get 2nd log: %v", err)
	}

	w.Close()

	// Verify lenient ACK-after-Close.
	if err = w.Ack(); err != nil {
		t.Errorf("Cannot make one last ACK after close: %v", err)
	}

	if err = w.Ack(); err == nil {
		t.Errorf("Another ACK after close did not fail")
	}

	if _, err = w.Next(); err == nil {
		t.Errorf("Get after close did not fail")
	}

	// A valid NewWatch() should fail after the watch service is closed.
	ws.Close()

	w, err = ws.NewWatch("bar", query, 0, 1, false)
	if err == nil {
		t.Errorf("Created watcher after service closed")
		w.Close()
	} else if err != ErrWatchServiceClosed {
		t.Errorf("New watcher failed with unexpected error: %v", err)
	}
}

func TestWatcherRestart(t *testing.T) {
	stFile := tempStoreFile()
	defer os.Remove(stFile)

	st, _ := storage.NewKVStoreSQL(stDriver, stFile)
	dal := storage.NewDAL(st)

	polling := 20 // msec
	blkSleep := time.Duration(polling) * time.Millisecond
	client := NewFakeClient(blkSleep, true)

	ws := makeWatchService(client, dal, uint64(polling))

	query := ethereum.FilterQuery{}
	w, err := ws.NewWatch("foo", query, 2, 1, false)
	if err != nil {
		t.Fatalf("Cannot create watcher: %v", err)
	}

	consumeSome(t, w, 4, 0) // fetch 4 log events

	// App termination.
	w.Close()
	ws.Close()
	client.Close()
	st.Close()

	// App restart.
	st, _ = storage.NewKVStoreSQL(stDriver, stFile)
	dal = storage.NewDAL(st)
	client = NewFakeClient(blkSleep, true)
	ws = makeWatchService(client, dal, uint64(polling))

	w, err = ws.NewWatch("foo", query, 2, 1, false)
	if err != nil {
		t.Fatalf("Cannot create watcher: %v", err)
	}

	consumeSome(t, w, 6, 4) // fetch the next 6 log events

	// App termination again.
	w.Close()
	ws.Close()
	client.Close()
	st.Close()

	// App restart, this time ignore persistence and reset the subscription.
	st, _ = storage.NewKVStoreSQL(stDriver, stFile)
	defer st.Close()
	dal = storage.NewDAL(st)
	client = NewFakeClient(blkSleep, true)
	ws = makeWatchService(client, dal, uint64(polling))

	w, err = ws.NewWatch("foo", query, 2, 1, true) // reset subscription
	if err != nil {
		t.Fatalf("Cannot create watcher: %v", err)
	}

	consumeSome(t, w, 10, 0)
}

func TestWatcherServiceClose(t *testing.T) {
	stFile := tempStoreFile()
	defer os.Remove(stFile)

	st, _ := storage.NewKVStoreSQL(stDriver, stFile)
	dal := storage.NewDAL(st)

	polling := 20 // msec
	blkSleep := time.Duration(polling) * time.Millisecond
	client := NewFakeClient(blkSleep, true)
	defer client.Close()

	ws := makeWatchService(client, dal, uint64(polling))

	hash := ethcommon.BigToHash(big.NewInt(404))
	query := ethereum.FilterQuery{
		BlockHash: &hash,
	}
	w, _ := ws.NewWatch("foo", query, 2, 1, false)

	go func() {
		// Block reading the next event log.
		_, err := w.Next()
		if err == nil {
			t.Errorf("Next() did not fail at service termination")
		}
	}()

	// App top-down termination, closing the watch service.
	time.Sleep(100 * time.Millisecond)
	ws.Close()
	time.Sleep(100 * time.Millisecond)
	st.Close()
}

func TestMakeFilterQuery(t *testing.T) {
	peer := "5963e46cf9f9700e70d4d1bc09210711ab4a20b4"
	abi := ledger.CelerLedgerABI

	stFile := tempStoreFile()
	defer os.Remove(stFile)

	st, _ := storage.NewKVStoreSQL(stDriver, stFile)
	defer st.Close()
	dal := storage.NewDAL(st)

	polling := 20 // msec
	blkSleep := time.Duration(polling) * time.Millisecond
	client := NewFakeClient(blkSleep, true)
	defer client.Close()

	ws := NewWatchService(client, dal, 1)
	defer ws.Close()

	addr := ctype.Hex2Addr(peer)
	name := "OpenChannel"
	startBlock := big.NewInt(1234)

	q, err := ws.MakeFilterQuery(addr, abi, name, startBlock)
	if err != nil {
		t.Errorf("Cannot create FilterQuery for %s: %v", name, err)
	}
	expAddr := []ctype.Addr{addr}
	if q.FromBlock.Cmp(startBlock) != 0 {
		t.Errorf("Wrong start block: %v != %v", q.FromBlock, startBlock)
	}
	if !reflect.DeepEqual(q.Addresses, expAddr) {
		t.Errorf("Wrong query addresses: %v != %v", q.Addresses, expAddr)
	}
	if len(q.Topics) != 1 || len(q.Topics[0]) != 1 {
		t.Errorf("Wrong query topics: %v", q.Topics)
	}

	name = "foobar"
	q, err = ws.MakeFilterQuery(addr, abi, name, nil)
	if err == nil {
		t.Errorf("Created FilterQuery for bad event name %s: %v", name, q)
	}

	name = "foobar"
	abi = "some [ bad { JSON"
	q, err = ws.MakeFilterQuery(addr, abi, name, nil)
	if err == nil {
		t.Errorf("Created FilterQuery for bad ABI: %v", q)
	}
}
