package main

type Document struct {
	PackageAlias string
	Package      string
	Imports      map[string]string
	Objects      []*Object
}

type Object struct {
	Name      string
	IsSigned  bool
	IsCommand bool
	Members   []*Member
}

type Member struct {
	Name string
	Type string
	Tag  string
	Hint string

	IsPrimitive bool
	IsObject    bool
	Required    bool
	IsRepeated  bool
}
