package kv

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type record struct {
	Key   string `gorm:"primaryKey"`
	Value []byte
}

func NewSQLStore[K, V any](db *gorm.DB, table string) (Store[K, V], error) {
	s := &SQLStore[K, V]{
		db:    db,
		table: table,
	}

	err := s.getDB().AutoMigrate(&record{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate record: %w", err)
	}

	return s, nil
}

type SQLStore[K, V any] struct {
	db    *gorm.DB
	table string
}

func (s *SQLStore[K, V]) getDB() *gorm.DB {
	return s.db.Table(s.table)
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

	err = s.getDB().
		Clauses(
			clause.OnConflict{
				DoNothing: true,
			},
		).
		Create(keyValue).
		Error
	if err != nil {
		return fmt.Errorf("failed to set: %w", err)
	}

	return nil
}

func (s *SQLStore[K, V]) Get(key K) (*V, error) {
	keyString := keyToString(key)
	var keyValues []record
	err := s.getDB().
		Where("key = ?", keyString).
		Find(&keyValues).
		Error
	if err != nil {
		return nil, err
	}

	if len(keyValues) == 0 {
		return nil, fmt.Errorf("value not found")
	}

	keyValue := keyValues[0]

	value := new(V)
	err = json.Unmarshal(keyValue.Value, value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *SQLStore[K, V]) GetPrefix(key K) ([]*V, error) {
	keyString := keyToString(key)
	var keyValues []record
	err := s.getDB().
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
