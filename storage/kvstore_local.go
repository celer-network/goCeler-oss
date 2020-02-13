// Copyright 2018-2019 Celer Network
//
// This is a wrapper on top of a Go LevelDB implementation.
// It is a layer that hides the namespace partitioning of the key space
// into tables and handles initialization scenarios.

package storage

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/syndtr/goleveldb/leveldb"
	dberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	separator     = "|" // reserved character for keys construction
	versionPrefix = "v" // table prefix for tracking data versions
	noVersion     = ""  // handle old store data before versioning
)

var (
	ErrTxConflict = errors.New("Transaction conflict")
	ErrTxInvalid  = errors.New("Invalid transaction")
	ErrReadOnly   = errors.New("Cannot modify read-only store")
)

type KVStoreLocal struct {
	rootDir  string      // store root directory
	readOnly bool        // read-only mode
	db       *leveldb.DB // DB handle
}

type TransactionLocal struct {
	store *KVStoreLocal        // local store handle
	itx   *leveldb.Transaction // inner transaction
}

// Create a new local K/V store at the given root directory.
func NewKVStoreLocal(rootDir string, readOnly bool) (*KVStoreLocal, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("rootDir is not specified")
	}

	if mode, err := os.Stat(rootDir); err == nil && !mode.IsDir() {
		return nil, fmt.Errorf("rootDir %s is not a directory", rootDir)
	}

	var opts *opt.Options
	if readOnly {
		opts = &opt.Options{
			ErrorIfMissing: true,
			ReadOnly:       true,
		}
	}

	db, err := leveldb.OpenFile(rootDir, opts)
	if err != nil {
		if readOnly {
			return nil, err
		}

		// Try to recover from corrupted DB files.
		if dberrors.IsCorrupted(err) {
			db, err = leveldb.RecoverFile(rootDir, nil)
		}
		if err != nil {
			return nil, err
		}
	}

	s := &KVStoreLocal{
		rootDir:  rootDir,
		readOnly: readOnly,
		db:       db,
	}
	return s, nil
}

// Close the local K/V store.
func (s *KVStoreLocal) Close() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
		s.rootDir = ""
	}
}

// Marshal the data into a byte-array if it is not one already.
func marshal(val interface{}) ([]byte, error) {
	switch v := val.(type) {
	case []byte:
		return v, nil
	default:
		return json.Marshal(val)
	}
}

func unmarshal(src []byte, dest interface{}) error {
	switch v := dest.(type) {
	case *[]byte:
		*v = append([]byte(nil), src...)
		return nil
	default:
		return json.Unmarshal(src, dest)
	}
}

// Store a key/value pair within a table's namespace.
func (s *KVStoreLocal) Put(table, key string, value interface{}) error {
	if s.readOnly {
		return ErrReadOnly
	}
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	data, err := marshal(value)
	if err == nil {
		sKey := storeKey(table, key)
		err = s.db.Put([]byte(sKey), data, nil)
	}
	return err
}

// Extract the value of the given key within a table's namespace into
// the given variable.
func (s *KVStoreLocal) Get(table, key string, value interface{}) error {
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	sKey := storeKey(table, key)
	data, err := s.db.Get([]byte(sKey), nil)
	if err == nil {
		err = unmarshal(data, value)
	}
	return err
}

// Delete the entry for a key within a table's namespace.
func (s *KVStoreLocal) Delete(table, key string) error {
	if s.readOnly {
		return ErrReadOnly
	}
	if err := checkTableKey(table, key); err != nil {
		return err
	}

	sKey := storeKey(table, key)
	return s.db.Delete([]byte(sKey), nil)
}

// Check if an entry exists for the given key within a table's namespace.
func (s *KVStoreLocal) Has(table, key string) (bool, error) {
	if err := checkTableKey(table, key); err != nil {
		return false, err
	}

	sKey := storeKey(table, key)
	return s.db.Has([]byte(sKey), nil)
}

// Return all keys for a given table and key prefix. The key prefix
// can be the empty string, which returns all keys within the table.
func (s *KVStoreLocal) GetKeysByPrefix(table, prefix string) ([]string, error) {
	if err := checkTableKey(table, " "); err != nil {
		return nil, err
	}

	storePrefix := []byte(table + separator + prefix)
	iter := s.db.NewIterator(util.BytesPrefix(storePrefix), nil)
	return getKeysByIter(iter)
}

