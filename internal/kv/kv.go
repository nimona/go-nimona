package kv

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Store[K, V any] interface {
	Set(K, *V) error
	Get(K) (*V, error)
}

type record struct {
	Key   []byte `gorm:"primaryKey"`
	Value []byte
}

func NewSQLStore[K, V any](db *gorm.DB) (Store[K, V], error) {
	err := db.AutoMigrate(&record{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate record: %w", err)
	}

	return &SQLStore[K, V]{
		db: db,
	}, nil
}

type SQLStore[K, V any] struct {
	db *gorm.DB
}

func (s *SQLStore[K, V]) Set(key K, value *V) error {
	jsonKey, err := json.Marshal(key)
	if err != nil {
		return err
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	keyValue := &record{
		Key:   jsonKey,
		Value: jsonValue,
	}

	return s.db.Create(keyValue).Error
}

func (s *SQLStore[K, V]) Get(key K) (*V, error) {
	jsonKey, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}

	var keyValue record
	if err := s.db.First(&keyValue, "key=?", jsonKey).Error; err != nil {
		return nil, err
	}

	value := new(V)
	if err := json.Unmarshal(keyValue.Value, value); err != nil {
		return nil, err
	}

	return value, nil
}
