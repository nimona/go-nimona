package main

import (
	"bufio"
	"bytes"
	"io"
)

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

func (s *Scanner) unread() {
	s.r.UnreadRune()
}

func (s *Scanner) scanWhiteSpace() (Token, string) {
	buf := bytes.Buffer{}
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WHITESPACE, buf.String()
}

func (s *Scanner) scanTEXT() (Token, string) {
	buf := bytes.Buffer{}
	buf.WriteRune(s.read())

	for {
		r := s.read()
		if r == eof {
			break
		} else if !isText(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	text := buf.String()
	if token, ok := keywords[text]; ok {
		return token, text
	}

	return TEXT, text
}

func (s *Scanner) Scan() (Token, string) {
	r := s.read()

	if isEOF(r) {
		return EOF, ""
	}

	if isWhitespace(r) {
		s.unread()
		return s.scanWhiteSpace()
	}

	if isText(r) {
		s.unread()
		return s.scanTEXT()
	}

	if isOpenBrace(r) {
		return OBRACE, string(r)
	}

	if isEndBrace(r) {
		return EBRACE, string(r)
	}

	return UNKNOWN, string(r)
}
