package object

import "github.com/mitchellh/copystructure"

func Copy(s *Object) *Object {
	r, err := copystructure.Copy(s)
	if err != nil {
		panic(err)
	}
	return r.(*Object)
}

// conv helper methods

func ToBoolArray(s []bool) BoolArray {
	r := make(BoolArray, len(s))
	for i, v := range s {
		r[i] = Bool(v)
	}
	return r
}

func ToDataArray(s [][]byte) DataArray {
	r := make(DataArray, len(s))
	for i, v := range s {
		r[i] = Data(v)
	}
	return r
}

func ToFloatArray(s []float64) FloatArray {
	r := make(FloatArray, len(s))
	for i, v := range s {
		r[i] = Float(v)
	}
	return r
}

func ToIntArray(s []int64) IntArray {
	r := make(IntArray, len(s))
	for i, v := range s {
		r[i] = Int(v)
	}
	return r
}

func ToStringArray(s []string) StringArray {
	r := make(StringArray, len(s))
	for i, v := range s {
		r[i] = String(v)
	}
	return r
}

func ToUintArray(s []uint64) UintArray {
	r := make(UintArray, len(s))
	for i, v := range s {
		r[i] = Uint(v)
	}
	return r
}

func ToCIDArray(s []string) CIDArray {
	r := make(CIDArray, len(s))
	for i, v := range s {
		r[i] = CID(v)
	}
	return r
}

func FromBoolArray(s BoolArray) []bool {
	r := make([]bool, len(s))
	for i, v := range s {
		r[i] = bool(v)
	}
	return r
}

func FromDataArray(s DataArray) [][]byte {
	r := make([][]byte, len(s))
	for i, v := range s {
		r[i] = []byte(v)
	}
	return r
}

func FromFloatArray(s FloatArray) []float64 {
	r := make([]float64, len(s))
	for i, v := range s {
		r[i] = float64(v)
	}
	return r
}

func FromIntArray(s IntArray) []int64 {
	r := make([]int64, len(s))
	for i, v := range s {
		r[i] = int64(v)
	}
	return r
}

func FromStringArray(s StringArray) []string {
	r := make([]string, len(s))
	for i, v := range s {
		r[i] = string(v)
	}
	return r
}

func FromUintArray(s UintArray) []uint64 {
	r := make([]uint64, len(s))
	for i, v := range s {
		r[i] = uint64(v)
	}
	return r
}

func FromCIDArray(s CIDArray) []string {
	r := make([]string, len(s))
	for i, v := range s {
		r[i] = string(v)
	}
	return r
}
