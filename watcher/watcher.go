// Copyright 2018 Celer Network
//
// This is a wrapper on top of Go-Ethereum's client API (ethclient)
// for fetching log events via filtered query requests.  It handles
// reliable restarts of the application by persisting into the KVStore
// the index of the last log event acknowledged by the application.
// This guarantees that the application does not miss log events that
// occur while the application is not running.

package watcher

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common/structs"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	notBlockIndex = ^uint(0) // 0xFFFFFFFF not-a-block-index value
)

var (
	ErrWatchServiceClosed = errors.New("Watch service closed")
	ErrWatcherTimeout     = errors.New("Watcher timeout")
)

// WatchService holds the active watchers and their connections
// to the Ethereum client and the KVStore persistence layer that
// provides resumability of the watcher after a restart.
type WatchService struct {
	client  WatchClient       // Ethereum client interface
	dal     WatchDAL          // KVStore data access layer
	polling uint64            // Polling interval (msec)
	quit    chan bool         // Terminate the watch service
	mu      sync.RWMutex      // Guards the fields that follow it.
	blkNum  uint64            // Current on-chain block number
	watches map[string]*Watch // Map of registered watches
}

// Watch provides an iterator over a stream of event logs that match
// an Ethereum filtering query.  It updates the KVStore to persist the
// position in the stream of the last event log that the application
// has acknowledged receiving.
// To handle chain reorganization (ephemeral forking), watch only
// requests from on-chain event logs that are older than a specified
// number of on-chain blocks.
type Watch struct {
	name         string               // Unique name of registered watch
	service      *WatchService        // Service owning this watch
	ackWait      bool                 // Is it waiting for an ACK from the app?
	lastAck      bool                 // One last ACK is allowed after close
	ackID        structs.LogEventID   // ID of log event pending an ACK
	lastID       *structs.LogEventID  // ID of log event for resuming (or nil)
	blkDelay     uint64               // Log event delay in number of blocks
	blkInterval  uint64               // Log event polling interval in blocks
	fromBlock    uint64               // Start a fetch from this block number
	query        ethereum.FilterQuery // On-chain event log query
	mu           sync.Mutex           // Guards log queue and "closed" flag
	logQueue     *list.List           // Queue of buffered log events
	logQueueCond *sync.Cond           // Condvar for waiting on the queue
	closed       bool                 // Is the watch closed?
}

// WatchClient is an interface for the subset of functions of the Go-Ethereum
// client API that the watch service uses.
type WatchClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// WatchDAL is an interface for the watch-specific API of the KVStore
// data access layer.
type WatchDAL interface {
	GetLogEventWatch(name string) (*structs.LogEventID, error)
	PutLogEventWatch(name string, id *structs.LogEventID) error
	DeleteLogEventWatch(name string) error
	HasLogEventWatch(name string) (bool, error)
	GetAllLogEventWatchKeys() ([]string, error)
}

// Create a watch service.
func NewWatchService(client WatchClient, dal WatchDAL, polling uint64) *WatchService {
	// Note: the incoming polling interval is in seconds.  Purely for
	// unit testing purposes, the internal "polling" variable is in
	// milliseconds to allow tests to set faster polling intervals.
	return makeWatchService(client, dal, polling*1000)
}

// Helper (for testing) to create a watch service with msec polling interval.
func makeWatchService(client WatchClient, dal WatchDAL, polling uint64) *WatchService {
	if client == nil || dal == nil || polling == 0 {
		return nil
	}

	ws := &WatchService{
		client:  client,
		dal:     dal,
		polling: polling,
		quit:    make(chan bool),
		watches: make(map[string]*Watch),
	}

	// Synchronously initialize the current head block number.
	ws.updateBlockNumber()

	// Start the common on-chain block number watcher.
	go ws.watchBlkNum()

	return ws
}

// Internal goroutine within the watch service that tracks the on-chain
// block number.
func (ws *WatchService) watchBlkNum() {
	ticker := time.NewTicker(time.Duration(ws.polling) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ws.quit:
			clog.Debugln("watchBlkNum: quit")
			return

		case <-ticker.C:
			ws.updateBlockNumber()
		}
	}
}

// Fetch the on-chain block number and update the local value if needed.
func (ws *WatchService) updateBlockNumber() {
	head, err := ws.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		clog.Traceln("cannot fetch on-chain block number:", err)
		return
	}

	blkNum := head.Number.Uint64()
	var topBlkNum uint64
	ws.mu.Lock()
	if blkNum > ws.blkNum {
		ws.blkNum = blkNum
	}
	topBlkNum = ws.blkNum
	ws.mu.Unlock()
	clog.Tracef("top block #: %d, on-chain #: %d", topBlkNum, blkNum)
}

// Return the most recent on-chain block number.
func (ws *WatchService) GetBlockNumber() uint64 {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return ws.blkNum
}

// Return the most recent on-chain block number in big.Int format.
func (ws *WatchService) GetCurrentBlockNumber() *big.Int {
	return new(big.Int).SetUint64(ws.GetBlockNumber())
}

