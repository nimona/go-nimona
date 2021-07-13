package keyvalue

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xujiajun/nutsdb"
)

// TestNutDBStore tests the nutdb implementation of the KeyValue interface.
func TestNutDBStore(t *testing.T) {
	opt := nutsdb.DefaultOptions
	opt.Dir = t.TempDir()
	db, err := nutsdb.Open(opt)
	require.NoError(t, err)
	s := NewNutsDBStore(db, "test")
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
