package database

import (
	"errors"
	"fmt"
)

// ErrKeyNotFound is an error returned by Store implementations
// when a certain key is not found.
var ErrKeyNotFound = errors.New("key not found")

type (
	// Store is anything that can persist and retrieve the blockchain.
	// information.
	Store interface {
		Delete(k []byte) error
		Get([]byte) ([]byte, error)
		Put(k, v []byte) error
		// Seek can guarantee that provided key (k) and value (v) are the only valid until the next call to f.
		// Key and value slices should not be modified.
		Seek(k []byte, f func(k, v []byte))
		Close() error
	}

	// KeyPrefix is a constant byte added as a prefix for each key
	// stored.
	KeyPrefix uint8
)

// Bytes returns the bytes representation of KeyPrefix.
func (k KeyPrefix) Bytes() []byte {
	return []byte{byte(k)}
}

// NewStore creates storage with preselected in configuration database type.
func NewStore(dbType string, name string) (Store, error) {
	var store Store
	var err error
	switch dbType {
	case "badgerdb":
		store, err = NewBadgerDBStore(name)
	default:
		return nil, fmt.Errorf("unknown storage: %s", dbType)
	}

	return store, err
}