// Close the watch service.
func (ws *WatchService) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.watches != nil {
		close(ws.quit)
		for _, w := range ws.watches {
			w.inner_close()
		}
		ws.watches = nil
	}
}

// Register a named watch.
func (ws *WatchService) register(name string, watch *Watch) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.watches == nil {
		return ErrWatchServiceClosed
	}

	if _, exist := ws.watches[name]; exist {
		return fmt.Errorf("watch name '%s' already in use", name)
	}
	ws.watches[name] = watch
	return nil
}

// Unregister a named watch
func (ws *WatchService) unregister(name string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.watches != nil {
		delete(ws.watches, name)
	}
}

// MakeFilterQuery constructs an Ethereum FilterQuery structure from these
// event and contract parameters: address, raw ABI string, event name, and
// the optional start block number.
func (ws *WatchService) MakeFilterQuery(addr ethcommon.Address, rawABI string, eventName string, startBlock *big.Int) (ethereum.FilterQuery, error) {
	var q ethereum.FilterQuery

	parsedABI, err := abi.JSON(strings.NewReader((rawABI)))
	if err != nil {
		return q, err
	}

	ev, exist := parsedABI.Events[eventName]
	if !exist {
		return q, fmt.Errorf("unknown event name: %s", eventName)
	}

	q.FromBlock = startBlock
	q.Addresses = []ethcommon.Address{addr}
	q.Topics = [][]ethcommon.Hash{{ev.Id()}}
	return q, nil
}

// Create a watch for the given Ethereum log filtering query.
// The block delay is the number of blocks mined used as a time delay
// for fetching event logs, mitigating the effects of chain reorg.
// The block interval controls the polling frequency of fetch logs
// from on-chain, but measured in block numbers (as a delta).
// If "reset" is enabled, the watcher ignores the previously stored
// position in the subscription which resets the stream to its start.
func (ws *WatchService) NewWatch(name string, query ethereum.FilterQuery, blkDelay, blkInterval uint64, reset bool) (*Watch, error) {
	if name == "" {
		return nil, fmt.Errorf("watch name not specified")
	}

	w := &Watch{
		name:        name,
		service:     ws,
		blkDelay:    blkDelay,
		blkInterval: blkInterval,
		query:       query,
		logQueue:    list.New(),
	}

	w.logQueueCond = sync.NewCond(&w.mu) // condvar uses "mu"

	// Register the named watch.
	if err := ws.register(name, w); err != nil {
		return nil, err
	}

	go w.watchLogEvents(reset)

	return w, nil
}

// Internal goroutine that periodically fetches and enqueues the log events
// requested by the watch query.
func (w *Watch) watchLogEvents(reset bool) {
	// Set the ID of the last log event acknowledged by the app (if any).
	// It is used to resume the watch from where the application left off.
	w.setWatchLastID(reset)
	clog.Debugf("watchLogEvents: start %s from %d", w.name, w.fromBlock)

	// The polling interval is computed in relation to block polling.
	polling := w.blkInterval * w.service.polling
	ticker := time.NewTicker(time.Duration(polling) * time.Millisecond)
	defer ticker.Stop()

	for !w.isClosed() {
		w.fetchLogEvents()

		select {
		case <-w.service.quit:
			clog.Debugln("watchLogEvents: quit:", w.name)
			return

		case <-ticker.C:
			continue
		}
	}

	clog.Debugln("watchLogEvents: closed:", w.name)
}

// Set the ID of the last log event acknowledged by the app (if any).
// This is used to resume the watcher from where it left off when the
// application was terminated.  If "reset" is enabled, ignore the value
// in the store and reset it to the one provided by the query.
//
// Note: The design of an ack'ed ID keeps open a small window between
// the first time a watch is started (from a user-specified block number)
// and the first log it gives to the app and gets its ACK.  During that
// time window, a server crash/restart will not have any persisted ID
// and would appear to watch as brand-new, even though that watch was
// previously started.  To handle this state (a watch previously started
// from a block number but never had the chance to give a log to the app)
// it is encoded into the persisted ID using the special "notBlockIndex"
// value for the block "Index": ID = {BlockNumber, notBlockIndex}.
// With that in place, a restarted watch will resume from that previous
// block number again, giving persistence to that special time window.
func (w *Watch) setWatchLastID(reset bool) {
	w.lastID = nil
	w.fromBlock = 0

	exist, err := w.service.dal.HasLogEventWatch(w.name)
	if err == nil && exist && !reset {
		if id, err2 := w.service.dal.GetLogEventWatch(w.name); err2 == nil {
			if id.Index != notBlockIndex {
				// This is a real ACK ID from the app.
				w.lastID = id
			}

			// Ignore the query's "FromBlock" and start from this block.
			w.fromBlock = id.BlockNumber
		}
		return
	}

	// No previously persisted resume pointer.  Remember this first-time
	// starting block number to resume from in case of a crash.
	if w.query.FromBlock != nil {
		w.fromBlock = w.query.FromBlock.Uint64()
	}
	lastID := structs.LogEventID{
		BlockNumber: w.fromBlock,
		Index:       notBlockIndex,
	}
	err = w.service.dal.PutLogEventWatch(w.name, &lastID)
	if err != nil {
		clog.Warnln("cannot persist 1st time resume pointer:", w.name, err)
	}
}

