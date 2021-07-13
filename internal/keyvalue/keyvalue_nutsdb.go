package keyvalue

import (
	"errors"

	"github.com/xujiajun/nutsdb"
)

// NewNutsDBStore returns a new Store that uses *nutsdb.DB as the underlying
// storage.
func NewNutsDBStore(db *nutsdb.DB, bucket string) Store {
	return &dbStore{
		db:     db,
		bucket: bucket,
	}
}

type dbStore struct {
	db     *nutsdb.DB
	bucket string
}

// Get gets the value for the given key.
// If the key does not exist, then Get returns ErrNotFound.
func (s *dbStore) Get(key string) (value []byte, err error) {
	err = s.db.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(s.bucket, []byte(key))
		if err != nil {
			return err
		}
		value = entry.Value
		return nil
	})
	if errors.Is(err, nutsdb.ErrNotFoundKey) {
		return nil, ErrNotFound
	}
	return value, err
}

// Set sets the value for the given key.
func (s *dbStore) Set(key string, value []byte) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(s.bucket, []byte(key), value, 0)
	})
}

// Delete deletes the value for the given key.
func (s *dbStore) Delete(key string) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(s.bucket, []byte(key))
	})
}

// Iter calls f for each key-value pair in the store.
func (s *dbStore) Iter(f func(key string, value []byte)) {
	s.db.View(func(tx *nutsdb.Tx) error {
		entries, _, err := tx.PrefixScan(s.bucket, []byte{}, 0, 0)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			f(string(entry.Key), entry.Value)
		}
		return nil
	})
}

// Close closes the store.
func (s *dbStore) Close() error {
	return s.db.Close()
}
