package fabric

import (
	"strings"
)

func NewAddress(a string) *Address {
	return &Address{
		original: a,
		stack:    strings.Split(a, "/"),
		index:    0,
	}
}

type Address struct {
	original string
	stack    []string
	index    int
}

func (a *Address) Reset() {
	a.index = 0
}

func (a *Address) Pop() string {
	ci := a.index
	if ci >= len(a.stack) {
		return ""
	}

	a.index++
	return a.stack[ci]
}

func (a *Address) Remaining() []string {
	return a.stack[a.index:]
}

func (a *Address) RemainingString() string {
	return strings.Join(a.stack[a.index:], "/")
}
