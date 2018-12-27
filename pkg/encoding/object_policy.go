package encoding

//go:generate go run nimona.io/tools/objectify -schema /policy -type Policy -in object_policy.go -out policy_generated.go

// Policy for object
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Effect      string   `json:"effect,omitempty"`
}
