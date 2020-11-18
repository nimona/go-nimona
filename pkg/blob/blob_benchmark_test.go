// +build benchmark flaky

package blob_test

import (
	"os"
	"testing"

	"nimona.io/pkg/blob"
)

func BenchmarkToBlob1(b *testing.B) {
	tempFile := newTestFile(1)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkToBlob100(b *testing.B) {
	tempFile := newTestFile(100)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkToBlob1000(b *testing.B) {
	tempFile := newTestFile(1000)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}
