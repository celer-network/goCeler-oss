// Copyright 2019-2020 Celer Network

package storage

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	stDriverLT = "sqlite3"
	stDir      = "/tmp/storage_test_sql_db"
)

// Return a temporary store file for testing without creating the file.
func tempStoreFile() string {
	user, _ := user.Current()
	ts := time.Now().UnixNano() / 1000
	dir := fmt.Sprintf("sql-store-%s-%d.db", user.Username, ts)
	return filepath.Join(os.TempDir(), dir)
}

func TestMain(m *testing.M) {
	flag.Parse()

	if err := setupDB(); err != nil {
		fmt.Println("cannot setup DB:", err)
		os.Exit(1)
	}

	exitCode := m.Run() // run all unittests

	teardownDB()
	os.Exit(exitCode)
}

func setupDB() error {
	err := os.RemoveAll(stDir)
	if err != nil {
		return fmt.Errorf("cannot remove old DB directory: %s: %s", stDir, err)
	}

	return nil
}

func teardownDB() {
	os.RemoveAll(stDir)
}

type TestFunc func(*testing.T, *KVStoreSQL)

// Create a SQL store of the client or server type and use it
// to run the given testing callback function.
func runWithDatabase(t *testing.T, client bool, testCallback TestFunc) {
	var st *KVStoreSQL
	var err error

	if client {
		stFile := tempStoreFile()
		st, err = NewKVStoreSQL(stDriverLT, stFile)
		if err != nil {
			t.Fatalf("cannot create SQLite store %s: %s", stFile, err)
		}
		defer st.Close()
		defer os.Remove(stFile)
	} else {
		t.Errorf("cannot test with non-client store")
		return
	}

	testCallback(t, st)
}

func testKVStoreSQLOps(t *testing.T, st *KVStoreSQL) {
	type Foo struct {
		Name  string
		Count int
	}

	val := Foo{
		Name:  "hello",
		Count: 567,
	}
	val2 := []byte("hello world 1234")

	for i := 1; i < 5; i++ {
		tbl := fmt.Sprintf("t%d", i)
		if exists, err := st.Has(tbl, "foo"); err != nil {
			t.Errorf("error checking if %s:foo exists: %s", tbl, err)
		} else if exists {
			t.Errorf("%s:foo should no exist", tbl)
		}
	}

	if err := st.Put("t1", "foo", "bar"); err != nil {
		t.Errorf("error storing t1:foo string: %s", err)
	}
	if err := st.Put("t2", "foo", 1234); err != nil {
		t.Errorf("error storing t2:foo int: %s", err)
	}
	if err := st.Put("t3", "foo", &val); err != nil {
		t.Errorf("error storing t3:foo struct: %s", err)
	}
	if err := st.Put("t4", "foo", val2); err != nil {
		t.Errorf("error storing t4:foo []byte: %s", err)
	}

	for i := 1; i < 5; i++ {
		tbl := fmt.Sprintf("t%d", i)
		if exists, err := st.Has(tbl, "foo"); err != nil {
			t.Errorf("error checking if %s:foo exists: %s", tbl, err)
		} else if !exists {
			t.Errorf("%s:foo should exist", tbl)
		}
	}

	var t1Foo string
	var t2Foo int
	var t3Foo Foo
	var t4Foo []byte

	if err := st.Get("t1", "foo", &t1Foo); err != nil {
		t.Errorf("error fetching t1:foo string: %s", err)
	} else if t1Foo != "bar" {
		t.Errorf("got bad t1:foo string value: %s", t1Foo)
	}

	if err := st.Get("t2", "foo", &t2Foo); err != nil {
		t.Errorf("error fetching t2:foo int: %s", err)
	} else if t2Foo != 1234 {
		t.Errorf("got bad t2:foo int value: %d", t2Foo)
	}

	if err := st.Get("t3", "foo", &t3Foo); err != nil {
		t.Errorf("error fetching t3:foo struct: %s", err)
	} else if !reflect.DeepEqual(t3Foo, val) {
		t.Errorf("got bad t3:foo struct value: %v", t3Foo)
	}

	if err := st.Get("t4", "foo", &t4Foo); err != nil {
		t.Errorf("error fetching t4:foo []byte: %s", err)
	} else if !reflect.DeepEqual(t4Foo, val2) {
		t.Errorf("got bad t4:foo []byte value: %v", t4Foo)
	}

	if err := st.Delete("t2", "foo"); err != nil {
		t.Errorf("cannot delete t2:foo: %s", err)
	}
	if exists, err := st.Has("t2", "foo"); err != nil {
		t.Errorf("error checking if t2:foo exists after delete: %s", err)
	} else if exists {
		t.Errorf("t2:foo still exists after delete")
	}
	t2Foo = -1
	if err := st.Get("t2", "foo", &t2Foo); err == nil {
		t.Errorf("can still get t2:foo after delete: %v", t2Foo)
	}

	val.Count = 999
	if err := st.Put("t3", "foo", &val); err != nil {
		t.Errorf("error updating t3:foo struct: %s", err)
	}

	t3Foo = Foo{}
	if err := st.Get("t3", "foo", &t3Foo); err != nil {
		t.Errorf("error fetching updated t3:foo struct: %s", err)
	} else if !reflect.DeepEqual(t3Foo, val) {
		t.Errorf("got bad updated t3:foo struct value: %v", t3Foo)
	}

	expKeys := []string{"foo"}
	if keys, err := st.GetKeysByPrefix("t1", ""); err != nil {
		t.Errorf("error fetching all t1 keys: %s", err)
	} else if !reflect.DeepEqual(keys, expKeys) {
		t.Errorf("got bad t1 keys: %v != %v", keys, expKeys)
	}

	if err := st.Put("t1", "boo", "aaahhh"); err != nil {
		t.Errorf("error storing t1:boo string: %s", err)
	}

	expKeys = []string{"boo", "foo"}
	if keys, err := st.GetKeysByPrefix("t1", ""); err != nil {
		t.Errorf("error fetching all t1 keys: %s", err)
	} else if !reflect.DeepEqual(keys, expKeys) {
		t.Errorf("got bad t1 keys: %v != %v", keys, expKeys)
	}

	expKeys = []string{"foo"}
	if keys, err := st.GetKeysByPrefix("t1", "f"); err != nil {
		t.Errorf("error fetching t1 keys starting with 'f': %s", err)
	} else if !reflect.DeepEqual(keys, expKeys) {
		t.Errorf("got bad t1 keys: %v != %v", keys, expKeys)
	}
}

