package main

type Document struct {
	PackageAlias string
	Package      string
	Imports      map[string]string
	Domains      []*Domain
	Objects      []*Object
}

type Domain struct {
	IsAbstract bool
	Extends    string
	Name       string
	Events     []*Event
	Objects    []*Object
}

type Event struct {
	Name      string
	IsSigned  bool
	IsCommand bool
	Members   []*Member
}

type Object struct {
	Name    string
	Members []*Member
}

type Member struct {
	Name string
	Type string
	Tag  string

	Required bool
	Repeated bool
}
