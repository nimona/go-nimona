package kv

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type record struct {
	Key   string `gorm:"primaryKey"`
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
	keyString := keyToString(key)
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	keyValue := &record{
		Key:   keyString,
		Value: jsonValue,
	}

	return s.db.Create(keyValue).Error
}

func (s *SQLStore[K, V]) Get(key K) (*V, error) {
	keyString := keyToString(key)
	var keyValue record
	if err := s.db.First(&keyValue, "key=?", keyString).Error; err != nil {
		return nil, err
	}

	value := new(V)
	if err := json.Unmarshal(keyValue.Value, value); err != nil {
		return nil, err
	}

	return value, nil
}

func (s *SQLStore[K, V]) GetPrefix(key K) ([]*V, error) {
	keyString := keyToString(key)
	var keyValues []record
	err := s.db.
		Where("key LIKE ?", keyString+"%").
		Find(&keyValues).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get prefix: %w", err)
	}

	values := []*V{}
	for _, keyValue := range keyValues {
		value := new(V)
		err := json.Unmarshal(keyValue.Value, value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal value: %w", err)
		}
		values = append(values, value)
	}

	return values, nil
}
