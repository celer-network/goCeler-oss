// Copyright 2018 Celer Network

package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Return a temporary store root directory for testing without creating it.
func storeRootDir() string {
	user, _ := user.Current()
	ts := time.Now().UnixNano() / 1000
	dir := fmt.Sprintf("store-%s-%d", user.Username, ts)
	return path.Join(os.TempDir(), dir)
}

func TestKVStoreLocalCreation(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	if _, err := os.Stat(dir); err == nil {
		t.Errorf("KVStoreLocal dir %s already exists", dir)
	}

	st, err := NewKVStoreLocal(dir, false)
	if err != nil {
		t.Errorf("Cannot create store with dir %s: %s", dir, err)
	}
	defer st.Close()

	if info, err := os.Stat(dir); err != nil {
		t.Errorf("KVStoreLocal dir %s was not created", dir)
	} else if !info.IsDir() {
		t.Errorf("KVStoreLocal dir %s is not a directory", dir)
	}
}

func TestKVStoreLocalReadOnly(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, err := NewKVStoreLocal(dir, true)
	if err == nil {
		t.Errorf("Store created even in read-only mode: %s", dir)
		st.Close()
	}

	st, err = NewKVStoreLocal(dir, false)
	if err != nil {
		t.Errorf("Cannot create store with dir %s: %s", dir, err)
	} else {
		st.Close()
	}

	st, err = NewKVStoreLocal(dir, true)
	if err != nil {
		t.Errorf("Cannot open existing store %s in read-only mode: %s", dir, err)
	} else {
		st.Close()
	}

	// Trying to modify the store should fail on a read-only store.
	if err := st.Put("t1", "foo", "bar"); err != ErrReadOnly {
		t.Errorf("wrong error for Put on read-only store: %v", err)
	}
	if err := st.Delete("t1", "foo"); err != ErrReadOnly {
		t.Errorf("wrong error for Delete on read-only store: %v", err)
	}
	if tx, err := st.OpenTransaction(); err != ErrReadOnly {
		t.Errorf("wrong error for OpenTransaction on read-only store: %v", err)
		if err == nil {
			tx.Discard()
		}
	}
}

func TestBadKVStoreLocalCreation(t *testing.T) {
	if st, err := NewKVStoreLocal("", false); err == nil {
		st.Close()
		t.Errorf("KVStoreLocal did not error on empty directory")
	}

	f, err := ioutil.TempFile("", "bad-store-file")
	if err != nil {
		t.Errorf("Cannot create a temp file for testing: %s", err)
	}
	fpath := f.Name()
	defer f.Close()
	defer os.Remove(fpath)

	if st, err := NewKVStoreLocal(fpath, false); err == nil {
		st.Close()
		t.Errorf("KVStoreLocal did not error on bad dir (a file) %s", fpath)
	}
}

func TestKVStoreLocalOps(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

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
}

