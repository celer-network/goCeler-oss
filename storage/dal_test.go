// Copyright 2018 Celer Network

package storage

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/ctype"
)

func TestPeerLookUpTable(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)

	key := ctype.Bytes2Cid([]byte{'a'})

	peer, err := dal.GetPeer(key)
	if err == nil {
		t.Errorf("found non-existing peer: %v", peer)
	}

	val := "peer0"

	err = dal.PutPeer(key, val)
	if err != nil {
		t.Errorf("error storing peer: %s", err)
	}

	peer, err = dal.GetPeer(key)
	if err != nil {
		t.Errorf("error getting peer: %s", err)
	} else if peer != val {
		t.Errorf("got bad peer: %s != %s", peer, val)
	}

	err = dal.DeletePeer(key)
	if err != nil {
		t.Errorf("error deleting peer: %s", err)
	}

	peer, err = dal.GetPeer(key)
	if err == nil {
		t.Errorf("found peer after delete: %s", peer)
	}

	exists, err := dal.HasPeer(key)
	if err != nil {
		t.Errorf("error checking peer existence: %s", err)
	} else if exists {
		t.Errorf("Peer still exists")
	}

	cids, err := dal.GetAllPeerLookUpTableKeys()
	if err != nil {
		t.Errorf("error getting all peer keys: %s", err)
	}
	if len(cids) != 0 {
		t.Errorf("got wrong cid length: %d", len(cids))
	}

	// Repeat the steps using transactions.
	tx, err := dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	err = tx.PutPeer(key, val)
	if err != nil {
		t.Errorf("error storing peer in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	peer, err = tx.GetPeer(key)
	if err != nil {
		t.Errorf("error getting peer in Tx: %s", err)
	} else if peer != val {
		t.Errorf("got bad peer in Tx: %s != %s", peer, val)
	}

	tx.Discard()

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 3rd transaction: %s", err)
	}

	err = tx.DeletePeer(key)
	if err != nil {
		t.Errorf("error deleting peer in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 4th transaction: %s", err)
	}

	peer, err = tx.GetPeer(key)
	if err == nil {
		t.Errorf("found peer after delete in Tx: %s", peer)
	}

	exists, err = tx.HasPeer(key)
	if err != nil {
		t.Errorf("error checking peer existence in Tx: %s", err)
	} else if exists {
		t.Errorf("Peer still exists in Tx")
	}

	cids, err = tx.GetAllPeerLookUpTableKeys()
	if err != nil {
		t.Errorf("error getting all peer keys in Tx: %s", err)
	}
	if len(cids) != 0 {
		t.Errorf("got wrong cid length in Tx: %d", len(cids))
	}

	tx.Discard()
}

func TestSecretRegistry(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)

	key := "hash0"

	preimage, err := dal.GetSecretRegistry(key)
	if err == nil {
		t.Errorf("found non-existing preimage: %v", preimage)
	}

	val := "preimage0"

	err = dal.PutSecretRegistry(key, val)
	if err != nil {
		t.Errorf("error storing preimage: %s", err)
	}

	preimage, err = dal.GetSecretRegistry(key)
	if err != nil {
		t.Errorf("error getting preimage: %s", err)
	} else if preimage != val {
		t.Errorf("got bad preimage: %s != %s", preimage, val)
	}

	exists, err := dal.HasSecretRegistry(key)
	if err != nil {
		t.Errorf("error checking preimage existence: %s", err)
	} else if !exists {
		t.Errorf("Preimage does not exist")
	}

	cids, err := dal.GetAllSecretRegistryKeys()
	if err != nil {
		t.Errorf("error getting all preimage keys: %s", err)
	}
	if len(cids) != 1 {
		t.Errorf("got wrong cid length: %d", len(cids))
	}

	err = dal.DeleteSecretRegistry(key)
	if err != nil {
		t.Errorf("error deleting preimage: %s", err)
	}

	preimage, err = dal.GetSecretRegistry(key)
	if err == nil {
		t.Errorf("found preimage after delete: %s", preimage)
	}

	// Repeat the steps using transactions.
	tx, err := dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	err = tx.PutSecretRegistry(key, val)
	if err != nil {
		t.Errorf("error storing preimage in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	preimage, err = tx.GetSecretRegistry(key)
	if err != nil {
		t.Errorf("error getting preimage in Tx: %s", err)
	} else if preimage != val {
		t.Errorf("got bad preimage in Tx: %s != %s", preimage, val)
	}

	exists, err = tx.HasSecretRegistry(key)
	if err != nil {
		t.Errorf("error checking preimage existence in Tx: %s", err)
	} else if !exists {
		t.Errorf("Preimage does not exist in Tx")
	}

	cids, err = tx.GetAllSecretRegistryKeys()
	if err != nil {
		t.Errorf("error getting all preimage keys in Tx: %s", err)
	}
	if len(cids) != 1 {
		t.Errorf("got wrong cid length in Tx: %d", len(cids))
	}

	tx.Discard()

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 3rd transaction: %s", err)
	}

	err = tx.DeleteSecretRegistry(key)
	if err != nil {
		t.Errorf("error deleting preimage in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 4th transaction: %s", err)
	}

	preimage, err = tx.GetSecretRegistry(key)
	if err == nil {
		t.Errorf("found preimage after delete in Tx: %s", preimage)
	}

	tx.Discard()
}