func TestKVStoreSQLOps_Client(t *testing.T) {
	runWithDatabase(t, true, testKVStoreSQLOps)
}

func testKVStoreSQLInvalidOps(t *testing.T, st *KVStoreSQL) {
	if err := st.Put("", "foo", "hello"); err == nil {
		t.Errorf("Put did not fail on empty table name")
	}
	if err := st.Put("ttt", "", "hello"); err == nil {
		t.Errorf("Put did not fail on empty key")
	}
	if err := st.Put("ttt", "foo", nil); err == nil {
		t.Errorf("Put did not fail on nil value")
	}
	if err := st.Put("a|b", "foo", "hello"); err == nil {
		t.Errorf("Put did not fail on bad table name")
	}

	if err := st.Put("ttt", "foo", "hello"); err != nil {
		t.Errorf("valid Put failed: %s", err)
	}

	var out string

	if err := st.Get("", "foo", &out); err == nil {
		t.Errorf("Get did not fail on empty table name")
	}
	if err := st.Get("ttt", "", &out); err == nil {
		t.Errorf("Get did not fail on empty key")
	}
	if err := st.Get("ttt", "foo", nil); err == nil {
		t.Errorf("Get did not fail on nil value")
	}
	if err := st.Get("a|b", "foo", &out); err == nil {
		t.Errorf("Get did not fail on bad table name")
	}

	if _, err := st.Has("", "foo"); err == nil {
		t.Errorf("Has did not fail on empty table name")
	}
	if _, err := st.Has("ttt", ""); err == nil {
		t.Errorf("Has did not fail on empty key")
	}
	if _, err := st.Has("a|b", "foo"); err == nil {
		t.Errorf("Has did not fail on bad table name")
	}

	if err := st.Delete("", "foo"); err == nil {
		t.Errorf("Delete did not fail on empty table name")
	}
	if err := st.Delete("ttt", ""); err == nil {
		t.Errorf("Delete did not fail on empty key")
	}
	if err := st.Delete("a|b", "foo"); err == nil {
		t.Errorf("Delete did not fail on bad table name")
	}

	if _, err := st.GetKeysByPrefix("", ""); err == nil {
		t.Errorf("GetKeysByPrefix did not fail on empty table name")
	}
	if _, err := st.GetKeysByPrefix("a|b", ""); err == nil {
		t.Errorf("GetKeysByPrefix did not fail on bad table name")
	}
}

