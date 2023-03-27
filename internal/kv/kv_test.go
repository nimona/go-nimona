package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestKVStore(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	type TestValue struct {
		Foo string `json:"foo"`
	}

	// Create a new SQLStore instance
	store, err := NewSQLStore[TestValue](db)
	assert.NoError(t, err)

	// Set a new key-value pair with a struct as value
	testValue := &TestValue{Foo: "bar"}
	err = store.Set("structValue", testValue)
	assert.NoError(t, err)

	// Get the value back and check that it matches the original value
	returnedValue, err := store.Get("structValue")
	assert.NoError(t, err)
	assert.Equal(t, testValue, returnedValue)
}
