package net

import (
	"context"
)

type Record interface {
	GetKey() string
	GetValue() string
	GetLabels() map[string]string
}

type DHT interface {
	Put(ctx context.Context, key, value string, labels map[string]string) error
	Get(ctx context.Context, key string) (chan Record, error)
	Filter(ctx context.Context, key string, filter map[string]string) (chan Record, error)
}