func TestKVStoreSQLInvalidOps_Client(t *testing.T) {
	runWithDatabase(t, true, testKVStoreSQLInvalidOps)
}

func testKVStoreSQLTransactions(t *testing.T, st *KVStoreSQL) {
	var err error
	if err = st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err = st.Put("ttt", "bar", 90); err != nil {
		t.Errorf("cannot store bar: %s", err)
	}

	tx, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}
	if err = tx.Commit(); err != nil {
		t.Errorf("transaction commit with no changes failed: %s", err)
	}

	tx, err = st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start another transaction: %s", err)
	}

	if err = tx.Put("ttt", "foo", 60); err != nil {
		t.Errorf("error updating foo inside transaction: %s", err)
	}
	if err = tx.Put("ttt", "bar", 40); err != nil {
		t.Errorf("error updating bar inside transaction: %s", err)
	}

	var foo, bar int
	if err = tx.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("cannot get foo while transaction is open: %s", err)
	}
	if err = tx.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("cannot get bar while transaction is open: %s", err)
	}
	if foo != 60 || bar != 40 {
		t.Errorf("data changes not visible within transaction: %d, %d", foo, bar)
	}

	expKeys := []string{"bar", "foo"}
	if keys, err := tx.GetKeysByPrefix("ttt", ""); err != nil {
		t.Errorf("error fetching all keys inside transaction: %s", err)
	} else if !reflect.DeepEqual(keys, expKeys) {
		t.Errorf("got bad keys: %v != %v", keys, expKeys)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("transaction commit failed: %s", err)
	}

	foo = 0
	bar = 0
	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != 60 || bar != 40 {
		t.Errorf("wrong data after commit: %d, %d", foo, bar)
	}
}

func TestKVStoreSQLTransactions_Client(t *testing.T) {
	runWithDatabase(t, true, testKVStoreSQLTransactions)
}

func testKVStoreSQLCancelTransaction(t *testing.T, st *KVStoreSQL) {
	var err error
	if err = st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err = st.Put("ttt", "bar", 90); err != nil {
		t.Errorf("cannot store bar: %s", err)
	}

	tx, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	if err = tx.Put("ttt", "foo", 60); err != nil {
		t.Errorf("error updating foo inside transaction: %s", err)
	}
	if err = tx.Put("ttt", "bar", 40); err != nil {
		t.Errorf("error updating bar inside transaction: %s", err)
	}
	if err = tx.Put("ttt", "baz", 555); err != nil {
		t.Errorf("error inserting baz inside transaction: %s", err)
	}

	if exists, err := tx.Has("ttt", "foo"); err != nil {
		t.Errorf("error checking if foo exists inside transaction: %s", err)
	} else if !exists {
		t.Errorf("foo does not exist inside transaction")
	}
	if exists, err := tx.Has("ttt", "bar"); err != nil {
		t.Errorf("error checking if bar exists inside transaction: %s", err)
	} else if !exists {
		t.Errorf("bar does not exist inside transaction")
	}
	if exists, err := tx.Has("ttt", "baz"); err != nil {
		t.Errorf("error checking if baz exists inside transaction: %s", err)
	} else if !exists {
		t.Errorf("baz does not exist inside transaction")
	}

	var foo, bar, baz int
	if err := tx.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("error getting foo inside transaction: %s", err)
	}
	if err := tx.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("error getting bar inside transaction: %s", err)
	}
	if err := tx.Get("ttt", "baz", &baz); err != nil {
		t.Errorf("error getting baz inside transaction: %s", err)
	}
	if foo != 60 || bar != 40 || baz != 555 {
		t.Errorf("wrong data inside transaction: %d, %d, %d", foo, bar, baz)
	}

	if err := tx.Delete("ttt", "foo"); err != nil {
		t.Errorf("error deleting foo inside transaction: %s", err)
	}
	foo = 0
	if err := tx.Get("ttt", "foo", &foo); err == nil {
		t.Errorf("foo exists after delete inside transaction: %d", foo)
	}

	tx.Discard()

	foo = 0
	bar = 0
	baz = 0
	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != 10 || bar != 90 {
		t.Errorf("wrong data after discard: %d, %d", foo, bar)
	}

	if exists, err := st.Has("ttt", "baz"); err != nil {
		t.Errorf("error checking if baz exists after discard: %s", err)
	} else if exists {
		t.Errorf("baz exists after transaction discard")
	}
}

