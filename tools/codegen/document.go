package main

type Document struct {
	PackageAlias string
	Package      string
	Imports      map[string]string
	Objects      []*Object
	Streams      []*Stream
}

type Stream struct {
	Name    string
	Objects []*Object
}

type Object struct {
	Name      string
	IsRoot    bool
	IsEvent   bool
	IsSigned  bool
	IsCommand bool
	Members   []*Member
}

// rootHashes repeated string nimona.io/tilde.Digest
// ^        ^        ^      ^
// |        |        |      |
// |        |        |      - GoType
// |        |        - SimpleType
// |        - IsRepeated
// - Tag
//
type Member struct {
	Name       string
	GoFullType string
	SimpleType string
	Tag        string
	Hint       string

	IsObject    bool
	IsPrimitive bool

	IsOptional bool
	IsRepeated bool
}