func TestKVStoreLocalInvalidOps(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

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

func TestKVStoreLocalTransactions(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	if err := st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err := st.Put("ttt", "bar", 90); err != nil {
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

	if err := tx.Put("ttt", "foo", 60); err != nil {
		t.Errorf("error updating foo inside transaction: %s", err)
	}
	if err := tx.Put("ttt", "bar", 40); err != nil {
		t.Errorf("error updating bar inside transaction: %s", err)
	}

	var foo, bar int
	if err := st.Get("ttt", "foo", &foo); err != nil {
		t.Errorf("cannot get foo while transaction is open: %s", err)
	}
	if err := st.Get("ttt", "bar", &bar); err != nil {
		t.Errorf("cannot get bar while transaction is open: %s", err)
	}
	if foo != 10 || bar != 90 {
		t.Errorf("data changed before transaction commit: %d, %d", foo, bar)
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

func TestKVStoreLocalCancelTransaction(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	if err := st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err := st.Put("ttt", "bar", 90); err != nil {
		t.Errorf("cannot store bar: %s", err)
	}

	tx, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	if err := tx.Put("ttt", "foo", 60); err != nil {
		t.Errorf("error updating foo inside transaction: %s", err)
	}
	if err := tx.Put("ttt", "bar", 40); err != nil {
		t.Errorf("error updating bar inside transaction: %s", err)
	}
	if err := tx.Put("ttt", "baz", 555); err != nil {
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

func TestKVStoreLocalTransactionOverlap(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	if err := st.Put("ttt", "foo", 10); err != nil {
		t.Errorf("cannot store foo: %s", err)
	}
	if err := st.Put("ttt", "bar", 90); err != nil {
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

		if err := tx2.Put("ttt", "foo", 30); err != nil {
			t.Errorf("error updating foo inside tx2: %s", err)
		}
		if err := tx2.Put("ttt", "bar", 70); err != nil {
			t.Errorf("error updating bar inside tx2: %s", err)
		}

		<-ch1to2

		if err := tx2.Commit(); err != nil {
			t.Errorf("cannot commit tx2: %s", err)
		}

		close(ch2to1)
	}()

	if err := tx1.Put("ttt", "foo", 20); err != nil {
		t.Errorf("error updating foo inside tx1: %s", err)
	}
	if err := tx1.Put("ttt", "bar", 80); err != nil {
		t.Errorf("error updating bar inside tx1: %s", err)
	}

	if err := tx1.Commit(); err != nil {
		t.Errorf("cannot commit tx1: %s", err)
	}

	close(ch1to2)
	<-ch2to1

	var foo, bar int
	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != 30 || bar != 70 {
		t.Errorf("wrong data after tx1 and tx2: %d, %d", foo, bar)
	}
}

func TestKVStoreLocalInvalidTransactionOps(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	tx, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	if err := tx.Put("", "foo", "hello"); err == nil {
		t.Errorf("Tx Put did not fail on empty table name")
	}
	if err := tx.Put("ttt", "", "hello"); err == nil {
		t.Errorf("Tx Put did not fail on empty key")
	}
	if err := tx.Put("ttt", "foo", nil); err == nil {
		t.Errorf("Tx Put did not fail on nil value")
	}
	if err := tx.Put("a|b", "foo", "hello"); err == nil {
		t.Errorf("Tx Put did not fail on bad table name")
	}

	if err := tx.Put("ttt", "foo", "hello"); err != nil {
		t.Errorf("valid Tx Put failed: %s", err)
	}

	var out string

	if err := tx.Get("", "foo", &out); err == nil {
		t.Errorf("Tx Get did not fail on empty table name")
	}
	if err := tx.Get("ttt", "", &out); err == nil {
		t.Errorf("Tx Get did not fail on empty key")
	}
	if err := tx.Get("ttt", "foo", nil); err == nil {
		t.Errorf("Tx Get did not fail on nil value")
	}
	if err := tx.Get("a|b", "foo", &out); err == nil {
		t.Errorf("Tx Get did not fail on bad table name")
	}

	if _, err := tx.Has("", "foo"); err == nil {
		t.Errorf("Tx Has did not fail on empty table name")
	}
	if _, err := tx.Has("ttt", ""); err == nil {
		t.Errorf("Tx Has did not fail on empty key")
	}
	if _, err := tx.Has("a|b", "foo"); err == nil {
		t.Errorf("Tx Has did not fail on bad table name")
	}

	if err := tx.Delete("", "foo"); err == nil {
		t.Errorf("Tx Delete did not fail on empty table name")
	}
	if err := tx.Delete("ttt", ""); err == nil {
		t.Errorf("Tx Delete did not fail on empty key")
	}
	if err := tx.Delete("a|b", "foo"); err == nil {
		t.Errorf("Tx Delete did not fail on bad table name")
	}

	if _, err := tx.GetKeysByPrefix("", ""); err == nil {
		t.Errorf("Tx GetKeysByPrefix did not fail on empty table name")
	}
	if _, err := tx.GetKeysByPrefix("a|b", ""); err == nil {
		t.Errorf("Tx GetKeysByPrefix did not fail on bad table name")
	}

	tx.Discard()
	tx.Discard() // another discard is a NOP

	if err := tx.Put("ttt", "foo", "hello"); err != ErrTxInvalid {
		t.Errorf("Tx Put did not detect invalid Tx: %s", err)
	}
	if err := tx.Get("ttt", "foo", &out); err != ErrTxInvalid {
		t.Errorf("Tx Get did not detect invalid Tx: %s", err)
	}
	if err := tx.Delete("ttt", "foo"); err != ErrTxInvalid {
		t.Errorf("Tx Delete did not detect invalid Tx: %s", err)
	}
	if _, err := tx.Has("ttt", "foo"); err != ErrTxInvalid {
		t.Errorf("Tx Has did not detect invalid Tx: %s", err)
	}
	if _, err := tx.GetKeysByPrefix("ttt", ""); err != ErrTxInvalid {
		t.Errorf("Tx GetKeysByPrefix did not detect invalid Tx: %s", err)
	}
	if err := tx.Commit(); err != ErrTxInvalid {
		t.Errorf("Tx Commit did not detect invalid Tx: %s", err)
	}
}

func TestKVStoreLocalRecovery(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)

	st.Put("ttt", "foo", "hello")
	st.Put("ttt", "bar", 1234)

	st.Close()

	// Delete the LevelDB manifest files (a corruption).
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Errorf("cannot list files in store dir: %s", err)
	}
	for _, f := range files {
		fname := f.Name()
		if strings.HasPrefix(fname, "MANIFEST") {
			os.Remove(path.Join(dir, fname))
		}
	}

	// Re-open the store, which triggers a recovery.
	st, _ = NewKVStoreLocal(dir, false)

	var foo string
	var bar int
	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != "hello" || bar != 1234 {
		t.Errorf("wrong data after store recovery: %s, %d", foo, bar)
	}

	st.Close()

	// Re-open the store in read-only mode and read the values.
	st, _ = NewKVStoreLocal(dir, true)
	defer st.Close()

	st.Get("ttt", "foo", &foo)
	st.Get("ttt", "bar", &bar)
	if foo != "hello" || bar != 1234 {
		t.Errorf("wrong data in read-only mode: %s, %d", foo, bar)
	}
}

func TestKVStoreLocalKeyIter(t *testing.T) {
	dir := storeRootDir()
	defer os.RemoveAll(dir)

	st, _ := NewKVStoreLocal(dir, false)
	defer st.Close()

	st.Put("ttt", "foo", 123)
	st.Put("ttt", "bar", 456)
	st.Put("ttt", "baz", 789)

	keys, err := st.GetKeysByPrefix("xyz", "")
	if err != nil {
		t.Errorf("getkeys on empty table failed: %s", err)
	} else if len(keys) != 0 {
		t.Errorf("getkeys on empty table returned data: %v", keys)
	}

	keys, err = st.GetKeysByPrefix("t", "")
	if err != nil {
		t.Errorf("getkeys on empty subname table failed: %s", err)
	} else if len(keys) != 0 {
		t.Errorf("getkeys on empty subname table returned data: %v", keys)
	}

	exp := []string{"bar", "baz", "foo"}
	keys, err = st.GetKeysByPrefix("ttt", "")
	if err != nil {
		t.Errorf("getkeys on full table failed: %s", err)
	} else if !reflect.DeepEqual(keys, exp) {
		t.Errorf("getkeys on full table got %v instead of %v", keys, exp)
	}

	exp = []string{"bar", "baz"}
	keys, err = st.GetKeysByPrefix("ttt", "b")
	if err != nil {
		t.Errorf("getkeys on partial table failed: %s", err)
	} else if !reflect.DeepEqual(keys, exp) {
		t.Errorf("getkeys on partial table got %v instead of %v", keys, exp)
	}

	tx, err := st.OpenTransaction()
	if err != nil {
		t.Errorf("cannot start transaction: %s", err)
	}

	tx.Put("ttt", "barbar", 555)
	tx.Put("ttt", "foo", 888)
	tx.Delete("ttt", "baz")

	exp = []string{"bar", "barbar", "foo"}
	keys, err = tx.GetKeysByPrefix("ttt", "")
	if err != nil {
		t.Errorf("Tx getkeys on full table failed: %s", err)
	} else if !reflect.DeepEqual(keys, exp) {
		t.Errorf("Tx getkeys on full table got %v instead of %v", keys, exp)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("cannot commit transaction: %s", err)
	}

	keys, err = st.GetKeysByPrefix("ttt", "")
	if err != nil {
		t.Errorf("Tx getkeys on full table failed: %s", err)
	} else if !reflect.DeepEqual(keys, exp) {
		t.Errorf("Tx getkeys on full table got %v instead of %v", keys, exp)
	}
}
