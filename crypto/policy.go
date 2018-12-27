package crypto

// Policy for Block
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Effect      string   `json:"effect,omitempty"`
}

// func ID(v interface{}) string {
// 	d, err := encoding.Marshal(v)
// 	if err != nil {
// 		panic(err)
// 	}

// 	h := NewSha3(d)
// 	b, err := encoding.Marshal(h)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return string(base58.Encode(b))
// }
