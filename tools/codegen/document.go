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
	Links     []*Link
}

type Link struct {
	Type       string
	Direction  string
	IsOptional bool
}

type Member struct {
	Name       string
	Type       string
	SimpleType string
	Tag        string
	Hint       string

	IsObject    bool
	IsPrimitive bool

	IsOptional bool
	IsRepeated bool
}
