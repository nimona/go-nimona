package object

//go:generate $GOBIN/objectify -schema /policy -type Policy -in object_policy.go -out object_policy_generated.go

// Policy for object
type Policy struct {
	Description string   `json:"description:s,omitempty"`
	Subjects    []string `json:"subjects:as,omitempty"`
	Actions     []string `json:"actions:as,omitempty"`
	Effect      string   `json:"effect:s,omitempty"`
}