func TestKVStoreSQLCancelTransaction_Client(t *testing.T) {
	runWithDatabase(t, true, testKVStoreSQLCancelTransaction)
}

func testKVStoreSQLTransactionOverlap(t *testing.T, st *KVStoreSQL) {
	var err error
	if err = st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err = st.Put("ttt", "bar", 90); err != nil {
		t.Errorf("cannot store bar: %s", err)
	}

	tx1, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 1st transaction: %s", err)
	}

	ch1to2 := make(chan int)
	ch2to1 := make(chan int)

	go func() {
		tx2, err := st.OpenTransaction()
		if err != nil {
			t.Errorf("cannot start 2nd transaction: %s", err)
		}

		ch2to1 <- 1 // notify tx2 created
		<-ch1to2    // wait for tx1 puts

		if err := tx2.Put("ttt", "foo", 30); err != nil {
			t.Errorf("error updating foo inside tx2: %s", err)
		}
		if err := tx2.Put("ttt", "bar", 70); err != nil {
			t.Errorf("error updating bar inside tx2: %s", err)
		}

		<-ch1to2 // wait for tx1 commit

		if err := tx2.Commit(); err != nil {
			t.Errorf("cannot commit tx2: %s", err)
		}

		ch2to1 <- 4 // notify tx2 commit done
	}()

	<-ch2to1 // wait for tx2 creation

	if err := tx1.Put("ttt", "foo", 20); err != nil {
		t.Errorf("error updating foo inside tx1: %s", err)
	}
	if err := tx1.Put("ttt", "bar", 80); err != nil {
		t.Errorf("error updating bar inside tx1: %s", err)
	}

	ch1to2 <- 2 // notify tx1 puts done

	if err := tx1.Commit(); err != nil {
		t.Errorf("cannot commit tx1: %s", err)
	}

	ch1to2 <- 3 // notify tx1 commit done
	<-ch2to1    // wait for tx2 commit

	var foo, bar int
	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != 30 || bar != 70 {
		t.Errorf("wrong data after tx1 and tx2: %d, %d", foo, bar)
	}
}

// With a single SQLite connection configured, this test deadlocks because
// it is designed to force the need of concurrent progress on 2 transactions.
// func TestKVStoreSQLTransactionOverlap_Client(t *testing.T) {
//	 runWithDatabase(t, true, testKVStoreSQLTransactionOverlap)
// }

