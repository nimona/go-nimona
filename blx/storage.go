package blx

import "errors"

type Storage interface {
	Store(string, *Block) error
	Get(string) (*Block, error)
}

var ErrNotFound = errors.New("Not Found")
