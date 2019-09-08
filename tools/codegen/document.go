package main

type Document struct {
	PackageAlias string
	Package      string
	Imports      map[string]string
	Domains      []*Domain
}

type Domain struct {
	Name   string
	Events []*Event
}

type Event struct {
	Name    string
	Members []*Member
}

type Member struct {
	Name string
	Type string
	Tag  string
}