func testKVStoreSQLTransactionConflict(t *testing.T, st *KVStoreSQL) {
	var err error
	if err = st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err = st.Put("ttt", "bar", 90); err != nil {
		t.Errorf("cannot store bar: %s", err)
	}

	// Create 2 transactions.
	tx1, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 1st transaction: %s", err)
	}
	tx2, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start 2nd transaction: %s", err)
	}

	// Both transactions get foo & bar.
	var foo, bar int
	if err = tx1.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("cannot get foo in 1st transaction: %s", err)
	}
	if err = tx1.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("cannot get bar in 1st transaction: %s", err)
	}
	if foo != 10 || bar != 90 {
		t.Errorf("invalid values in 1st transaction: %d, %d", foo, bar)
	}

	if err = tx2.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("cannot get foo in 2nd transaction: %s", err)
	}
	if err = tx2.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("cannot get bar in 2nd transaction: %s", err)
	}
	if foo != 10 || bar != 90 {
		t.Errorf("invalid values in 2nd transaction: %d, %d", foo, bar)
	}

	// Both transactions update foo & bar and try to commit their changes.
	// Only one transaction should succeed and the other should fail.
	// Which one succeeds depends on the database implementation.
	if err = tx1.Put("ttt", "foo", 20); err != nil {
		t.Errorf("cannot update foo in 1st transaction: %s", err)
	}
	if err = tx1.Put("ttt", "bar", 80); err != nil {
		t.Errorf("cannot update bar in 1st transaction: %s", err)
	}

	errTx1 := tx1.Commit()

	// Note: In CockroachDB the error maybe reported earlier at the
	// first conflicting Put instead of being delayed till the Commit.
	cancelTx2 := false
	if err = tx2.Put("ttt", "foo", 30); err != nil {
		err = tx2.ConvertError(err)
		if err != ErrTxConflict {
			t.Errorf("cannot update foo in 2nd transaction: %s", err)
		} else {
			cancelTx2 = true
		}
	}

	errTx2 := ErrTxConflict
	if !cancelTx2 {
		if err = tx2.Put("ttt", "bar", 70); err != nil {
			t.Errorf("cannot update bar in 2nd transaction: %s", err)
		}

		errTx2 = tx2.Commit()
	}

	var expFoo, expBar int
	if errTx1 == nil && errTx2 == nil {
		t.Errorf("conflict was not detected, both transactions passed")
	} else if errTx1 == nil && errTx2 == ErrTxConflict {
		expFoo, expBar = 20, 80
	} else if errTx1 == ErrTxConflict && errTx2 == nil {
		expFoo, expBar = 30, 70
	} else {
		t.Logf("both transactions failed: tx1 %v, tx2 %v", errTx1, errTx2)
	}

	tx1.Discard()
	tx2.Discard()

	// The foo & bar values should be those of the successful transaction.
	if err = st.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("cannot get foo from store: %s", err)
	}
	if err = st.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("cannot get bar from store: %s", err)
	}

	if foo != expFoo || bar != expBar {
		t.Errorf("invalid final foo (%d) or bar (%d) values", foo, bar)
	}
}

// With a single SQLite connection configured, this test deadlocks because
// it is designed to force the need of concurrent progress on 2 transactions.
// func TestKVStoreSQLTransactionConflict_Client(t *testing.T) {
//	 runWithDatabase(t, true, testKVStoreSQLTransactionConflict)
// }

func testDalSqlChan(t *testing.T, st *KVStoreSQL) {
	var err error
	dal := NewDAL(st)

	cid := ctype.Hex2Cid("abcdef")
	peer := ctype.Hex2Addr("bcd123")
	token := utils.GetTokenInfoFromAddress(peer)
	now := time.Now().UTC()
	openrs := &rpc.OpenChannelResponse{}
	balance := &structs.OnChainBalance{}
	simplex := &rpc.SignedSimplexState{}
	ledger := ctype.Hex2Addr("6666666666666666666666666666666666666666")

	err = dal.InsertChan(cid, peer, token, ledger, 55,
		openrs, balance, 1, 2, 3, 4, simplex, simplex)
	if err != nil {
		t.Errorf("failed InsertChan: %v", err)
	}

	err = dal.InsertClosedChan(cid, peer, token, now, now)
	if err != nil {
		t.Errorf("failed InsertClosedChan: %v", err)
	}

	base, lastUsed, lastAcked, lastNacked, found, err := dal.GetChanSeqNums(cid)
	if err != nil {
		t.Errorf("failed GetChanSeqNums: %v", err)
	} else if !found {
		t.Errorf("GetChanSeqNums did not find entry")
	} else if base != 1 || lastUsed != 2 || lastAcked != 3 || lastNacked != 4 {
		t.Errorf("wrong seqnums: %d, %d, %d, %d",
			base, lastUsed, lastAcked, lastNacked)
	}

	err = dal.UpdateChanSeqNums(cid, 5, 6, 7, 8)
	if err != nil {
		t.Errorf("failed UpdateChanSeqNums: %v", err)
	}

	base, lastUsed, lastAcked, lastNacked, found, err = dal.GetChanSeqNums(cid)
	if err != nil {
		t.Errorf("failed GetChanSeqNums after update: %v", err)
	} else if !found {
		t.Errorf("GetChanSeqNums did not find entry after update")
	} else if base != 5 || lastUsed != 6 || lastAcked != 7 || lastNacked != 8 {
		t.Errorf("wrong seqnums after update: %d, %d, %d, %d",
			base, lastUsed, lastAcked, lastNacked)
	}

	_, _, _, _, _, _, _, found, err = dal.GetChanViewInfoByID(cid)
	if err != nil {
		t.Errorf("failed GetChanViewInfoByID: %v", err)
	}

	err = dal.DeleteChan(cid)
	if err != nil {
		t.Errorf("failed DeleteChan: %v", err)
	}

	err = dal.DeleteChan(cid)
	if err == nil {
		t.Errorf("DeleteChan did not fail")
	}
}

