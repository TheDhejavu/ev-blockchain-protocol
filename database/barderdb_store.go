package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
	logger "github.com/sirupsen/logrus"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// root folder of this project
	root = filepath.Join(filepath.Dir(b), "../")
)

// BadgerDBStore is the official storage implementation for storing and retrieving
// blockchain data.
type BadgerDBStore struct {
	db *badger.DB
}

// NewBadgerDBStore returns a new BadgerDBStore object that will
// initialize the database found at the given path.
func NewBadgerDBStore(name string) (*BadgerDBStore, error) {
	// BadgerDB isn't able to make nested directories
	path := getDatabasePath(name)
	// Open the Badger database located in the /storage/blocks directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions(path)
	opts.ValueDir = path
	db, err := openDB(path, opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDBStore{
		db: db,
	}, nil
}

// Delete implements the Store interface.
func (b *BadgerDBStore) Delete(key []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Get implements the Store interface.
func (b *BadgerDBStore) Get(key []byte) ([]byte, error) {
	var val []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

// Put implements the Store interface.
func (b *BadgerDBStore) Put(key, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

// Seek implements the Store interface.
func (b *BadgerDBStore) Seek(key []byte, f func(k, v []byte)) {
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   100,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         key,
			InternalAccess: false,
		})
		defer it.Close()
		for it.Seek(key); it.ValidForPrefix(key); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			f(k, v)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

// Close releases all db resources.
func (b *BadgerDBStore) Close() error {
	return b.db.Close()
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				logger.Panicln("database unlocked , value log truncated ")
				return db, nil
			}
			logger.Panicln("could not unlock database", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func getDatabasePath(name string) string {
	if name != "" {
		return filepath.Join(root, fmt.Sprintf("./storage/blocks_%s", name))
	}
	return filepath.Join(root, "./storage/blocks")
}

// Check if Blockchain Database already exist
func databaseExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveDatabase(name string) error {
	path := getDatabasePath(name)
	err := os.RemoveAll(path)
	if err != nil {
		logger.Error("Remove Error occurred:", err)
		return err
	}
	return nil
}
func openBardgerDB(name string) *badger.DB {
	path := getDatabasePath(name)

	opts := badger.DefaultOptions(path)
	db, err := openDB(path, opts)
	if err != nil {
		logger.Panic(err)
	}

	return db
}
