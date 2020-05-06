// Copyright 2018-2020 Celer Network

package fsm

import (
	"fmt"

	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
)

// Define directed edge for channel states. NO cycles!
// key is current state, value is allowed possible next states.
var validChanStateTransitions = map[int][]int{
	enums.ChanState_TRUST_OPENED:  []int{enums.ChanState_INSTANTIATING, enums.ChanState_OPENED},
	enums.ChanState_INSTANTIATING: []int{enums.ChanState_OPENED},
	enums.ChanState_OPENED:        []int{enums.ChanState_SETTLING, enums.ChanState_CLOSED},
	enums.ChanState_SETTLING:      []int{enums.ChanState_CLOSED},
}

func ChanStateName(state int) string {
	switch state {
	case enums.ChanState_TRUST_OPENED:
		return "CHANNEL_TRUST_OPENED"
	case enums.ChanState_INSTANTIATING:
		return "CHANNEL_INSTANTIATING"
	case enums.ChanState_OPENED:
		return "CHANNEL_OPENED"
	case enums.ChanState_SETTLING:
		return "CHANNEL_SETTLING"
	case enums.ChanState_CLOSED:
		return "CHANNEL_CLOSED"
	default:
		return "CHANNEL_NULL"
	}
}

func OnChannelUpdate(cid ctype.CidType, currState int) error {
	switch currState {
	case enums.ChanState_TRUST_OPENED, enums.ChanState_INSTANTIATING, enums.ChanState_OPENED:
		return nil
	default:
		return fmt.Errorf("%x OnChannelUpdate err, state %s", cid, ChanStateName(currState))
	}
}

func OnChannelIntendSettle(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	state, found, err := tx.GetChanState(cid)
	if err != nil {
		return fmt.Errorf("%x OnChannelIntendSettle err, state %s, err %w", cid, ChanStateName(state), err)
	}
	if !found {
		return fmt.Errorf("%x OnChannelIntendSettle err, channel not found", cid)
	}
	switch state {
	case enums.ChanState_TRUST_OPENED, enums.ChanState_INSTANTIATING:
		return fmt.Errorf("%x needs to be instantiated on-chain before dispute", cid)
	case enums.ChanState_OPENED:
		return tx.UpdateChanState(cid, enums.ChanState_SETTLING)
	case enums.ChanState_SETTLING:
		return nil
	case enums.ChanState_CLOSED:
		return fmt.Errorf("%x OnPscIntendSettle err, state %s", cid, ChanStateName(state))
	}
	return nil
}

func OnChannelConfirmSettle(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	peer, state, found, err := tx.GetChanPeerState(cid)
	if err != nil {
		return fmt.Errorf("%x OnChannelConfirmSettle err, state %s, err %w", cid, ChanStateName(state), err)
	}
	if !found {
		return fmt.Errorf("%x OnChannelConfirmSettle err, channel not found", cid)
	}
	switch state {
	case enums.ChanState_TRUST_OPENED, enums.ChanState_INSTANTIATING, enums.ChanState_OPENED, enums.ChanState_SETTLING:
		err = tx.UpdatePeerCid(peer, cid, false) // remove cid
		if err != nil {
			return fmt.Errorf("%x OnChannelConfirmSettle err %w", cid, err)
		}
		return tx.UpdateChanState(cid, enums.ChanState_CLOSED)
	case enums.ChanState_CLOSED:
		return nil
	}
	return nil
}

// IsChanStateChangeValid returns true if new state is possible to reach in channel lifecycle
// also return true if new == old
func IsChanStateChangeValid(old, new int) bool {
	return isReachable(validChanStateTransitions, old, new)
}

// ---------------------------- Deprecated state ----------------------------

/**
 *	Currently support 5 states and 4 events
 *	State:
 *		PscTrustOpened
 *		PscInstantiatingTCB
 *		PscOpen
 *		PscDispute
 *		PscClosed
 *	Event:
 *		OnPscAuthOpen
 *		OnPscIntendSettle
 *		OnPscConfirmSettle
 *		OnPscUpdateSimplex
 */

// do NOT modify the string representation of existing states
const (
	PscTrustOpened = "TRUST_OPENED"
	// client waiting for instantiating tcb tx, after tx is mined,
	// state will move to PscOpen
	PscInstantiatingTCB = "INSTANTIATING_TCB"
	PscOpen             = "OPENED"
	PscDispute          = "DISPUTING"
	PscClosed           = "CLOSED"
)
