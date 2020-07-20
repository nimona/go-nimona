package main

import "regexp"

type Token string

const (
	UNKNOWN    Token = "UNKNOWN"
	STREAM     Token = "STREAM"
	EBRACE     Token = "EBRACE"
	EOF        Token = "EOF"
	OBJECT     Token = "OBJECT"
	EVENT      Token = "EVENT"
	MAP        Token = "MAP"
	LINK       Token = "LINK"
	ROOT       Token = "ROOT"
	REPEATED   Token = "REPEATED"
	OPTIONAL   Token = "OPTIONAL"
	SIGNED     Token = "SIGNED"
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
		"stream":   STREAM,
		"object":   OBJECT,
		"event":    EVENT,
		"map":      MAP,
		"@link":    LINK,
		"root":     ROOT,
		"repeated": REPEATED,
		"optional": OPTIONAL,
		"signed":   SIGNED,
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
