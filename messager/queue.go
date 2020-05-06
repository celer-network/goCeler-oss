// Copyright 2019-2020 Celer Network
//
// Message queue to store and track the delivery of Celer Messages
// to payment channels, one queue per channel (destination).
//
// * It allows message pipelining per payment channel.
// * Messages are sent and processed in sequence number order.
// * Messages are persisted to the database and deleted when ACKed.
// * The set of active channels (message queues) is dynamic as local
//   peers change when clients connect & disconnect from a server.
// * Per queue it tracks the last-ACKed, last-sent, and last-added
//   messages. The 3 numbers represent the state of a queue.
// * One goroutine handles all payment channel queues, sending one message
//   from each non-empty queue in a round-robin manner.

package messager

import (
	"fmt"
	"sync"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
)

const (
	workChunk = 64 // limit of # active peers to fetch
)

type MsgQueue struct {
	dal          *storage.DAL
	streamWriter common.StreamWriter
	myAddr       ctype.Addr
	mu           sync.Mutex
	workCond     *sync.Cond                     // has work to do (cond var)
	work         map[ctype.CidType]bool         // set of cids to work on
	queues       map[ctype.CidType]*Queue       // one queue per cid
	peerCids     map[ctype.Addr][]ctype.CidType // set of cids per peer
}

type Queue struct {
	peer  ctype.Addr               // channel peer
	acked uint64                   // last ACKed message
	sent  uint64                   // last sent message
	added uint64                   // last added message
	msgs  map[uint64]*rpc.CelerMsg // messages to be sent
}

func NewMsqQueue(dal *storage.DAL, streamWriter common.StreamWriter, myAddr ctype.Addr) *MsgQueue {
	m := &MsgQueue{
		dal:          dal,
		streamWriter: streamWriter,
		myAddr:       myAddr,
		queues:       make(map[ctype.CidType]*Queue),
		work:         make(map[ctype.CidType]bool),
		peerCids:     make(map[ctype.Addr][]ctype.CidType),
	}
	m.workCond = sync.NewCond(&m.mu)

	go m.run()
	return m
}

// Message queue runner goroutine that sends messages to all peers,
// one message per active payment channel.
func (m *MsgQueue) run() {
	for {
		cids := m.waitForWork()

		var wg sync.WaitGroup
		wg.Add(len(cids))

		for _, cid := range cids {
			go m.sendNextMessage(cid, &wg)
		}

		wg.Wait()
	}
}

// Wait for some active channels (i.e. non-empty message queues).
func (m *MsgQueue) waitForWork() []ctype.CidType {
	log.Tracef("MsgQueue: waiting for work")

	m.mu.Lock()
	for len(m.work) == 0 {
		m.workCond.Wait()
	}

	// Return a chunk of the active channels to avoid making this call
	// too slow when the number of active channels is too large.  This
	// makes it less fair than strict round-robin across all channels,
	// but it is mitigated by Golang's randomization of the traversal
	// of a map, so each waitForWork() call picks a different chunk
	// from the map, even if the map did not change.
	size := len(m.work)
	if size > workChunk {
		size = workChunk
	}

	cids := make([]ctype.CidType, 0, size)
	for cid := range m.work {
		if size == 0 {
			break
		}
		cids = append(cids, cid)
		size--
	}

	m.mu.Unlock()

	log.Tracef("MsgQueue: got work for %d channels", len(cids))
	return cids
}

// Send to this channel its next queued message (by sequence number).
func (m *MsgQueue) sendNextMessage(cid ctype.CidType, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Tracef("MsgQueue: sending next message to %x", cid)

	m.mu.Lock()
	q := m.queues[cid]
	if q == nil || q.sent == q.added {
		log.Tracef("MsgQueue: nothing to do for %x", cid)
		delete(m.work, cid)
		m.mu.Unlock()
		return
	}

	seqnum := q.sent + 1
	msg := q.msgs[seqnum]
	m.mu.Unlock()

	// After a restart, messages are fetched here on-demand.
	if msg == nil {
		var err error
		var found bool
		msg, found, err = m.dal.GetChanMessage(cid, seqnum)
		if err != nil || !found {
			m.updateQueueSent(cid, seqnum)
			log.Errorf("MsgQueue: cannot get msg %d from storage to send to %x: %w", seqnum, cid, err)
			return
		}
	}

	log.Tracef("MsgQueue: sending msg %d to %x", seqnum, cid)
	// Send the message
	err := m.streamWriter.WriteCelerMsg(q.peer, msg)
	if err != nil {
		log.Warnf("MsgQueue: cannot send msg %d to %x,%x: %s", seqnum, q.peer, cid, err)
		return
	}

	// Peer queue may have been removed in between locks.
	recorded := m.updateQueueSent(cid, seqnum)
	if recorded {
		log.Tracef("MsgQueue: msg %d sent to %x, %x", seqnum, q.peer, cid)
	} else {
		log.Tracef("MsgQueue: msg %d sent to %x, %x but not recorded", seqnum, q.peer, cid)
	}
}

