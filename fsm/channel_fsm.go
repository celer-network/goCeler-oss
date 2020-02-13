// Copyright 2018-2019 Celer Network

package fsm

import (
	"fmt"

	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/storage"
)

/**
 *	Currently support 5 states and 4 events
 *	State:
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
	PscOpen    = "OPENED"
	PscDispute = "DISPUTING"
	PscClosed  = "CLOSED"
)

func OnPscAuthOpen(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	state, _, err := tx.GetChannelState(cid)
	if err == nil {
		return fmt.Errorf("%x OnPscAuthOpen err, state %s", cid, state)
	}
	err = addPeerActiveChannel(tx, cid)
	if err != nil {
		return err
	}
	return tx.PutChannelState(cid, PscOpen)
}

func OnPscIntendSettle(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	state, _, err := tx.GetChannelState(cid)
	if err != nil {
		return fmt.Errorf("%x OnPscIntendSettle err, state %s, err %s", cid, state, err)
	}
	switch state {
	case PscOpen:
		return tx.PutChannelState(cid, PscDispute)
	case PscDispute:
		return nil
	case PscClosed:
		return fmt.Errorf("%x OnPscIntendSettle err, state %s", cid, state)
	}
	return nil
}

func OnPscConfirmSettle(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	state, _, err := tx.GetChannelState(cid)
	if err != nil {
		return fmt.Errorf("%x OnPscConfirmSettle err, state %s, err %s", cid, state, err)
	}
	switch state {
	case PscOpen, PscDispute:
		err = deletePeerActiveChannel(tx, cid)
		if err != nil {
			return err
		}
		return tx.PutChannelState(cid, PscClosed)
	case PscClosed:
		return nil
	}
	return nil
}

func OnPscUpdateSimplex(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	state, _, err := tx.GetChannelState(cid)
	if err != nil {
		return fmt.Errorf("%x OnPscUpdateSimplex err, state %s, err %s", cid, state, err)
	}
	switch state {
	case PscOpen:
		// put the same state again to update the activity timestamp
		return tx.PutChannelState(cid, state)
	case PscDispute, PscClosed:
		return fmt.Errorf("%x OnPscUpdateSimplex err, state %s", cid, state)
	}
	return nil
}

func addPeerActiveChannel(tx *storage.DALTx, cid ctype.CidType) error {
	peer, err := tx.GetPeer(cid)
	if err != nil {
		return err
	}
	var cids map[ctype.CidType]bool
	exist, err := tx.HasPeerActiveChannels(peer)
	if err != nil {
		return err
	}
	if exist {
		cids, err = tx.GetPeerActiveChannels(peer)
		if err != nil {
			return err
		}
	} else {
		cids = make(map[ctype.CidType]bool)
	}
	cids[cid] = true
	return tx.PutPeerActiveChannels(peer, cids)
}

func deletePeerActiveChannel(tx *storage.DALTx, cid ctype.CidType) error {
	var err error
	peer, err := tx.GetPeer(cid)
	if err != nil {
		return err
	}
	var cids map[ctype.CidType]bool
	exist, err := tx.HasPeerActiveChannels(peer)
	if err != nil {
		return err
	}
	if exist {
		cids, err = tx.GetPeerActiveChannels(peer)
		if err != nil {
			return err
		}
		delete(cids, cid)
		if len(cids) == 0 {
			return tx.DeletePeerActiveChannels(peer)
		} else {
			return tx.PutPeerActiveChannels(peer, cids)
		}
	}
	return nil
}
