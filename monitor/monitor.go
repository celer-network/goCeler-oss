// Copyright 2018-2020 Celer Network

package monitor

import (
	"container/heap"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/watcher"
	"github.com/celer-network/goutils/log"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	// default log watch polling as a multiplier of block number if not specified
	// ie. 1 means check log every block
	defaultCheckInterval = uint64(1)
)

// CallbackID is the unique callback ID for deadlines and events
type CallbackID uint64

// Deadline is the metadata of a deadline
type Deadline struct {
	BlockNum *big.Int
	Callback func()
}

// DeadlineQueue is the priority queue for deadlines
type DeadlineQueue []*big.Int

func (dq DeadlineQueue) Len() int { return len(dq) }

func (dq DeadlineQueue) Less(i, j int) bool { return dq[i].Cmp(dq[j]) == -1 }

func (dq DeadlineQueue) Swap(i, j int) { dq[i], dq[j] = dq[j], dq[i] }

func (dq *DeadlineQueue) Push(x interface{}) { *dq = append(*dq, x.(*big.Int)) }

func (dq *DeadlineQueue) Pop() (popped interface{}) {
	popped = (*dq)[len(*dq)-1]
	*dq = (*dq)[:len(*dq)-1]
	return
}

func (dq *DeadlineQueue) Top() (top interface{}) {
	if len(*dq) > 0 {
		top = (*dq)[0]
	}
	return
}

// Config is used by external callers to pass in info, will be converted to Event for internal use
// Reason not use Event directly: Event is more like internal struct
// most fields are from previous MonitorService Monitor func args
// CheckInterval is newly added, meanging to check log for event every x blocks.
// if 0 or not set, defaultCheckInterval (1) will be used
type Config struct {
	EventName            string
	Contract             chain.Contract
	StartBlock, EndBlock *big.Int
	QuickCatch, Reset    bool
	CheckInterval        uint64
}

// Event is the metadata for an event
type Event struct {
	Addr          ctype.Addr
	RawAbi        string
	Name          string
	WatchName     string
	StartBlock    *big.Int
	EndBlock      *big.Int
	BlockDelay    uint64
	CheckInterval uint64
	Callback      func(CallbackID, ethtypes.Log)
	watch         *watcher.Watch
}

// Service struct stores service parameters and registered deadlines and events
type Service struct {
	watch         *watcher.WatchService   // persistent watch service
	deadlines     map[CallbackID]Deadline // deadlines
	deadlinecbs   map[string][]CallbackID // deadline callbacks
	deadlineQueue DeadlineQueue           // deadline priority queue
	events        map[CallbackID]Event    // events
	mu            sync.Mutex
	blockDelay    uint64
	// HACK HACK, the following fields should be removed
	enabled bool
	rpcAddr string
}

// NewService starts a new monitor service. Currently, if "enabled" is false,
// event monitoring will be disabled, and the IP address of the cNode given as
// "rpcAddr" will be printed.
func NewService(
	watch *watcher.WatchService, blockDelay uint64, enabled bool, rpcAddr string) *Service {
	s := &Service{
		watch:       watch,
		deadlines:   make(map[CallbackID]Deadline),
		deadlinecbs: make(map[string][]CallbackID),
		events:      make(map[CallbackID]Event),
		blockDelay:  blockDelay,
		enabled:     enabled,
		rpcAddr:     rpcAddr,
	}
	return s
}

// Init creates the event map
func (s *Service) Init() {
	heap.Init(&s.deadlineQueue)
	go s.monitorDeadlines() // start monitoring deadlines
}

// Close only set events map to empty map so all monitorEvent will exit due to isEventRemoved is true
func (s *Service) Close() {
	s.mu.Lock()
	s.events = make(map[CallbackID]Event)
	s.mu.Unlock()
}

func (s *Service) GetCurrentBlockNumber() *big.Int {
	return s.watch.GetCurrentBlockNumber()
}

// RegisterDeadline registers the deadline and returns the ID
func (s *Service) RegisterDeadline(d Deadline) CallbackID {
	// get a unique callback ID
	s.mu.Lock()
	defer s.mu.Unlock()
	var id CallbackID
	for {
		id = CallbackID(rand.Uint64())
		if _, exist := s.deadlines[id]; !exist {
			break
		}
	}

	// register deadline
	s.deadlines[id] = d
	_, ok := s.deadlinecbs[d.BlockNum.String()]
	if !ok {
		heap.Push(&s.deadlineQueue, d.BlockNum)
	}
	s.deadlinecbs[d.BlockNum.String()] = append(s.deadlinecbs[d.BlockNum.String()], id)
	return id
}

// continuously monitoring deadlines
func (s *Service) monitorDeadlines() {
	for {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		blockNumber := s.GetCurrentBlockNumber()
		for s.deadlineQueue.Len() > 0 && s.deadlineQueue.Top().(*big.Int).Cmp(blockNumber) < 1 {
			timeblock := heap.Pop(&s.deadlineQueue).(*big.Int)
			cbs, ok := s.deadlinecbs[timeblock.String()]
			if ok {
				dlCbs := make(map[CallbackID]Deadline)
				for _, id := range cbs {
					deadline, ok := s.deadlines[id]
					if ok {
						dlCbs[id] = deadline
						delete(s.deadlines, id)
					}
				}
				delete(s.deadlinecbs, timeblock.String())

				s.mu.Unlock()
				for _, deadline := range dlCbs {
					deadline.Callback()
				}
				s.mu.Lock()
			}
		}
		s.mu.Unlock()
	}
}

