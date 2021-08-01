package keyvalue

import (
	"testing"
)

// TestInMemoryStore tests the in-memory implementation of the KeyValue
// interface.
func TestInMemoryStore(t *testing.T) {
	s := NewInMemoryStore()
	defer s.Close()
	if err := s.Set("foo", []byte("bar")); err != nil {
		t.Fatal(err)
	}
	if value, err := s.Get("foo"); err != nil {
		t.Fatal(err)
	} else if string(value) != "bar" {
		t.Fatalf("expected %q, got %q", "bar", string(value))
	}
	if err := s.Delete("foo"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("foo"); err != ErrNotFound {
		t.Fatalf("expected %q, got %q", ErrNotFound, err)
	}
}
