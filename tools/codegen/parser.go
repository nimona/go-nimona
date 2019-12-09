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
	ErrNotDomain = errors.New("not a domain")
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
		DOMAIN,
		OBJECT:
		_, text, err := p.expect(TEXT)
		if err != nil {
			return token, text, fmt.Errorf(
				"found %q, didn't find expected TEXT after", value,
			)
		}
		return token, text, nil
	}

	return token, value, nil
}

func (p *Parser) parseField() (*Member, error) {
	token, value := p.scanIgnoreWhiteSpace()
	member := &Member{}
	if token != TEXT {
		p.unscan()
		return nil, ErrNotField
	}
	member.Tag = value
	member.Name = memberName(value)
	token, value = p.scanIgnoreWhiteSpace()
	if token == REPEATED {
		member.IsRepeated = true
		token, value = p.scanIgnoreWhiteSpace()
	}
	if token != TEXT {
		return nil, fmt.Errorf("found %q, expected member.type", value)
	}
	switch value {
	case "string":
		member.Type += "string"
		member.Hint = "s"
	case "int":
		member.Type += "int64"
		member.Hint = "i"
	case "uint":
		member.Type += "uint64"
		member.Hint = "u"
	case "float":
		member.Type += "float64"
		member.Hint = "f"
	case "bool":
		member.Type += "bool"
		member.Hint = "b"
	case "data":
		member.Type += "[]byte"
		member.Hint = "d"
	default:
		member.Type += value
		member.Hint = "o"
		member.IsObject = true
	}
	fmt.Println("\tFound attribute", member.Name, "of type", member.Type)
	return member, nil
}

func (p *Parser) parseEvent() (*Object, error) {
	// create object
	object := &Object{}

	// expect SIGNED, or OBJECT
	token, value, err := p.expect(SIGNED, OBJECT)
	if err != nil {
		return nil, err
	}

	if token == SIGNED {
		object.IsSigned = true
		token, value, err = p.expect(OBJECT)
		if err != nil {
			return nil, err
		}
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
		member, err := p.parseField()
		switch {
		case err != nil && err == ErrNotField:
			continue
		case err != nil:
			return nil, err
		case err == nil:
			object.Members = append(object.Members, member)
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
		if token != TEXT {
			return nil, fmt.Errorf("found %q, expected TEXT for import", token)
		}
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

		// parse object
		object, err := p.parseEvent()
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