// Create a watch for the given event.  Use or skip using the StartBlock
// value from the event: the first time a watch is created for an event,
// the StartBlock should be used.  In follow-up re-creation of the watch
// after the previous watch was disconnected, skip using the StartBlock
// because the watch itself has persistence and knows the most up-to-date
// block to resume from instead of the original event StartBlock which is
// stale information by then.  If "reset" is enabled, the watcher ignores the
// previously stored position in the subscription which resets the stream to its
// start.
func (s *Service) createEventWatch(
	e Event, useStartBlock bool, reset bool) (*watcher.Watch, error) {
	var startBlock *big.Int
	if useStartBlock {
		startBlock = e.StartBlock
	}

	q, err := s.watch.MakeFilterQuery(e.Addr, e.RawAbi, e.Name, startBlock)
	if err != nil {
		return nil, err
	}
	if e.CheckInterval == 0 {
		e.CheckInterval = defaultCheckInterval
	}
	return s.watch.NewWatch(e.WatchName, q, e.BlockDelay, e.CheckInterval, reset)
}

func (s *Service) Monitor(cfg *Config, callback func(CallbackID, ethtypes.Log)) (CallbackID, error) {
	if !s.enabled {
		log.Infof("OSP (%s) not listening to on-chain logs", s.rpcAddr)
		return 0, nil
	}
	addr := cfg.Contract.GetAddr()
	watchName := fmt.Sprintf("%s-%s", addr.String(), cfg.EventName)
	eventToListen := &Event{
		Addr:          addr,
		RawAbi:        cfg.Contract.GetABI(),
		Name:          cfg.EventName,
		WatchName:     watchName,
		StartBlock:    cfg.StartBlock,
		EndBlock:      cfg.EndBlock,
		BlockDelay:    s.blockDelay,
		CheckInterval: cfg.CheckInterval,
		Callback:      callback,
	}
	if cfg.QuickCatch {
		eventToListen.BlockDelay = config.QuickCatchBlockDelay
	}
	if eventToListen.CheckInterval == 0 {
		eventToListen.CheckInterval = defaultCheckInterval
	}
	log.Infof("Starting watch: %s. startBlk: %s, endBlk: %s, blkDelay: %d, checkInterval: %d, reset: %t", watchName,
		cfg.StartBlock, cfg.EndBlock, eventToListen.BlockDelay, eventToListen.CheckInterval, cfg.Reset)
	id, err := s.MonitorEvent(*eventToListen, cfg.Reset)
	if err != nil {
		log.Errorf("Cannot register event %s: %s", watchName, err)
		return 0, err
	}
	return id, nil
}

func (s *Service) MonitorEvent(e Event, reset bool) (CallbackID, error) {
	// Construct the watch now to return up-front errors to the caller.
	w, err := s.createEventWatch(e, true /* useStartBlock */, reset)
	if err != nil {
		log.Errorln("register event error:", err)
		return 0, err
	}

	// get a unique callback ID
	s.mu.Lock()
	var id CallbackID
	for {
		id = CallbackID(rand.Uint64())
		if _, exist := s.events[id]; !exist {
			break
		}
	}
	e.watch = w

	// register event
	s.events[id] = e
	s.mu.Unlock()

	go s.monitorEvent(e, id)

	return id, nil
}

func (s *Service) isEventRemoved(id CallbackID) bool {
	s.mu.Lock()
	_, ok := s.events[id]
	s.mu.Unlock()
	return !ok
}

// subscribes to events using a persistent watcher.
func (s *Service) monitorEvent(e Event, id CallbackID) {
	// WatchEvent blocks until an event is caught
	log.Debugln("monitoring event", e.Name)
	for {
		eventLog, err := e.watch.Next()
		if err != nil {
			log.Errorln("monitoring event error:", e.Name, err)
			e.watch.Close()

			var w *watcher.Watch
			for {
				if s.isEventRemoved(id) {
					e.watch.Close()
					return
				}
				w, err = s.createEventWatch(e, false /* useStartBlock */, false /* reset */)
				if err == nil {
					break
				}
				log.Errorln("recreate event watch error:", e.Name, err)
				time.Sleep(1 * time.Second)
			}

			s.mu.Lock()
			e.watch = w
			s.mu.Unlock()
			log.Debugln("event watch recreated", e.Name)
			continue
		}

		// When event log is removed due to chain re-org, just ignore it
		// TODO: emit error msg and properly roll back upon catching removed event log
		if eventLog.Removed {
			log.Warnf("Receive removed %s event log", e.Name)
			if err = e.watch.Ack(); err != nil {
				log.Errorln("monitoring event ACK error:", e.Name, err)
				e.watch.Close()
				return
			}
			continue
		}

		// Stop watching if the event was removed
		// TODO(mzhou): Also stop monitoring if timeout has passed
		if s.isEventRemoved(id) {
			e.watch.Close()
			return
		}

		e.Callback(id, eventLog)

		if err = e.watch.Ack(); err != nil {
			// This is a coding bug, just exit the loop.
			log.Errorln("monitoring event ACK error:", e.Name, err)
			e.watch.Close()
			return
		}
		if s.isEventRemoved(id) {
			e.watch.Close()
			return
		}
	}
}

// RemoveDeadline removes a deadline from the monitor
func (s *Service) RemoveDeadline(id CallbackID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Debugf("revoke deadline monitoring %d", id)
	delete(s.deadlines, id)
}

// RemoveEvent removes an event from the monitor
func (s *Service) RemoveEvent(id CallbackID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.events[id]
	if ok {
		log.Debugf("revoke event monitoring %d event %s", id, e.Name)
		e.watch.Close()
		delete(s.events, id)
	}
}

// NewEventStr generates the event using contract address and event name in format "<addr>-<name>"
func NewEventStr(ledgerAddr ctype.Addr, eventName string) string {
	return fmt.Sprintf("%s-%s", ctype.Addr2Hex(ledgerAddr), eventName)
}
