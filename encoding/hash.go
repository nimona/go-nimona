package encoding

// // Hash is a block (container) for a hash
// type Hash struct {
// 	Alg    string `json:"alg"`
// 	Digest []byte `json:"dig"`
// }

// // // Bytes returns the marshaled block
// // func (h *Hash) Bytes() []byte {
// // 	b, err := encoding.Marshal(h)
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	return b
// // }

// // // Base58 returns the base58 representation of the marshaled block
// // func (h *Hash) Base58() string {
// // 	b, err := encoding.Marshal(h)
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	return base58.Encode(b)
// // }

// // NewSha3 creates a new block (container) for a sha3 hash given a payload
// func NewSha3(p []byte) *Hash {
// 	d := sha3.Sum256(p)
// 	return &Hash{
// 		Alg:    "SHA3",
// 		Digest: d[:],
// 	}
// }

func Hash(o *Object) []byte {
	b, err := ObjectHash(o)
	if err != nil {
		panic(err)
	}

	return b
}