func TestDalSqlChan_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlChan)
}

func testDalSqlPay(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	payID := ctype.Hex2PayID("abcdef")
	payBytes := []byte{0, 1, 2, 3, 4}
	note := &any.Any{}
	cid := ctype.Hex2Cid("abcdef")

	err := dal.InsertPayment(payID, payBytes, nil, note, cid, 1, cid, 1)
	if err != nil {
		t.Errorf("failed InsertPayment: %v", err)
	}

	payIDs, err := dal.GetAllPayIDs()
	if err != nil {
		t.Errorf("failed GetAllPayIDs: %v", err)
	}
	if len(payIDs) != 1 || payIDs[0] != payID {
		t.Errorf("wrong pay IDs: %v", payIDs)
	}

	dest := ctype.Hex2Addr("bcd123")
	err = dal.InsertDelegatedPay(payID, dest, 5)
	if err != nil {
		t.Errorf("failed InsertPayDelegation: %v", err)
	}

	err = dal.DeleteDelegatedPay(payID)
	if err != nil {
		t.Errorf("failed DeleteDelegatedPay: %v", err)
	}

	err = dal.DeleteDelegatedPay(payID)
	if err == nil {
		t.Errorf("DeleteDelegatedPay did not fail")
	}
}

func TestDalSqlPay_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlPay)
}

func testDalSqlSecret(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	hash := "hello"
	preimg := "world"
	payid := ctype.Hex2PayID("abcdef")

	err := dal.InsertSecret(hash, preimg, payid)
	if err != nil {
		t.Errorf("failed InsertSecret: %v", err)
	}

	err = dal.DeleteSecret(hash)
	if err != nil {
		t.Errorf("failed DeleteSecret: %v", err)
	}

	err = dal.DeleteSecret(hash)
	if err == nil {
		t.Errorf("DeleteSecret did not fail")
	}

	err = dal.InsertSecret(hash, preimg, payid)
	if err != nil {
		t.Errorf("failed InsertSecret: %v", err)
	}

	hash2 := "hello2"
	preimg2 := "world2"
	err = dal.InsertSecret(hash2, preimg2, payid)
	if err != nil {
		t.Errorf("failed InsertSecret: %v", err)
	}

	secret, found, err := dal.GetSecret(hash)
	if err != nil {
		t.Errorf("failed GetSecret: %v", err)
	}
	if !found {
		t.Errorf("secret not found")
	}
	if secret != preimg {
		t.Errorf("secrets does not match")
	}

	_, found, err = dal.GetSecret(hash2)
	if err != nil {
		t.Errorf("failed GetSecret: %v", err)
	}
	if !found {
		t.Errorf("secret not found")
	}

	err = dal.DeleteSecretByPayID(payid)
	if err != nil {
		t.Errorf("failed DeleteSecretByPayID: %v", err)
	}

	_, found, err = dal.GetSecret(hash)
	if found {
		t.Errorf("found deleted secret")
	}

	_, found, err = dal.GetSecret(hash2)
	if found {
		t.Errorf("found deleted secret")
	}

	err = dal.DeleteSecretByPayID(payid)
	if err != nil {
		t.Errorf("failed DeleteSecretByPayID: %v", err)
	}

}

func TestDalSqlSecret_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlSecret)
}

func testDalSqlTcb(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	addr := ctype.Hex2Addr("bcd123")
	token := utils.GetTokenInfoFromAddress(addr)
	deposit := big.NewInt(567)

	err := dal.InsertTcb(addr, token, deposit)
	if err != nil {
		t.Errorf("failed InsertTcb: %v", err)
	}

	val, found, err := dal.GetTcbDeposit(addr, token)
	if err != nil {
		t.Errorf("failed GetTcbDeposit: %v", err)
	} else if !found {
		t.Errorf("GetTcbDeposit did not find entry")
	} else if val.Cmp(deposit) != 0 {
		t.Errorf("wrong deposit: %s != %s", val.String(), deposit.String())
	}
}

func TestDalSqlTcb_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlTcb)
}

