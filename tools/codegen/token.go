package main

import "regexp"

type Token string

const (
	UNKNOWN    Token = "UNKNOWN"
	DOMAIN     Token = "DOMAIN"
	EBRACE     Token = "EBRACE"
	EOF        Token = "EOF"
	EVENT      Token = "EVENT"
	COMMAND    Token = "COMMAND"
	STRUCT     Token = "STRUCT"
	REPEATED   Token = "REPEATED"
	SIGNED     Token = "SIGNED"
	ABSTRACT   Token = "ABSTRACT"
	EXTENDS    Token = "EXTENDS"
	PRIMARY    Token = "PRIMARY"
	IMPORT     Token = "IMPORT"
	OBRACE     Token = "OBRACE"
	PACKAGE    Token = "PACKAGE"
	TEXT       Token = "TEXT"
	WHITESPACE Token = "WS"
)

var (
	eof       = rune(0)
	wsRegex   = regexp.MustCompile("^[\\n\\t\\s]+")
	textRegex = regexp.MustCompile("^[a-zA-Z0-9\\._@\\/]+$")
	keywords  = map[string]Token{
		"package":  PACKAGE,
		"import":   IMPORT,
		"domain":   DOMAIN,
		"event":    EVENT,
		"command":  COMMAND,
		"struct":   STRUCT,
		"repeated": REPEATED,
		"primary":  PRIMARY,
		"signed":   SIGNED,
		"abstract": ABSTRACT,
		"extends":  EXTENDS,
	}
)

func isWhitespace(s rune) bool {
	return wsRegex.MatchString(string(s))
}

func isText(s rune) bool {
	return textRegex.MatchString(string(s))
}

func isOpenBrace(ch rune) bool {
	return ch == '{'
}

func isEndBrace(ch rune) bool {
	return ch == '}'
}

func isEOF(ch rune) bool {
	return ch == eof
}
