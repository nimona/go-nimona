package object

import "strconv"

func traverse(k string, v Value, f func(string, Value) bool) bool {
	cont := f(k, v)
	if !cont {
		return false
	}
	if k != "" {
		k += pathSeperator
	}
	switch cv := v.(type) {
	case Map:
		cont = cv.iterateSorted(func(ik string, iv Value) bool {
			cont = traverse(k+ik, iv, f)
			return cont
		})
	case List:
		i := 0
		cv.Iterate(func(ii int, iv Value) bool {
			cont = traverse(k+strconv.Itoa(ii), iv, f)
			i++
			return cont
		})
	}
	return cont
}

func Traverse(v Value, f func(string, Value) bool) {
	traverse("", v, f)
}
