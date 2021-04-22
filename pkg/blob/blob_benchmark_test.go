package blob_test

import (
	"testing"

	"github.com/docker/go-units"

	"nimona.io/internal/iotest"
	"nimona.io/pkg/blob"
)

func BenchmarkNewBlob1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fr := iotest.ZeroReader(1 * units.MB)
		_, _, err := blob.NewBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkNewBlob100(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fr := iotest.ZeroReader(100 * units.MB)
		_, _, err := blob.NewBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkNewBlob1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fr := iotest.ZeroReader(1000 * units.MB)
		_, _, err := blob.NewBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}