func (m *MsgQueue) updateQueueSent(cid ctype.CidType, seqnum uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	recorded := false

	q := m.queues[cid]
	if q != nil {
		q.sent = seqnum
		recorded = true
	}
	return recorded
}

// Indicate that this channel has pending work and notify the runner if needed.
// The caller must hold the mutex when calling this function.
func (m *MsgQueue) hasWork(cid ctype.CidType) {
	prevLen := len(m.work)
	m.work[cid] = true
	if prevLen == 0 {
		m.workCond.Signal() // wakeup the runner goroutine
	}
}

// Add a message for a channel. The message itself must have been saved
// to storage before calling this function.  This is typically done
// atomically inside a store transaction along with other updates, and
// if successful, AddMsg() is called to notify the message queue.
func (m *MsgQueue) AddMsg(peer ctype.Addr, cid ctype.CidType, seqnum uint64, msg *rpc.CelerMsg) error {
	log.Tracef("MsgQueue: add msg %d to cid %x", seqnum, cid)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.peerCids[peer] == nil {
		return fmt.Errorf("MsgQueue: cannot add msg %d, unconnected peer %x", seqnum, peer)
	}

	q := m.queues[cid]
	if q == nil {
		if seqnum == 1 { // new channel added after peer connected
			m.queues[cid] = &Queue{
				peer: peer,
				msgs: make(map[uint64]*rpc.CelerMsg),
			}
			m.addPeerCid(peer, cid)
			q = m.queues[cid]
		} else {
			return fmt.Errorf("MsgQueue: cannot add msg %d, unknown cid %x", seqnum, cid)
		}
	}

	if q.added < seqnum {
		q.added = seqnum
	}
	q.msgs[seqnum] = msg
	m.hasWork(cid)
	return nil
}

// Acknowledge the message being received by the peer, either accepted (ack)
// or not accepted (nack). This removes the  message from the queue. If the
// message being ACKed is not the next expected ACK, this is treated as all
// messages in between being received.

