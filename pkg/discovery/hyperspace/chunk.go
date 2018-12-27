package hyperspace

func chunk(b []byte, n int) [][]byte {
	r := [][]byte{}
	for i := range b {
		if i > 0 && i%n == 0 {
			r = append(r, b[len(r)*n:i])
		}
	}
	if len(b) > len(r)*n {
		r = append(r, b[len(r)*n:])
	}
	return r
}
