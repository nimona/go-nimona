package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
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
		EVENT:
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

func (p *Parser) parseEvent() (*Event, error) {
	// expect "event" and value
	_, eventName, err := p.expect(EVENT)
	if err != nil {
		return nil, err
	}

	// create event
	event := &Event{
		Name: eventName,
	}

	fmt.Println("\tFound event", event.Name)

	// parse attributes
	for {
		token, value := p.scanIgnoreWhiteSpace()
		if token == EBRACE {
			token, _ = p.scanIgnoreWhiteSpace()
			p.unscan()
			break
		}
		if token == TEXT {
			member := &Member{
				Name: memberName(value),
				Tag:  value,
			}
			tokNext, memberType := p.scanIgnoreWhiteSpace()
			if tokNext != TEXT {
				return nil, fmt.Errorf("found %q, expected TEXT", memberType)
			}
			switch memberType {
			case "string":
				member.Type = "string"
				member.Tag += ":s"
			case "int":
				member.Type = "int64"
				member.Tag += ":i"
			case "uint":
				member.Type = "uint64"
				member.Tag += ":u"
			case "float":
				member.Type = "float64"
				member.Tag += ":f"
			case "bool":
				member.Type = "bool"
				member.Tag += ":b"
			case "data":
				member.Type = "[]byte"
				member.Tag += ":d"
			default:
				member.Type = memberType
				member.Tag += ":o"
			}
			event.Members = append(event.Members, member)
			fmt.Println("\t\tFound attribute", value, "of type", memberType)
		}
	}
	return event, nil
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
		if token != TEXT {
			return nil, fmt.Errorf("found %q, expected TEXT for import alias", token)
		}
		fmt.Println("Found import", importPkg, "as", importAlias)
		doc.Imports[importAlias] = importPkg
	}

	// gather domains
	for {
		domain := &Domain{}

		// expect "domain" and value
		token, domainName, err := p.expect(DOMAIN)
		if err != nil {
			if token == EOF {
				break
			}
			return nil, err
		}

		domain.Name = domainName

		fmt.Println("Found domain", domain.Name)

		// expect "{"
		if _, value, err := p.expect(OBRACE); err != nil {
			return nil, fmt.Errorf("found %q, expected EVENT", value)
		}

		for {
			// if "}", break
			if token, _ := p.scanIgnoreWhiteSpace(); token == EBRACE {
				break
			}

			p.unscan()

			// parse event
			event, err := p.parseEvent()
			if err != nil {
				return nil, err
			}
			domain.Events = append(domain.Events, event)
		}

		doc.Domains = append(doc.Domains, domain)
	}

	return doc, nil
}

var memberNameRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func memberName(str string) string {
	// remove special chars
	str = memberNameRegex.ReplaceAllString(str, "")
	// uppercase first letter
	for i, v := range str {
		str = string(unicode.ToUpper(v)) + str[i+1:]
		break
	}
	return str
}

func eventName(str string) string {
	// uppercase first letter
	for i, v := range str {
		str = string(unicode.ToUpper(v)) + str[i+1:]
		break
	}
	return str
}