// Helper function to extract and return keys from a storage iterator.
func getKeysByIter(iter iterator.Iterator) ([]string, error) {
	var keys []string
	for iter.Next() {
		_, k := tableKey(iter.Key())
		keys = append(keys, k)
	}
	iter.Release()
	return keys, iter.Error()
}

// Check if the table and key parameters are valid.
func checkTableKey(table, key string) error {
	if table == "" || key == "" {
		return fmt.Errorf("table and key parameters must be specified")
	}

	// The separator character cannot be used in the table name.
	if strings.Contains(table, separator) {
		return fmt.Errorf("invalid table name: %s", table)
	}
	return nil
}

// storeKey returns the store's key for the given table and entry key.
func storeKey(table, key string) string {
	return table + separator + key
}

// tableKey returns the user visible (table, key) info from a store key.
func tableKey(skey []byte) (string, string) {
	parts := strings.SplitN(string(skey), separator, 2)
	return parts[0], parts[1]
}

// keysHash computes a hash for a list of keys.
func keysHash(keys []string) (string, error) {
	var hash string
	data, err := json.Marshal(keys)
	if err == nil {
		hash = fmt.Sprintf("%x", md5.Sum(data))
	}
	return hash, err
}

// Start a store transaction.
func (s *KVStoreLocal) OpenTransaction() (Transaction, error) {
	if s.readOnly {
		return nil, ErrReadOnly
	}

	itx, err := s.db.OpenTransaction()
	if err != nil {
		return nil, err
	}

	tx := &TransactionLocal{
		store: s,
		itx:   itx,
	}
	return tx, nil
}

// Discard a transaction.
func (tx *TransactionLocal) Discard() {
	if tx.store == nil {
		return
	}

	if tx.itx != nil {
		tx.itx.Discard()
	}

	tx.store = nil
	tx.itx = nil
}

func tsNow() int64 {
	return time.Now().UnixNano() / 1000 // usec
}

func (tx *TransactionLocal) ConvertError(err error) error {
	return err
}

// Commit a transaction.
func (tx *TransactionLocal) Commit() error {
	if tx.store == nil || tx.itx == nil {
		log.Traceln("commit: invalid transaction")
		return ErrTxInvalid
	}

	defer tx.Discard()
	return tx.itx.Commit()
}

// In a transaction, store a key/value pair within a table's namespace.
func (tx *TransactionLocal) Put(table, key string, value interface{}) error {
	if tx.store == nil || tx.itx == nil {
		return ErrTxInvalid
	}
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	data, err := marshal(value)
	if err == nil {
		sKey := storeKey(table, key)
		err = tx.itx.Put([]byte(sKey), data, nil)
	}
	return err
}

// In a transaction, extract the value of the given key within a table's
// namespace into the given variable.
func (tx *TransactionLocal) Get(table, key string, value interface{}) error {
	if tx.store == nil || tx.itx == nil {
		return ErrTxInvalid
	}
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	sKey := storeKey(table, key)
	data, err := tx.itx.Get([]byte(sKey), nil)
	if err == nil {
		err = unmarshal(data, value)
	}
	return err
}

// In a transaction, delete the entry for a key within a table's namespace.
func (tx *TransactionLocal) Delete(table, key string) error {
	if tx.store == nil || tx.itx == nil {
		return ErrTxInvalid
	}
	if err := checkTableKey(table, key); err != nil {
		return err
	}

	sKey := storeKey(table, key)
	return tx.itx.Delete([]byte(sKey), nil)
}

// In a transaction, check if an entry exists for the given key within
// a table's namespace.
func (tx *TransactionLocal) Has(table, key string) (bool, error) {
	if tx.store == nil || tx.itx == nil {
		return false, ErrTxInvalid
	}
	if err := checkTableKey(table, key); err != nil {
		return false, err
	}

	sKey := storeKey(table, key)
	return tx.itx.Has([]byte(sKey), nil)
}

// In a transaction, return all keys for a given table and key prefix.
// The key prefix can be the empty string, which returns all keys within
// the table.
func (tx *TransactionLocal) GetKeysByPrefix(table, prefix string) ([]string, error) {
	if tx.store == nil || tx.itx == nil {
		return nil, ErrTxInvalid
	}
	if err := checkTableKey(table, " "); err != nil {
		return nil, err
	}

	// Fetch the matching keys from the DB.
	storePrefix := table + separator + prefix
	iter := tx.itx.NewIterator(util.BytesPrefix([]byte(storePrefix)), nil)
	return getKeysByIter(iter)
}