// Note: this function is not symmetrical to AddMsg() which only adds a
// message to the queue separately from it being written to storage before.
// This flows from the different requirements in how messages are created
// compared to how they are ACKed and deleted.
func (m *MsgQueue) AckMsg(cid ctype.CidType, ack, nack uint64) error {
	if ack == nack {
		// log err and let continue as it won't trigger worse consequence
		log.Errorf("MsgQueue: ACK and NACK should not have the same seq %d, cid %x", ack, cid)
	}

	if nack > ack {
		log.Tracef("MsgQueue: Acking msg %d and NACKing msg %d to cid %x", ack, nack, cid)
	} else {
		log.Tracef("MsgQueue: ACKing msg %d to cid %x", ack, cid)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.queues[cid]
	if q == nil {
		return fmt.Errorf("MsgQueue: cannot ACK msg %d, unknown cid %x", ack, cid)
	}

	if nack > q.sent {
		// messages with seqnum smaller than nack do not need to be sent,
		// because they are all based on the unaccepted (nacked) state.
		q.sent = nack
	}

	if ack <= q.acked {
		return nil // previously ACKed
	}

	if ack > q.sent {
		// acked may be larger than sent because the node starts to resend from acked+1
		// on peer reconnect, and the reponse msg may ack a larger seq received by the
		// peer before it disconnected. Then msg equal or smaller than the new acked do
		// not need to be sent again
		q.sent = ack
	}

	from, to := q.acked+1, ack
	q.acked = ack

	for i := from; i <= to; i++ {
		delete(q.msgs, i)
	}

	if q.acked == q.added {
		delete(m.work, cid) // empty queue
	}

	return nil
}

func (m *MsgQueue) ResendMsg(cid ctype.CidType, seqnum uint64) error {
	log.Tracef("MsgQueue: resend msg %d to cid %x", seqnum, cid)

	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.queues[cid]
	if q == nil {
		return fmt.Errorf("MsgQueue: cannot resend msg %d, unknown cid %x", seqnum, cid)
	}

	if seqnum <= q.acked {
		return fmt.Errorf("MsgQueue: should not resend msg %d to cid %x: already acked at %d", seqnum, cid, q.acked)
	}

	if seqnum-1 < q.sent {
		q.sent = seqnum - 1
		m.hasWork(cid)
	}

	return nil
}

// GetMsg returns <msg, exist> from the queue
func (m *MsgQueue) GetMsg(cid ctype.CidType, seqnum uint64) (*rpc.CelerMsg, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.queues[cid]
	if q == nil {
		return nil, false
	}
	msg, ok := q.msgs[seqnum]
	return msg, ok
}

// Fetch the message queue status from storage for this channel.
// Return the (acked, sent, added) sequence numbers. The returned value sent = acked
// because node always resends from the first unacked msg on peer reconnect.
func (m *MsgQueue) getChannelQueueStatus(cid ctype.CidType) (uint64, uint64, uint64, error) {
	_, lastUsed, lastAcked, lastNacked, found, err := m.dal.GetChanSeqNums(cid)
	if err != nil {
		return 0, 0, 0, err
	}
	if !found {
		return 0, 0, 0, common.ErrChannelNotFound
	}

	if lastAcked > lastUsed {
		err = fmt.Errorf("invalid queue status: (acked %d, added %d)", lastAcked, lastUsed)
		return 0, 0, 0, err
	}
	sent := lastAcked
	if lastNacked > lastAcked {
		sent = lastNacked
	}
	return lastAcked, sent, lastUsed, nil
}

func (m *MsgQueue) addQueue(peer ctype.Addr, cid ctype.CidType) error {
	acked, sent, added, err := m.getChannelQueueStatus(cid)
	if err != nil {
		return fmt.Errorf("MsgQueue: cannot init queue status for %x: %w", cid, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.queues[cid] != nil {
		return fmt.Errorf("MsgQueue: cannot add channel %x, already exists", cid)
	}

	m.queues[cid] = &Queue{
		peer:  peer,
		acked: acked,
		sent:  sent,
		added: added,
		msgs:  make(map[uint64]*rpc.CelerMsg),
	}

	if sent < added {
		m.hasWork(cid)
	}

	m.addPeerCid(peer, cid)

	return nil
}

func (m *MsgQueue) addPeerCid(peer ctype.Addr, cid ctype.CidType) {
	m.peerCids[peer] = append(m.peerCids[peer], cid)
}

// A peer connected to this server, add its channels to the message queue and
// start (or resume) managing its messages.
func (m *MsgQueue) AddPeer(peer ctype.Addr) error {
	log.Tracef("MsgQueue: adding peer %x", peer)
	cids, found, err := m.dal.GetPeerCids(peer)
	if err != nil {
		return fmt.Errorf("MsgQueue: cannot add peer %x, err: %w", peer, err)
	}

	m.mu.Lock()
	if len(m.peerCids[peer]) > 0 {
		m.mu.Unlock()
		return fmt.Errorf("MsgQueue: cannot add peer %x, already exist", peer)
	}
	// Continue running even if no cids were found because the "peerCids"
	// map entry must be initialized.
	if !found {
		log.Tracef("MsgQueue: no active channels for peer %x: %w", peer, err)
	}
	m.peerCids[peer] = make([]ctype.CidType, 0, len(cids))
	m.mu.Unlock()
	for _, cid := range cids {
		err := m.addQueue(peer, cid)
		if err != nil {
			log.Warnln(err)
		}
	}
	return nil
}

// A peer disconnected from this server, remove it from the message queue
// and stop managing its messages.
func (m *MsgQueue) RemovePeer(peer ctype.Addr) error {
	log.Tracef("MsgQueue: removing peer %x", peer)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.peerCids[peer] == nil {
		return fmt.Errorf("MsgQueue: cannot remove peer %x, does not exist", peer)
	}
	for _, cid := range m.peerCids[peer] {
		delete(m.queues, cid)
		delete(m.work, cid)
	}
	delete(m.peerCids, peer)

	return nil
}
