package fabric

import (
	"strings"
)

// NewAddress creates a new ADdress from a string address
func NewAddress(a string) *Address {
	return &Address{
		original: a,
		stack:    strings.Split(a, "/"),
		index:    0,
	}
}

// Address allows traversing and validating an address
type Address struct {
	original string
	stack    []string
	index    int
}

// Reset the stack index
func (a *Address) Reset() {
	a.index = 0
}

// Pop the first stack item and return it
func (a *Address) Pop() string {
	ci := a.index
	if ci >= len(a.stack) {
		return ""
	}

	a.index++
	return a.stack[ci]
}

// Remaining returns the remaining stack items
func (a *Address) Remaining() []string {
	return a.stack[a.index:]
}

// RemainingString returns the remaining stack items as a string
func (a *Address) RemainingString() string {
	return strings.Join(a.stack[a.index:], "/")
}