// Fetch a batch of log events from on-chain and enqueue them.
// This function is called periodically in a ticker polling loop.
func (w *Watch) fetchLogEvents() {
	// Do nothing if the on-chain block number has not moved forward
	// beyond the desired block delay (to protect from on-chain reorg).
	blkNum := w.service.GetBlockNumber()
	if w.fromBlock+w.blkDelay > blkNum {
		clog.Tracef("skip log fetching: %s: want %d, delay %d, blk %d", w.name, w.fromBlock, w.blkDelay, blkNum)
		return
	}

	// Fetch server-side filtered log events in the target range of
	// block numbers.  The block delay limit is used to avoid fetching
	// recently mined blocks that may still be undone by a chain reorg.
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	toBlock := blkNum - w.blkDelay
	w.query.FromBlock = new(big.Int).SetUint64(w.fromBlock)
	w.query.ToBlock = new(big.Int).SetUint64(toBlock)

	clog.Tracef("fetch logs: %s: [%d-%d]", w.name, w.fromBlock, toBlock)

	logs, err := w.service.client.FilterLogs(ctx, w.query)
	if err != nil {
		clog.Tracef("cannot fetch logs: %s: [%d-%d]: %s", w.name, w.fromBlock, toBlock, err)
		return
	}

	// Enqueue log events.  In the first fetch after an application resume,
	// lastID is used to skip log events previously handled within a block.
	// This is needed because log fetching is at the block granularity, but
	// processing is at the log event finer granularity.  After that initial
	// resuming block is refetched and partially de-duped, the lastID is no
	// longer needed.
	maxBlock := uint64(0)
	count := 0
	for i := range logs {
		log := &logs[i]
		if log.BlockNumber > maxBlock {
			maxBlock = log.BlockNumber
		}
		if w.lastID == nil {
			w.enqueue(log)
			count++
		} else if greaterThanLastID(log, w.lastID) {
			w.enqueue(log)
			count++
			if log.BlockNumber > w.lastID.BlockNumber {
				w.lastID = nil
			}
		}
	}

	// Update the next block number to start fetching.
	if maxBlock >= w.fromBlock {
		w.fromBlock = maxBlock + 1
	}

	clog.Tracef("added %d logs to queue: %s: next from %d", count, w.name, w.fromBlock)
}

// Return true if the log event ID is strictly greater than the last ID.
func greaterThanLastID(log *types.Log, lastID *structs.LogEventID) bool {
	if log.BlockNumber > lastID.BlockNumber {
		return true
	} else if log.BlockNumber < lastID.BlockNumber {
		return false
	} else if log.Index > lastID.Index {
		return true
	}
	return false
}

// Enqueue an event log.
func (w *Watch) enqueue(log *types.Log) {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldLen := w.logQueue.Len()
	w.logQueue.PushBack(log)
	if oldLen == 0 {
		w.logQueueCond.Signal()
	}
}

// Dequeue and return the first event log in the queue, or block waiting
// for one to arrive or the watcher to be closed.
func (w *Watch) dequeue() (*types.Log, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for !w.closed && w.logQueue.Len() == 0 {
		w.logQueueCond.Wait()
	}

	if w.closed {
		return nil, fmt.Errorf("watch name '%s' closed", w.name)
	}

	elem := w.logQueue.Front()
	w.logQueue.Remove(elem)
	return elem.Value.(*types.Log), nil
}

// Fetch the next log event.  The function will block until either an
// event log is available, or the watcher is closed.
func (w *Watch) Next() (types.Log, error) {
	var empty types.Log

	if w.isClosed() {
		return empty, fmt.Errorf("watch name '%s' already closed", w.name)
	}
	if w.ackWait {
		return empty, fmt.Errorf("last event log received not yet ACKed")
	}

	nextLog, err := w.dequeue()
	if err != nil {
		return empty, err
	}

	w.ackID.BlockNumber = nextLog.BlockNumber
	w.ackID.Index = nextLog.Index
	w.ackWait = true
	return *nextLog, nil
}

// The app ACKs the complete processing of the last received event log.
// Be lenient in one case: after the watch is closed, allow at most one
// more ACK to be done.  This allows event processing that was completed
// by the application when an asynchronous Close() took place (between
// the Next() and the Ack() calls) to be persisted into storage instead
// of having it be re-done after the application is restarted.
func (w *Watch) Ack() error {
	if w.isClosed() {
		if w.lastAck {
			return fmt.Errorf("watch name '%s' already closed", w.name)
		}
		w.lastAck = true
	}
	if !w.ackWait {
		return fmt.Errorf("last event log received already ACKed")
	}

	if err := w.service.dal.PutLogEventWatch(w.name, &w.ackID); err != nil {
		return err
	}
	w.ackWait = false
	return nil
}

func (w *Watch) isClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

func (w *Watch) inner_close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.closed {
		w.closed = true
		w.logQueueCond.Broadcast()
	}
}

// Close a watch subscription.
func (w *Watch) Close() {
	w.service.unregister(w.name) // remove from watch service
	w.inner_close()
}
