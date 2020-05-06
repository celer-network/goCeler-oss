// Copyright 2018-2020 Celer Network

package monitor_test

import (
	"container/heap"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/watcher"
	"github.com/celer-network/goutils/log"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Fake Ethereum client.
type fakeClient struct {
	quit    chan bool
	blkChan chan int64
}

func NewFakeClient() *fakeClient {
	fc := &fakeClient{
		quit:    make(chan bool),
		blkChan: make(chan int64),
	}
	go fc.blkTick()
	return fc
}

func (fc *fakeClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	blkNum := <-fc.blkChan
	head := &types.Header{
		Number: big.NewInt(blkNum),
	}
	return head, nil
}

func (fc *fakeClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}

func (fc *fakeClient) Close() {
	close(fc.quit)
}

func (fc *fakeClient) blkTick() {
	blkNum := int64(1)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-fc.quit:
			return

		case <-ticker.C:
			blkNum++

		case fc.blkChan <- blkNum:
		}
	}
}

// Creates a watch service with a KVStore and a fake Ethereum client.
// Returns the watch service, its fake client, and the KVStore directory.
func watchService() (*watcher.WatchService, *fakeClient, string, error) {
	dir, err := ioutil.TempDir("", "store-")
	if err != nil {
		return nil, nil, "", err
	}
	stFile := filepath.Join(dir, "sql-store.db")

	st, err := storage.NewKVStoreSQL("sqlite3", stFile)
	if err != nil {
		return nil, nil, "", err
	}
	dal := storage.NewDAL(st)

	client := NewFakeClient()

	ws := watcher.NewWatchService(client, dal, 1)
	if ws == nil {
		return nil, nil, "", fmt.Errorf("cannot create watch service")
	}
	return ws, client, dir, nil
}

var deadline_cb_count uint64

func deadline_callback() {
	n := atomic.AddUint64(&deadline_cb_count, 1)
	log.Infof("Call deadline callback: %d", n)
}

func TestDeadline(t *testing.T) {
	deadline_cb_count = 0

	ws, client, storeDir, err := watchService()
	if err != nil {
		t.Fatalf("fail to create watch service: %s", err)
	}
	defer client.Close()
	defer os.RemoveAll(storeDir)
	defer ws.Close()

	ms := monitor.NewService(ws, 0 /* blockDelay */, true /* enabled */, "" /* rpcAddr */)
	ms.Init()

	// Note: the block number producer is ticking every 10ms.
	// The code below adds 4 deadline events then removes one
	// before the block number reaches the one removed, thus
	// only 3 callbacks should trigger.

	e1 := monitor.Deadline{
		BlockNum: big.NewInt(100),
		Callback: deadline_callback,
	}
	id1 := ms.RegisterDeadline(e1)
	log.Infof("Register callback %d\n", id1)

	e2 := monitor.Deadline{
		BlockNum: big.NewInt(90),
		Callback: deadline_callback,
	}
	id2 := ms.RegisterDeadline(e2)
	log.Infof("Register callback %d\n", id2)

	e3 := monitor.Deadline{
		BlockNum: big.NewInt(1000000),
		Callback: deadline_callback,
	}
	id3 := ms.RegisterDeadline(e3)
	log.Infof("Register callback %d\n", id3)

	e4 := monitor.Deadline{
		BlockNum: big.NewInt(80),
		Callback: deadline_callback,
	}
	id4 := ms.RegisterDeadline(e4)
	log.Infof("Register callback %d\n", id4)

	ms.RemoveDeadline(id3)

	time.Sleep(3 * time.Second)

	count := atomic.LoadUint64(&deadline_cb_count)
	if count != 3 {
		t.Errorf("wrong count of deadline callback: %d != 3", count)
	}
}

func TestDeadlineQueue(t *testing.T) {
	dq := &monitor.DeadlineQueue{}
	heap.Init(dq)

	heap.Push(dq, big.NewInt(2))
	heap.Push(dq, big.NewInt(4))
	heap.Push(dq, big.NewInt(3))
	heap.Push(dq, big.NewInt(5))
	heap.Push(dq, big.NewInt(1))

	for dq.Len() != 0 {
		// dequeue
		log.Infoln(dq.Top())
		log.Infoln(heap.Pop(dq))
	}
}

func event_callback(id monitor.CallbackID, ethlog types.Log) {
	log.Infof("Call event callback with log: %v\n", ethlog)
}

func TestEvent(t *testing.T) {
	ws, client, storeDir, err := watchService()
	if err != nil {
		t.Fatalf("fail to create watch service: %s", err)
	}
	defer client.Close()
	defer os.RemoveAll(storeDir)
	defer ws.Close()

	ms := monitor.NewService(ws, 0 /* blockDelay */, true /* enabled */, "" /* rpcAddr */)
	ms.Init()

	addr := "5963e46cf9f9700e70d4d1bc09210711ab4a20b4"
	e1 := monitor.Event{
		Name:      "OpenChannel",
		Addr:      ctype.Hex2Addr(addr),
		RawAbi:    ledger.CelerLedgerABI,
		WatchName: "Event 1",
		Callback:  event_callback,
	}

	id1, err := ms.MonitorEvent(e1, false /* reset */)
	if err != nil {
		t.Fatalf("register event #1 failed: %s", err)
	}
	log.Infof("Register callback %d\n", id1)

	e2 := monitor.Event{
		Name:      "ConfirmSettle",
		Addr:      ctype.Hex2Addr(addr),
		RawAbi:    ledger.CelerLedgerABI,
		WatchName: "Event 2",
		Callback:  event_callback,
	}

	id2, err := ms.MonitorEvent(e2, false /* reser */)
	if err != nil {
		t.Fatalf("register event #2 failed: %s", err)
	}
	log.Infof("Register callback %d\n", id2)

	time.Sleep(1 * time.Second)
	ms.RemoveEvent(id1)

	time.Sleep(2 * time.Second)
}
