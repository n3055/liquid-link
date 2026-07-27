package storage

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/config"
)

// Store defines core operations supported by the BadgerDB wrapper.
type Store interface {
	// Set writes a key-value pair to the database.
	Set(ctx context.Context, key, value []byte) error

	// Get retrieves a value for the given key from the database.
	Get(ctx context.Context, key []byte) ([]byte, error)

	// Delete removes a key-value pair.
	Delete(ctx context.Context, key []byte) error

	// Close terminates the storage connection.
	Close() error

	// View executes a read-only transaction.
	View(fn func(txn *badger.Txn) error) error

	// Update executes a read-write transaction.
	Update(fn func(txn *badger.Txn) error) error
}

// BadgerStore implements the Store interface using BadgerDB.
type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore initializes and opens a BadgerDB instance.
func NewBadgerStore(cfg *config.StorageConfig) (*BadgerStore, error) {
	opts := badger.DefaultOptions(cfg.Path)
	opts.ValueLogFileSize = cfg.ValueLogFileSize

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

// Set writes a key-value pair to BadgerDB.
func (s *BadgerStore) Set(ctx context.Context, key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves a key's value from BadgerDB.
func (s *BadgerStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	var val []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	return val, nil
}

// Delete removes a key from BadgerDB.
func (s *BadgerStore) Delete(ctx context.Context, key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close closes the BadgerDB database.
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// View executes a read-only transaction inside BadgerDB.
func (s *BadgerStore) View(fn func(txn *badger.Txn) error) error {
	return s.db.View(fn)
}

// Update executes a read-write transaction inside BadgerDB.
func (s *BadgerStore) Update(fn func(txn *badger.Txn) error) error {
	return s.db.Update(fn)
}
