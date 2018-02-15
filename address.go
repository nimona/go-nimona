package fabric

import (
	"strings"
)

// NewAddress creates a new ADdress from a string address
func NewAddress(a string) *Address {
	return &Address{
		stack: strings.Split(a, "/"),
		index: 0,
	}
}

// Address allows traversing and validating an address
type Address struct {
	stack []string
	index int
}

// String version of the full address
func (a *Address) String() string {
	return strings.Join(a.stack, "/")
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

// Current returns the current item
func (a *Address) Current() string {
	// TODO index could be out of range
	return a.stack[a.index]
}

// CurrentProtocol return the current item's protocol
func (a *Address) CurrentProtocol() string {
	// TODO index could be out of range
	return strings.Split(a.stack[a.index], ":")[0]
}

// CurrentParams return the current item's params
func (a *Address) CurrentParams() string {
	// TODO index could be out of range
	pr := strings.Split(a.stack[a.index], ":")
	if len(pr) == 1 {
		return ""
	}

	return strings.Join(pr[1:], ":")
}

// Remaining returns the remaining stack items
func (a *Address) Remaining() []string {
	// TODO index could be out of range
	return a.stack[a.index:]
}

// RemainingProtocols returns the remaining protocols in the stack
func (a *Address) RemainingProtocols() []string {
	// TODO index could be out of range
	prs := []string{}
	for _, pr := range a.stack[a.index:] {
		pas := strings.Split(pr, ":")
		prs = append(prs, pas[0])
	}
	return prs
}

// RemainingString returns the remaining stack items as a string
func (a *Address) RemainingString() string {
	// TODO index could be out of range
	return strings.Join(a.stack[a.index:], "/")
}