func testDalSqlMonitor(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	event := "foobar"

	err := dal.InsertMonitor(event, 1234, 22, true)
	if err != nil {
		t.Errorf("failed InsertMonitor: %v", err)
	}

	blockNum, blockIdx, found, err := dal.GetMonitorBlock(event)
	if err != nil {
		t.Errorf("failed GetMonitorBlock: %v", err)
	} else if !found {
		t.Errorf("GetMonitorBlock did not find entry")
	} else if blockNum != 1234 || blockIdx != 22 {
		t.Errorf("wrong block: %d, %d", blockNum, blockIdx)
	}

	restart, found, err := dal.GetMonitorRestart(event)
	if err != nil {
		t.Errorf("failed GetMonitorRestart: %v", err)
	} else if !found {
		t.Errorf("GetMonitorRestart did not find entry")
	} else if !restart {
		t.Errorf("wrong restart: %t", restart)
	}

	err = dal.UpdateMonitorBlock(event, 5678, 11)
	if err != nil {
		t.Errorf("failed UpdateMonitorBlock: %v", err)
	}

	blockNum, blockIdx, found, err = dal.GetMonitorBlock(event)
	if err != nil {
		t.Errorf("failed GetMonitorBlock after update: %v", err)
	} else if !found {
		t.Errorf("GetMonitorBlock did not find entry after update")
	} else if blockNum != 5678 || blockIdx != 11 {
		t.Errorf("wrong block after update: %d, %d", blockNum, blockIdx)
	}

	err = dal.UpsertMonitorRestart(event, false)
	if err != nil {
		t.Errorf("failed UpsertMonitorRestart: %v", err)
	}

	restart, found, err = dal.GetMonitorRestart(event)
	if err != nil {
		t.Errorf("failed GetMonitorRestart after update: %v", err)
	} else if !found {
		t.Errorf("GetMonitorRestart did not find entry after update")
	} else if restart {
		t.Errorf("wrong restart: %t", restart)
	}

	event = "6666666666666666666666666666666666666666-withdraw"
	event2 := "6666666666666666666666666666666666666667-withdraw"
	event3 := "6666666666666666666666666666666666666666-deposit"
	event4 := "6666666666666666666666666666666666666667-deposit"
	err = dal.InsertMonitor(event, 111, 1, true)
	if err != nil {
		t.Errorf("failed InsertMonitor: %v", err)
	}
	err = dal.InsertMonitor(event2, 111, 1, true)
	if err != nil {
		t.Errorf("failed InsertMonitor: %v", err)
	}
	err = dal.InsertMonitor(event3, 111, 1, true)
	if err != nil {
		t.Errorf("failed InsertMonitor: %v", err)
	}
	err = dal.InsertMonitor(event4, 111, 1, false)
	if err != nil {
		t.Errorf("failed InsertMonitor: %v", err)
	}
	addrs, err := dal.GetMonitorAddrsByEventAndRestart("withdraw", true)
	if err != nil {
		t.Errorf("failed GetMonitorAddrsByEventAndRestart: %v", err)
	}
	if len(addrs) != 2 {
		t.Errorf("wrong address number: want(2), get(%d)", len(addrs))
	}
	addrs, err = dal.GetMonitorAddrsByEventAndRestart("deposit", true)
	if err != nil {
		t.Errorf("failed GetMonitorAddrsByEventAndRestart: %v", err)
	}
	if len(addrs) != 1 {
		t.Errorf("wrong address number: want(1), get(%d)", len(addrs))
	}
	addrs, err = dal.GetMonitorAddrsByEventAndRestart("dispute", true)
	if err != nil {
		t.Errorf("failed GetMonitorAddrsByEventAndRestart: %v", err)
	}
	if len(addrs) != 0 {
		t.Errorf("wrong address number: want(0), get(%d)", len(addrs))
	}
}

func TestDalSqlMonitor_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlMonitor)
}

func testDalSqlRouting(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	cid := ctype.Hex2Cid("abcdef")
	dest := ctype.Hex2Addr("bcd123")
	token := utils.GetTokenInfoFromAddress(dest)

	err := dal.UpsertRouting(dest, token, cid)
	if err != nil {
		t.Errorf("failed UpsertRouting: %v", err)
	}

	cid2, found, err := dal.GetRoutingCid(dest, token)
	if err != nil {
		t.Errorf("failed GetRoutingCid: %v", err)
	} else if !found {
		t.Errorf("GetRoutingCid did not find entry")
	} else if cid != cid2 {
		t.Errorf("wrong cid: %v, %v", cid2, cid)
	}
}

