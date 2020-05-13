// Copyright 2019-2020 Celer Network
//
// Support the KVStore interface using a SQL database server.

package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"
	"time"

	"github.com/celer-network/goutils/log"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tevino/abool"
)

const (
	dbPingPolling = 1 * time.Minute
	// sql driver does dynamic conn pooling and is agressive open/closing connections,
	// causing unnecessary churns in high concurrent scenario. we can adjust the value
	// in the future if db tx latency is high due to queued tx
	maxIdleConns = 50
	maxOpenConns = 50
)

var (
	ErrNilValue = errors.New("Value cannot be nil")
)

type KVStoreSQL struct {
	driver string            // database driver
	info   string            // database connection info
	crdb   bool              // database is CockroachDB
	db     *sql.DB           // database access object
	quit   chan bool         // quit background goroutines (e.g. dbPing)
	closed *abool.AtomicBool // set to true in Close
}

type TransactionSQL struct {
	store *KVStoreSQL // remote store handle
	dbTx  *sql.Tx     // database transaction
}

type dbOrTx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func exists(fpath string) (bool, error) {
	_, err := os.Stat(fpath)
	if err == nil || os.IsExist(err) {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Create a new remote K/V store.
func NewKVStoreSQL(driver, info string) (*KVStoreSQL, error) {
	s := &KVStoreSQL{
		driver: driver,
		info:   info,
		crdb:   true,
		quit:   make(chan bool),
		closed: abool.New(),
	}

	// Special check for SQLite on the client: if the file
	// does not already exist, then initialize its schema.
	// Note: in the "sqlite3" case "info" is the file path.
	initSchema := false
	if driver == "sqlite3" {
		s.crdb = false
		if ok, err := exists(info); err != nil {
			log.Debugln("NewKVStoreSQL: cannot Stat() file:", info, err)
			return nil, err
		} else if !ok {
			initSchema = true
			dir := path.Dir(info)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Debugln("NewKVStoreSQL: cannot create dir:", dir, err)
				return nil, err
			}
		}
		info += "?cache=shared"
	}

	db, err := sql.Open(driver, info)
	if err != nil {
		log.Debugln("NewKVStoreSQL: sql.Open() failed:", info, err)
		return nil, err
	}

	s.db = db

	// Initialize the database schema if needed.
	if initSchema {
		for _, cmd := range sqlSchemaCmds {
			_, err = db.Exec(cmd)
			if err != nil {
				db.Close()
				return nil, err
			}
		}
	}

	// For CockroachDB start a background DB connection pinger.
	if driver == "postgres" {
		go s.dbPing(s.db)
		s.db.SetMaxIdleConns(maxIdleConns)
		s.db.SetMaxOpenConns(maxOpenConns)
	} else if driver == "sqlite3" {
		// force single conn to serialze sqlite ops, to avoid `database is locked`
		// because sqlite doesn't support concurrent write
		// the downside is read ops are also affected
		s.db.SetMaxOpenConns(1)
	}

	return s, nil
}

// Periodically Ping() the DB connection and log any errors.
// The Ping() recreates DB connections if needed.
// Note: the db object is passed as a parameter instead of using
// "s.db" to avoid concurrent read/write between dbPing() and Close().
func (s *KVStoreSQL) dbPing(db *sql.DB) {
	ticker := time.NewTicker(dbPingPolling)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			log.Debugln("dbPing: quit")
			return

		case <-ticker.C:
			if err := db.Ping(); err != nil {
				log.Errorln("dbPing:", err)
			} else {
				log.Traceln("dbPing: OK")
			}
		}
	}
}

// Close the remote K/V store.
func (s *KVStoreSQL) Close() {
	if s.closed.IsSet() {
		log.Warn("db.Close called multiple times!")
		return
	}
	s.closed.Set()
	close(s.quit) // we could use s.closed to signal all goroutines
	s.db.Close()
	// previously we also set s.db to nil, but had database/sql.(*DB).conn nil panic, suspect some polling still tries to access db
	// so we just call db.Close without set s.db to nil. Future .conn calls will return errDBClosed
}

func (s *KVStoreSQL) put(db dbOrTx, table, key string, value interface{}) error {
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return ErrNilValue
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	q := `INSERT INTO keyvals (key, tbl, val) VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET val = excluded.val`
	_, err = db.Exec(q, storeKey(table, key), table, data)
	return err
}

func (s *KVStoreSQL) get(db dbOrTx, table, key string, value interface{}) error {
	if err := checkTableKey(table, key); err != nil {
		return err
	}
	if value == nil {
		return ErrNilValue
	}

	var data []byte
	q := "SELECT val FROM keyvals WHERE key = $1"
	err := db.QueryRow(q, storeKey(table, key)).Scan(&data)
	if err == nil {
		err = json.Unmarshal(data, value)
	}
	return err
}