func TestRoutingTable(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)

	key := "faraway"

	route, err := dal.GetRoute(key, common.EthContractAddr)
	if err == nil {
		t.Errorf("found non-existing route: %s", route)
	}

	val := ctype.Bytes2Cid([]byte{'a'})

	err = dal.PutRoute(key, common.EthContractAddr, val)
	if err != nil {
		t.Errorf("error storing route: %s", err)
	}

	route, err = dal.GetRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error getting route: %s", err)
	} else if route != val {
		t.Errorf("got bad route: %s != %s", route, val)
	}

	err = dal.DeleteRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error deleting route: %s", err)
	}

	route, err = dal.GetRoute(key, common.EthContractAddr)
	if err == nil {
		t.Errorf("found route after delete: %s", route)
	}

	exists, err := dal.HasRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error checking route existence: %s", err)
	} else if exists {
		t.Errorf("Route still exists")
	}

	cids, err := dal.GetAllRoutingTableKeys()
	if err != nil {
		t.Errorf("error getting all routing table keys: %s", err)
	}
	if len(cids) != 0 {
		t.Errorf("got wrong cid length: %d", len(cids))
	}

	// Repeat the steps using transactions.
	tx, err := dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	err = tx.PutRoute(key, common.EthContractAddr, val)
	if err != nil {
		t.Errorf("error storing route in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	route, err = tx.GetRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error getting route in Tx: %s", err)
	} else if route != val {
		t.Errorf("got bad route in Tx: %s != %s", route, val)
	}

	tx.Discard()

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}
	cids, err = tx.GetAllRoutingTableKeysToDest(key)
	if err != nil {
		t.Errorf("error getting all routing table keys to %s in Tx: %s", key, err)
	}
	if len(cids) != 1 {
		t.Errorf("got wrong cid length in Tx: %d", len(cids))
	}
	if cids[0] != key+common.RoutingTableDestTokenSpliter+common.EthContractAddr {
		t.Errorf("got wrong routing entry in Tx: %s", cids[0])
	}
	tx.Discard()

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 3rd transaction: %s", err)
	}

	err = tx.DeleteRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error deleting route in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 4th transaction: %s", err)
	}

	route, err = tx.GetRoute(key, common.EthContractAddr)
	if err == nil {
		t.Errorf("found route after delete in Tx: %s", route)
	}

	exists, err = tx.HasRoute(key, common.EthContractAddr)
	if err != nil {
		t.Errorf("error checking route existence in Tx: %s", err)
	} else if exists {
		t.Errorf("Route still exists in Tx")
	}

	cids, err = tx.GetAllRoutingTableKeys()
	if err != nil {
		t.Errorf("error getting all routing table keys in Tx: %s", err)
	}
	if len(cids) != 0 {
		t.Errorf("got wrong cid length in Tx: %d", len(cids))
	}

	tx.Discard()
}

