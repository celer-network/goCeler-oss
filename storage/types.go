// Copyright 2019 Celer Network

package storage

const (
	NoTxID = uint32(0) // reserved for non-transactional operations
)

// KVStore is the interface implemented by the local store (LevelDB
// wrapper) and by the remote store (gRPC calls to a store server).
type KVStore interface {
	Close()
	OpenTransaction() (Transaction, error)
	Put(table, key string, value interface{}) error
	Get(table, key string, value interface{}) error
	Delete(table, key string) error
	Has(table, key string) (bool, error)
	GetKeysByPrefix(table, prefix string) ([]string, error)
}

// Transaction is the interface implemented by the local and remote stores.
type Transaction interface {
	Commit() error
	Discard()
	ConvertError(err error) error
	Put(table, key string, value interface{}) error
	Get(table, key string, value interface{}) error
	Delete(table, key string) error
	Has(table, key string) (bool, error)
	GetKeysByPrefix(table, prefix string) ([]string, error)
}
