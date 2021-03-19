package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"nimona.io/pkg/errors"
)

var (
	ErrNotField  = errors.Error("not a field")
	ErrNotStream = errors.Error("not a stream")
	ErrNotStruct = errors.Error("not a object")
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
		EVENT,
		OBJECT:
		_, text := p.scanIgnoreWhiteSpace()
		return token, text, nil
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
	member.Tag = value
	member.Name = memberName(value)
	token, value = p.scanIgnoreWhiteSpace()
	if token == REPEATED {
		member.IsRepeated = true
		token, value = p.scanIgnoreWhiteSpace()
	}
	switch value {
	// case "relationship":
	// 	member.GoFullType = "object.CID"
	// 	member.SimpleType = "relationship"
	// 	member.Hint = "r"
	case "string":
		member.GoFullType = "string"
		member.SimpleType = "string"
		member.Hint = "s"
		member.IsPrimitive = true
	case "int":
		member.GoFullType = "int64"
		member.SimpleType = "int"
		member.Hint = "i"
		member.IsPrimitive = true
	case "uint":
		member.GoFullType = "uint64"
		member.SimpleType = "uint"
		member.Hint = "u"
		member.IsPrimitive = true
	case "float":
		member.GoFullType = "float64"
		member.SimpleType = "float"
		member.Hint = "f"
		member.IsPrimitive = true
	case "bool":
		member.GoFullType = "bool"
		member.SimpleType = "bool"
		member.Hint = "b"
		member.IsPrimitive = true
	case "data":
		member.GoFullType = "[]byte"
		member.SimpleType = "data"
		member.Hint = "d"
		member.IsPrimitive = true
	case "map":
		member.GoFullType = "map[string]interface{}"
		member.SimpleType = "map"
		member.Hint = "m"
	default:
		member.GoFullType = value
		member.SimpleType = value
		member.Hint = "o"
		member.IsObject = true
	}
	// expect TEXT for optional Member.GoFullType
	token, value = p.scanIgnoreWhiteSpace()
	if strings.HasPrefix(value, "type=") {
		member.GoFullType = strings.TrimPrefix(value, "type=")
		member.IsObject = true
		member.IsPrimitive = false
	} else {
		p.unscan()
	}
	fmt.Println("\tFound attribute", member.Name, "with hint", member.SimpleType, "of type", member.GoFullType)
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
	// expect SIGNED, ROOT, EVENT, or OBJECT
	token, value, err := p.expect(SIGNED, ROOT, OBJECT, EVENT)
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
	case EVENT:
		object.IsEvent = true
	}

	object.Name = value

	if object.IsEvent {
		fmt.Println("\tFound event", object.Name)
	} else {
		fmt.Println("\tFound object", object.Name)
	}
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
