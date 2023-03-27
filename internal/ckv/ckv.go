package ckv

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type (
	Store interface {
		Put(key string, value []byte) error
		Get(key string) ([][]byte, error)
	}
	MemoryStore struct {
		mutex sync.RWMutex
		data  map[string][]byte
	}
	SQLStore struct {
		db *gorm.DB
	}
	record struct {
		gorm.Model
		Key   string
		Value datatypes.JSON
	}
)

func NewMemoryStore() Store {
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}

func (s *MemoryStore) Put(key string, value []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
	return nil
}

func (s *MemoryStore) Get(key string) ([][]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result [][]byte
	for k, v := range s.data {
		if strings.HasPrefix(k, key) {
			result = append(result, v)
		}
	}
	return result, nil
}

func NewSQLStore(db *gorm.DB) Store {
	err := db.AutoMigrate(&record{})
	if err != nil {
		panic(fmt.Errorf("error migrating records: %w", err))
	}
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) Put(key string, value []byte) error {
	return s.db.Create(&record{
		Key:   key,
		Value: value,
	}).Error
}

func (s *SQLStore) Get(key string) ([][]byte, error) {
	var records []*record
	err := s.db.
		Where("key LIKE ?", key+"%").
		Find(&records).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting records: %w", err)
	}

	var result [][]byte
	for _, r := range records {
		result = append(result, r.Value)
	}

	return result, nil
}

func NewCollection[T any](s Store, prefix string, value T) *Collection[T] {
	return &Collection[T]{
		store:  s,
		prefix: prefix,
	}
}

type Collection[T any] struct {
	store  Store
	prefix string
}

func (c *Collection[T]) Put(key []string, value T) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling value: %w", err)
	}
	return c.store.Put(getKey(c.prefix, key), b)
}

func (c *Collection[T]) Get(key []string) ([]T, error) {
	b, err := c.store.Get(getKey(c.prefix, key))
	if err != nil {
		return nil, fmt.Errorf("error getting value: %w", err)
	}

	var result []T
	for _, v := range b {
		value := new(T)
		err = json.Unmarshal(v, value)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling value: %w", err)
		}
		result = append(result, *value)
	}

	return result, nil
}

func getKey(prefix string, key []string) string {
	return fmt.Sprintf("%s/%s/", prefix, strings.Join(key, "/"))
}
