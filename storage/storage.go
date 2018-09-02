package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("Not Found")
	ErrEmpty    = errors.New("Empty")
	ErrExists   = errors.New("Already Exists")
)

type Storage interface {
	Store(string, []byte) error
	Get(string) ([]byte, error)
	List() ([]string, error)
}
