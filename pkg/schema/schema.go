package schema

type (
	Property struct {
		Name       string      `json:"name:s,omitempty"`
		Type       string      `json:"type:s,omitempty"`
		Hint       string      `json:"hint:s,omitempty"`
		IsRepeated bool        `json:"isRepeated:b,omitempty"`
		IsOptional bool        `json:"isOptional:b,omitempty"`
		Properties []*Property `json:"properties:ao,omitempty"`
	}
	Object struct {
		Properties []*Property `json:"properties:ao,omitempty"`
	}
)
