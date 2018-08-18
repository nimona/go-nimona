package storage

import (
	"errors"

	"github.com/nimona/go-nimona/blocks"
)

var (
	ErrNotFound = errors.New("Not Found")
	ErrEmpty    = errors.New("Empty")
	ErrExists   = errors.New("Already Exists")
)

type Storage interface {
	Store(string, *blocks.Block) error
	Get(string) (*blocks.Block, error)
	List() ([]string, error)
}
