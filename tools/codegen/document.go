package main

type Document struct {
	PackageAlias string
	Package      string
	Imports      map[string]string
	Domains      []*Domain
	Structs      []*Struct
}

type Domain struct {
	IsAbstract bool
	Extends    string
	Name       string
	Events     []*Event
}

type Event struct {
	Name     string
	IsSigned bool
	Members  []*Member
}

type Struct struct {
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
