package kv

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Store[V any] interface {
	Set(string, V) error
	Get(string) (V, error)
}

type record struct {
	Key   string `gorm:"primaryKey"`
	Value []byte
}

func NewSQLStore[V any](db *gorm.DB) (*SQLStore[V], error) {
	err := db.AutoMigrate(&record{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate record: %w", err)
	}

	return &SQLStore[V]{
		db: db,
	}, nil
}

type SQLStore[V any] struct {
	db *gorm.DB
}

func (s *SQLStore[V]) Set(key string, value *V) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	keyValue := &record{
		Key:   key,
		Value: jsonValue,
	}

	return s.db.Create(keyValue).Error
}

func (s *SQLStore[V]) Get(key string) (*V, error) {
	var keyValue record
	if err := s.db.First(&keyValue, "key=?", key).Error; err != nil {
		return nil, err
	}

	value := new(V)
	if err := json.Unmarshal(keyValue.Value, value); err != nil {
		return nil, err
	}

	return value, nil
}
