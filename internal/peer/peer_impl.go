package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/storage"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
)

// badgerRegistry implements Registry using storage.Store.
type badgerRegistry struct {
	store storage.Store
}

// NewRegistry creates a new BadgerDB-backed Peer Registry.
func NewRegistry(store storage.Store) Registry {
	return &badgerRegistry{store: store}
}

func (r *badgerRegistry) RegisterPeer(ctx context.Context, p *Info) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid peer info")
	}

	p.LastSeen = time.Now()
	encoded, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal peer info: %w", err)
	}

	key := []byte(fmt.Sprintf("peer:%s", p.ID))
	return r.store.Set(ctx, key, encoded)
}

func (r *badgerRegistry) UnregisterPeer(ctx context.Context, id string) error {
	key := []byte(fmt.Sprintf("peer:%s", id))
	return r.store.Delete(ctx, key)
}

func (r *badgerRegistry) GetPeer(ctx context.Context, id string) (*Info, error) {
	key := []byte(fmt.Sprintf("peer:%s", id))
	data, err := r.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrPeerNotFound
		}
		return nil, err
	}

	var p Info
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peer info: %w", err)
	}

	return &p, nil
}

func (r *badgerRegistry) ListPeers(ctx context.Context) ([]*Info, error) {
	var peers []*Info
	err := r.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("peer:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var p Info
				if err := json.Unmarshal(val, &p); err != nil {
					return err
				}
				peers = append(peers, &p)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return peers, nil
}