func TestDalSqlRouting_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlRouting)
}

func testDalSqlMessage(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	cid := ctype.Hex2Cid("abcdef")
	msg := &rpc.CelerMsg{}

	err := dal.InsertChanMessage(cid, 33, msg)
	if err != nil {
		t.Errorf("failed InsertChanMessage: %v", err)
	}

	msg2, found, err := dal.GetChanMessage(cid, 33)
	if err != nil {
		t.Errorf("failed GetChanMessage: %v", err)
	} else if !found {
		t.Errorf("GetChanMessage did not find entry")
	} else if !reflect.DeepEqual(msg2, msg) {
		t.Errorf("wrong msg: %v, %v", msg2, msg)
	}

	msg3, found, err := dal.GetChanMessage(cid, 789)
	if err != nil {
		t.Errorf("failed GetChanMessage for non-existing msg: %v", err)
	} else if found {
		t.Errorf("GetChanMessage found non-existing msg: %v", msg3)
	}

	err = dal.DeleteChanMessage(cid, 33)
	if err != nil {
		t.Errorf("failed DeleteChanMessage: %v", err)
	}

	err = dal.DeleteChanMessage(cid, 33)
	if err == nil {
		t.Errorf("DeleteChanMessage did not fail after delete")
	}

	err = dal.DeleteChanMessage(cid, 456)
	if err == nil {
		t.Errorf("DeleteChanMessage did not fail on non-existing msg")
	}

	msg4, found, err := dal.GetChanMessage(cid, 33)
	if err != nil {
		t.Errorf("failed GetChanMessage after delete: %v", err)
	} else if found {
		t.Errorf("GetChanMessage found msg after delete: %v", msg4)
	}
}

func TestDalSqlMessage_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlMessage)
}

func testDalSqlPeer(t *testing.T, st *KVStoreSQL) {
	dal := NewDAL(st)

	svr := "def456"

	for num := 0; num < 4; num++ {
		peer := ctype.Hex2Addr(fmt.Sprintf("abc123%d", num))
		var cids []ctype.CidType
		for i := 0; i < num; i++ {
			cid := ctype.Hex2Cid(fmt.Sprintf("abcdef%d", i))
			cids = append(cids, cid)
		}

		err := dal.InsertPeer(peer, svr, cids)
		if err != nil {
			t.Errorf("failed InsertPeer (%d): %v", num, err)
		}

		svr2, found, err := dal.GetPeerServer(peer)
		if err != nil {
			t.Errorf("failed GetPeerServer (%d): %v", num, err)
		} else if !found {
			t.Errorf("GetPeerServer (%d) did not find entry", num)
		} else if svr2 != svr {
			t.Errorf("wrong server (%d): %v, %v", num, svr2, svr)
		}

		cids2, found, err := dal.GetPeerCids(peer)
		if err != nil {
			t.Errorf("failed GetPeerCids (%d): %v", num, err)
		} else if !found {
			t.Errorf("GetPeerCids (%d) did not find entry:", num)
		} else if !reflect.DeepEqual(cids2, cids) {
			t.Errorf("wrong cids (%d): %v: %v", num, cids2, cids)
		}
	}
}

func TestDalSqlPeer_Client(t *testing.T) {
	runWithDatabase(t, true, testDalSqlPeer)
}

func TestStr2Time(t *testing.T) {
	goodTs := []string{
		"2019-12-11T23:09:11.09099Z",       // cockroachdb
		"2019-12-11 23:09:11.112291+00:00", // sqlite
	}

	for _, ts := range goodTs {
		if _, err := str2Time(ts); err != nil {
			t.Errorf("cannot parse %s: %v", ts, err)
		}
	}

	badTs := []string{
		"2019-12-11.23:09:11.09099Z",
		"2019-12-11.23:09:11.112291+00:00",
		"hello world",
	}

	for _, ts := range badTs {
		if timestamp, err := str2Time(ts); err == nil {
			t.Errorf("parse %s did not fail: %v", ts, timestamp)
		}
	}
}