func TestLogEventWatch(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)

	key := "foobar"

	id, err := dal.GetLogEventWatch(key)
	if err == nil {
		t.Errorf("found non-existing log event ID: %v", *id)
	}

	val := structs.LogEventID{
		BlockNumber: 1234,
		Index:       3,
	}

	err = dal.PutLogEventWatch(key, &val)
	if err != nil {
		t.Errorf("error storing log event ID: %s", err)
	}

	id, err = dal.GetLogEventWatch(key)
	if err != nil {
		t.Errorf("error getting log event ID: %s", err)
	} else if !reflect.DeepEqual(*id, val) {
		t.Errorf("got bad log event ID: %v != %v", *id, val)
	}

	names, err := dal.GetAllLogEventWatchKeys()
	if err != nil {
		t.Errorf("error getting all log event ID keys: %s", err)
	}
	exp := []string{key}
	if !reflect.DeepEqual(names, exp) {
		t.Errorf("got wrong names: %v != %v", names, exp)
	}

	err = dal.DeleteLogEventWatch(key)
	if err != nil {
		t.Errorf("error deleting log event ID: %s", err)
	}

	id, err = dal.GetLogEventWatch(key)
	if err == nil {
		t.Errorf("found log event ID after delete: %v", *id)
	}

	exists, err := dal.HasLogEventWatch(key)
	if err != nil {
		t.Errorf("error checking log event ID existence: %s", err)
	} else if exists {
		t.Errorf("Log event ID still exists")
	}

	// Repeat the steps using transactions.
	tx, err := dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	err = tx.PutLogEventWatch(key, &val)
	if err != nil {
		t.Errorf("error storing log event ID in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	id, err = tx.GetLogEventWatch(key)
	if err != nil {
		t.Errorf("error getting log event ID in Tx: %s", err)
	} else if !reflect.DeepEqual(*id, val) {
		t.Errorf("got bad log event ID in Tx: %v != %v", *id, val)
	}

	names, err = tx.GetAllLogEventWatchKeys()
	if err != nil {
		t.Errorf("error getting all log event ID keys in Tx: %s", err)
	}
	if !reflect.DeepEqual(names, exp) {
		t.Errorf("got wrong names in Tx: %v != %v", names, exp)
	}

	tx.Discard()

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 3rd transaction: %s", err)
	}

	err = tx.DeleteLogEventWatch(key)
	if err != nil {
		t.Errorf("error deleting log event ID in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 4th transaction: %s", err)
	}

	id, err = tx.GetLogEventWatch(key)
	if err == nil {
		t.Errorf("found log event ID after delete in Tx: %v", *id)
	}

	if exists, err := tx.HasLogEventWatch(key); err != nil {
		t.Errorf("error checking log event ID existence in Tx: %s", err)
	} else if exists {
		t.Errorf("Log event ID still exists in Tx")
	}

	tx.Discard()
}

func TestTokenContract(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)
	key := ctype.Bytes2Cid([]byte{'a'})

	tokenAddr, err := dal.GetTokenContractAddr(key)
	if err == nil {
		t.Errorf("found non-existing token addr: %v", tokenAddr)
	}

	val := "tokenAddr0"

	err = dal.PutTokenContractAddr(key, val)
	if err != nil {
		t.Errorf("error storing token addr: %s", err)
	}

	tokenAddr, err = dal.GetTokenContractAddr(key)
	if err != nil {
		t.Errorf("error getting token addr: %s", err)
	} else if tokenAddr != val {
		t.Errorf("got bad token addr: %v != %v", tokenAddr, val)
	}

	// Repeat the steps using transactions.
	tx, err := dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	err = tx.PutTokenContractAddr(key, val)
	if err != nil {
		t.Errorf("error storing token addr in Tx: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit Tx: %s", err)
	}

	tx, err = dal.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	tokenAddr, err = tx.GetTokenContractAddr(key)
	if err != nil {
		t.Errorf("error getting token addr in Tx: %s", err)
	} else if tokenAddr != val {
		t.Errorf("got bad token addr in Tx: %v != %v", tokenAddr, val)
	}

	tx.Discard()
}

func TestTransactional(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	dal := NewDAL(st)

	// Use the peer lookup table to test the Transactional() API.
	key := ctype.Bytes2Cid([]byte{'a'})
	val := "peer0"

	peer, err := dal.GetPeer(key)
	if err == nil {
		t.Errorf("found non-existing peer: %v", peer)
	}

	// Put the entry using Transactional().
	txBody := func(tx *DALTx, args ...interface{}) error {
		cid := args[0].(ctype.CidType)
		val := args[1].(string)
		return tx.PutPeer(cid, val)
	}

	err = dal.Transactional(txBody, key, val)
	if err != nil {
		t.Errorf("failed transactional commit: %s", err)
	}

	peer, err = dal.GetPeer(key)
	if err != nil {
		t.Errorf("error getting peer: %s", err)
	} else if peer != val {
		t.Errorf("got bad peer: %s != %s", peer, val)
	}

	// Use a failing callback.
	txBody = func(tx *DALTx, args ...interface{}) error {
		errMsg := args[0].(string)
		return fmt.Errorf(errMsg)
	}

	errMsg := "error inside Transactional()"
	err = dal.Transactional(txBody, errMsg)
	if err == nil {
		t.Errorf("transactional did not fail")
	} else if err.Error() != errMsg {
		t.Errorf("got bad error message: %s != %s", err.Error(), errMsg)
	}
}
