package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrNotField  = errors.New("not a field")
	ErrNotStream = errors.New("not a stream")
	ErrNotStruct = errors.New("not a object")
)

type Parser struct {
	s   *Scanner
	buf struct {
		token Token
		value string
		n     int
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s: NewScanner(r),
	}
}

func (p *Parser) scan() (Token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.token, p.buf.value
	}

	p.buf.token, p.buf.value = p.s.Scan()
	return p.buf.token, p.buf.value
}

func (p *Parser) unscan() {
	p.buf.n = 1
}

func (p *Parser) scanIgnoreWhiteSpace() (Token, string) {
	token, value := p.scan()
	if token == WHITESPACE {
		token, value = p.scan()
	}
	return token, value
}

func (p *Parser) expect(ets ...Token) (Token, string, error) {
	token, value := p.scanIgnoreWhiteSpace()
	found := false
	for _, et := range ets {
		if et == token {
			found = true
			break
		}
	}
	if !found {
		etss := []string{}
		for _, et := range ets {
			etss = append(etss, string(et))
		}
		return token, value, fmt.Errorf("found %q, expected %v", value, etss)
	}

	switch token {
	// if given token always expects TEXT afterwards, let's find it
	case PACKAGE,
		STREAM,
		OBJECT:
		_, text := p.scanIgnoreWhiteSpace()
		return TEXT, text, nil
	}

	return token, value, nil
}

func (p *Parser) parseField() (interface{}, error) {
	token, value := p.scanIgnoreWhiteSpace()
	member := &Member{}
	if token == OPTIONAL {
		member.IsOptional = true
		token, value = p.scanIgnoreWhiteSpace()
	}
	if token == LINK {
		link := &Link{
			IsOptional: member.IsOptional,
		}
		_, value = p.scanIgnoreWhiteSpace()
		switch strings.ToLower(value) {
		case "in":
			link.Direction = "in"
		case "out":
			link.Direction = "out"
		default:
			return nil, fmt.Errorf("found %q, expected in or out for link.direction", value)
		}
		_, value = p.scanIgnoreWhiteSpace()
		link.Type = value
		return link, nil
	}
	member.Tag = value
	member.Name = memberName(value)
	token, value = p.scanIgnoreWhiteSpace()
	if token == REPEATED {
		member.IsRepeated = true
		token, value = p.scanIgnoreWhiteSpace()
	}
	switch value {
	case "string":
		member.Type = "string"
		member.SimpleType = "string"
		member.Hint = "s"
	case "int":
		member.Type = "int64"
		member.SimpleType = "int"
		member.Hint = "i"
	case "uint":
		member.Type = "uint64"
		member.SimpleType = "uint"
		member.Hint = "u"
	case "float":
		member.Type = "float64"
		member.SimpleType = "float"
		member.Hint = "f"
	case "bool":
		member.Type = "bool"
		member.SimpleType = "bool"
		member.Hint = "b"
	case "data":
		member.Type = "[]byte"
		member.SimpleType = "data"
		member.Hint = "d"
	default:
		member.Type = value
		member.SimpleType = value
		member.Hint = "o"
		member.IsObject = true
	}
	fmt.Println("\tFound attribute", member.Name, "of type", member.Type)
	return member, nil
}

func (p *Parser) parseStream() (*Stream, error) {
	// create stream
	stream := &Stream{}

	// expect STREAM
	_, value, err := p.expect(STREAM)
	if err != nil {
		p.unscan()
		return nil, ErrNotStream
	}

	stream.Name = value

	fmt.Println("Found stream", stream.Name)

	_, value, err = p.expect(OBRACE)
	if err != nil {
		return nil, fmt.Errorf("found %q, expected OBRACE", value)
	}

	for {
		// if "}", break
		token, _ := p.scanIgnoreWhiteSpace()
		if token == EBRACE {
			break
		}

		p.unscan()

		// parse object
		object, err := p.parseObject()
		if err != nil {
			return nil, err
		}
		stream.Objects = append(stream.Objects, object)
	}

	return stream, nil
}

func (p *Parser) parseObject() (*Object, error) {
	// create object
	object := &Object{}

start:
	// expect SIGNED, ROOT, or OBJECT
	token, value, err := p.expect(SIGNED, ROOT, OBJECT)
	if err != nil {
		return nil, err
	}

	switch token {
	case SIGNED:
		object.IsSigned = true
		goto start
	case ROOT:
		object.IsRoot = true
		goto start
	}

	object.Name = value

	fmt.Println("\tFound object", object.Name)

	if _, _, err := p.expect(OBRACE); err != nil {
		return nil, err
	}

	// parse attributes
	for {
		if token, _ := p.scanIgnoreWhiteSpace(); token == EBRACE {
			break
		}
		p.unscan()
		res, err := p.parseField()
		switch {
		case err != nil && err == ErrNotField:
			continue
		case err != nil:
			return nil, err
		case err == nil:
			switch v := res.(type) {
			case *Link:
				object.Links = append(object.Links, v)
			case *Member:
				object.Members = append(object.Members, v)
			}
		}
	}

	return object, nil
}

func (p *Parser) Parse() (*Document, error) {
	doc := &Document{
		Imports: map[string]string{},
	}

	// excpect "package" and value
	_, pkg, err := p.expect(PACKAGE)
	if err != nil {
		return nil, err
	}

	doc.Package = pkg
	ps := strings.Split(pkg, "/")
	doc.PackageAlias = ps[len(ps)-1]

	fmt.Println("Found package name", doc.Package)

	// gather imports
	for {
		token, _ := p.scanIgnoreWhiteSpace()
		if token != IMPORT {
			p.unscan()
			break
		}
		token, importPkg := p.scanIgnoreWhiteSpace()
		token, importAlias := p.scanIgnoreWhiteSpace()
		// doesn't really matter what the token is here, as long as it's not eof
		if token == EOF {
			return nil, fmt.Errorf("found %q, expected TEXT for import alias", token)
		}
		fmt.Println("Found import", importPkg, "as", importAlias)
		doc.Imports[importAlias] = importPkg
	}

	for {
		// if "}", break
		token, _ := p.scanIgnoreWhiteSpace()
		if token == EOF {
			break
		}

		p.unscan()

		// expect stream
		stream, err := p.parseStream()
		if err != nil && err != ErrNotStream {
			return nil, err
		} else if err == nil {
			doc.Streams = append(doc.Streams, stream)
			continue
		}

		p.unscan()

		// parse object
		object, err := p.parseObject()
		if err != nil {
			return nil, err
		}
		doc.Objects = append(doc.Objects, object)
	}

	return doc, nil
}

var memberNameRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func memberName(str string) string {
	// remove special chars
	str = memberNameRegex.ReplaceAllString(str, "")
	return ucFirst(str)
}

func eventName(str string) string {
	return ucFirst(str)
}

func ucFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
		break
	}
	return ""
}
