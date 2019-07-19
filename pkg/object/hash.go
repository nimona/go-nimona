package object

// Hash returns the ObjectHash of the object
func Hash(o Object) []byte {
	b, err := ObjectHash(o)
	if err != nil {
		panic(err)
	}

	return b
}
