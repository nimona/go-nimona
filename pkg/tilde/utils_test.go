package tilde

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolArray(t *testing.T) {
	tests := []struct {
		have []bool
		want BoolArray
	}{{
		have: []bool{false, true},
		want: BoolArray{false, true},
	}}
	for _, tt := range tests {
		t.Run("BoolArray", func(t *testing.T) {
			got := ToBoolArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDataArray(t *testing.T) {
	tests := []struct {
		have [][]byte
		want DataArray
	}{{
		have: [][]byte{{1, 2}, {3, 4}},
		want: DataArray{[]byte{1, 2}, []byte{3, 4}},
	}}
	for _, tt := range tests {
		t.Run("DataArray", func(t *testing.T) {
			got := ToDataArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFloatArray(t *testing.T) {
	tests := []struct {
		have []float64
		want FloatArray
	}{{
		have: []float64{1.0, 1.1},
		want: FloatArray{1.0, 1.1},
	}}
	for _, tt := range tests {
		t.Run("FloatArray", func(t *testing.T) {
			got := ToFloatArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIntArray(t *testing.T) {
	tests := []struct {
		have []int64
		want IntArray
	}{{
		have: []int64{1, 2},
		want: IntArray{1, 2},
	}}
	for _, tt := range tests {
		t.Run("IntArray", func(t *testing.T) {
			got := ToIntArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringArray(t *testing.T) {
	tests := []struct {
		have []string
		want StringArray
	}{{
		have: []string{"foo", "bar"},
		want: StringArray{"foo", "bar"},
	}}
	for _, tt := range tests {
		t.Run("StringArray", func(t *testing.T) {
			got := ToStringArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUintArray(t *testing.T) {
	tests := []struct {
		have []uint64
		want UintArray
	}{{
		have: []uint64{1, 2},
		want: UintArray{1, 2},
	}}
	for _, tt := range tests {
		t.Run("UintArray", func(t *testing.T) {
			got := ToUintArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDigestArray(t *testing.T) {
	tests := []struct {
		have [][]byte
		want DigestArray
	}{{
		have: [][]byte{[]byte("foo"), []byte("bar")},
		want: DigestArray{Digest("foo"), Digest("bar")},
	}}
	for _, tt := range tests {
		t.Run("DigestArray", func(t *testing.T) {
			got := ToDigestArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromBoolArray(t *testing.T) {
	tests := []struct {
		have BoolArray
		want []bool
	}{{
		have: BoolArray{false, true},
		want: []bool{false, true},
	}}
	for _, tt := range tests {
		t.Run("BoolArray", func(t *testing.T) {
			got := FromBoolArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromDataArray(t *testing.T) {
	tests := []struct {
		have DataArray
		want [][]byte
	}{{
		have: DataArray{[]byte{1, 2}, []byte{3, 4}},
		want: [][]byte{{1, 2}, {3, 4}},
	}}
	for _, tt := range tests {
		t.Run("DataArray", func(t *testing.T) {
			got := FromDataArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromFloatArray(t *testing.T) {
	tests := []struct {
		have FloatArray
		want []float64
	}{{
		have: FloatArray{1.0, 1.1},
		want: []float64{1.0, 1.1},
	}}
	for _, tt := range tests {
		t.Run("FloatArray", func(t *testing.T) {
			got := FromFloatArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromIntArray(t *testing.T) {
	tests := []struct {
		have IntArray
		want []int64
	}{{
		have: IntArray{1, 2},
		want: []int64{1, 2},
	}}
	for _, tt := range tests {
		t.Run("IntArray", func(t *testing.T) {
			got := FromIntArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromStringArray(t *testing.T) {
	tests := []struct {
		have StringArray
		want []string
	}{{
		have: StringArray{"foo", "bar"},
		want: []string{"foo", "bar"},
	}}
	for _, tt := range tests {
		t.Run("StringArray", func(t *testing.T) {
			got := FromStringArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromUintArray(t *testing.T) {
	tests := []struct {
		have UintArray
		want []uint64
	}{{
		have: UintArray{1, 2},
		want: []uint64{1, 2},
	}}
	for _, tt := range tests {
		t.Run("UintArray", func(t *testing.T) {
			got := FromUintArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromDigestArray(t *testing.T) {
	tests := []struct {
		have DigestArray
		want [][]byte
	}{{
		have: DigestArray{Digest("foo"), Digest("bar")},
		want: [][]byte{[]byte("foo"), []byte("bar")},
	}}
	for _, tt := range tests {
		t.Run("DigestArray", func(t *testing.T) {
			got := FromDigestArray(tt.have)
			assert.Equal(t, tt.want, got)
		})
	}
}