func (s *KVStoreSQL) del(db dbOrTx, table, key string) error {
	if err := checkTableKey(table, key); err != nil {
		return err
	}

	q := "DELETE FROM keyvals WHERE key = $1"
	_, err := db.Exec(q, storeKey(table, key))
	return err
}

func (s *KVStoreSQL) has(db dbOrTx, table, key string) (bool, error) {
	if err := checkTableKey(table, key); err != nil {
		return false, err
	}

	var data int
	q := "SELECT 1 FROM keyvals WHERE key = $1"
	err := db.QueryRow(q, storeKey(table, key)).Scan(&data)
	if err == nil {
		return true, nil
	} else if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *KVStoreSQL) getKeys(db dbOrTx, table, prefix string) ([]string, error) {
	if err := checkTableKey(table, " "); err != nil {
		return nil, err
	}

	var params []interface{}

	// For an empty prefix the query uses only the indexed "tbl" table.
	q := "SELECT key FROM keyvals WHERE tbl = $1"
	params = append(params, table)

	if prefix != "" {
		// Further filtering on the keys using LIKE prefix matching.
		q += " AND key LIKE $2"
		like := storeKey(table, prefix) + "%"
		params = append(params, like)
	}
	q += " ORDER BY key"

	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			return nil, err
		}

		_, key = tableKey([]byte(key))
		keys = append(keys, key)
	}

	return keys, nil
}

func (s *KVStoreSQL) Put(table, key string, value interface{}) error {
	return s.put(s.db, table, key, value)
}

func (s *KVStoreSQL) Get(table, key string, value interface{}) error {
	return s.get(s.db, table, key, value)
}

func (s *KVStoreSQL) Delete(table, key string) error {
	return s.del(s.db, table, key)
}

func (s *KVStoreSQL) Has(table, key string) (bool, error) {
	return s.has(s.db, table, key)
}

func (s *KVStoreSQL) GetKeysByPrefix(table, prefix string) ([]string, error) {
	return s.getKeys(s.db, table, prefix)
}

func (s *KVStoreSQL) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *KVStoreSQL) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

func (s *KVStoreSQL) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

func (s *KVStoreSQL) OpenTransaction() (Transaction, error) {
	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	if s.crdb {
		_, err = dbTx.Exec("SAVEPOINT cockroach_restart")
		if err != nil {
			dbTx.Rollback()
			return nil, err
		}
	}

	tx := &TransactionSQL{
		store: s,
		dbTx:  dbTx,
	}
	return tx, nil
}

func (tx *TransactionSQL) Discard() {
	if tx.dbTx != nil {
		err := tx.dbTx.Rollback()
		if err == nil {
			tx.dbTx = nil
		}
	}
}

func (tx *TransactionSQL) ConvertError(err error) error {
	if err == nil {
		return nil
	}

	// Special re-mapping of this error back to transaction conflict.
	var patterns []string
	if tx.store.crdb {
		patterns = []string{"retry transaction", "restart transaction",
			"current transaction is aborted", "40001", "cr000"}
	} else {
		patterns = []string{"database is locked"}
	}

	errMsg := strings.ToLower(err.Error())
	for _, pat := range patterns {
		if strings.Contains(errMsg, pat) {
			return ErrTxConflict
		}
	}

	return err
}

func (tx *TransactionSQL) Commit() error {
	var err error
	if tx.store.crdb {
		// For CockroachDB, both "release savepoint" and the follow-up
		// "commit" may fail.  The commit after a successful "release"
		// is not a NOP.
		_, err = tx.dbTx.Exec("RELEASE SAVEPOINT cockroach_restart")
	}

	if err == nil {
		err = tx.dbTx.Commit()
		if err == nil {
			tx.dbTx = nil
			return nil
		}
	}

	return tx.ConvertError(err)
}

func (tx *TransactionSQL) Put(table, key string, value interface{}) error {
	return tx.store.put(tx.dbTx, table, key, value)
}

func (tx *TransactionSQL) Get(table, key string, value interface{}) error {
	return tx.store.get(tx.dbTx, table, key, value)
}

func (tx *TransactionSQL) Delete(table, key string) error {
	return tx.store.del(tx.dbTx, table, key)
}

func (tx *TransactionSQL) Has(table, key string) (bool, error) {
	return tx.store.has(tx.dbTx, table, key)
}

func (tx *TransactionSQL) GetKeysByPrefix(table, prefix string) ([]string, error) {
	return tx.store.getKeys(tx.dbTx, table, prefix)
}

func (tx *TransactionSQL) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.dbTx.Exec(query, args...)
}

func (tx *TransactionSQL) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.dbTx.Query(query, args...)
}

func (tx *TransactionSQL) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.dbTx.QueryRow(query, args...)
}
