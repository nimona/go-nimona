package net

import "errors"

var (
	// ErrNotFound = errors.New("Not Found")
	ErrEmpty  = errors.New("Empty")
	ErrExists = errors.New("Already Exists")
)

type Storage interface {
	Store(string, *Block) error
	Get(string) (*Block, error)
	List() ([]string, error)
}
