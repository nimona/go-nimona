package crypto

// Policy for Object
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Effect      string   `json:"effect,omitempty"`
}
