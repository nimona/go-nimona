package encoding

//go:generate go run nimona.io/go/generators/objectify -schema /policy -type Policy -out policy_generated.go

// Policy for object
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Effect      string   `json:"effect,omitempty"`
}
