package tilde

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-linereader"
	"github.com/valyala/fastjson"
)

type Scanner struct {
	lineReader  *linereader.Reader
	jsonScanner *fastjson.Scanner
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		lineReader:  linereader.New(r),
		jsonScanner: &fastjson.Scanner{},
	}
}

func (s *Scanner) Scan() (Map, error) {
	// read a line from the reader
	line := <-s.lineReader.Ch

	// init the json scanner
	s.jsonScanner.Init(line)

	// if the line is empty, we're done
	// TODO: not sure if this is correct, there is an issue were if there is
	// a trailing newline, the last line is read repeatedly.
	if line == "" {
		return nil, io.EOF
	}

	// if the line is not a json object, return an error
	if !s.jsonScanner.Next() {
		err := s.jsonScanner.Error()
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}

	o := s.jsonScanner.Value()
	if o.Type() != fastjson.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", o.Type())
	}

	nm, err := unmarshalValue(o)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling: %w", err)
	}

	return nm.(Map), nil
}

func unmarshalValue(v *fastjson.Value, nextHints ...Hint) (Value, error) {
	switch v.Type() {
	case fastjson.TypeString:
		if len(nextHints) == 0 {
			nextHints = []Hint{HintString}
		}
		switch nextHints[0] {
		case HintRef:
			nb, err := v.StringBytes()
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling ref: %w", err)
			}
			nb, err = base64.StdEncoding.DecodeString(string(nb))
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling ref: %w", err)
			}
			return Ref(nb), nil
		case HintBytes:
			nb, err := v.StringBytes()
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling bytes: %w", err)
			}
			nb, err = base64.StdEncoding.DecodeString(string(nb))
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling bytes: %w", err)
			}
			return Bytes(nb), nil
		case HintString:
			s, err := v.StringBytes()
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling string: %w", err)
			}
			return String(s), nil
		default:
			return nil, fmt.Errorf("unknown string type: %s", string(nextHints[0]))
		}
	case fastjson.TypeNumber:
		if len(nextHints) == 0 {
			nextHints = []Hint{HintInt64}
		}
		switch nextHints[0] {
		case HintInt64:
			nb, err := v.Float64()
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling int64: %w", err)
			}
			return Int64(int64(nb)), nil
		case HintUint64:
			nb, err := v.Float64()
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling uint64: %w", err)
			}
			return Uint64(uint64(nb)), nil
		default:
			return nil, fmt.Errorf("unknown number type: %s", string(nextHints[0]))
		}
	case fastjson.TypeTrue:
		return Bool(true), nil
	case fastjson.TypeFalse:
		return Bool(false), nil
	case fastjson.TypeNull:
		return nil, nil
	case fastjson.TypeObject:
		o, err := v.Object()
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling object: %w", err)
		}
		nm := Map{}
		var errs error
		o.Visit(func(k []byte, v *fastjson.Value) {
			ck, hs, err := getKindHints(string(k))
			if err != nil {
				errs = multierror.Append(
					errs,
					fmt.Errorf("error unmarshaling object key '%s': %w", string(k), err),
				)
				return
			}
			nv, err := unmarshalValue(v, hs...)
			if err != nil {
				errs = multierror.Append(
					errs,
					fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
				)
				return
			}
			if nv == nil {
				return
			}
			nm[ck] = nv
		})
		if errs != nil {
			return nil, errs
		}
		return nm, nil
	case fastjson.TypeArray:
		nv, err := v.Array()
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling list: %w", err)
		}
		nl := List{}
		for _, v := range nv {
			remainingHints := nextHints
			if len(remainingHints) > 0 {
				remainingHints = remainingHints[1:]
			}
			ne, err := unmarshalValue(v, remainingHints...)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling list value: %w", err)
			}
			if ne == nil {
				continue
			}
			nl = append(nl, ne)
		}
		return nl, nil
	}
	return nil, fmt.Errorf("unknown type %s", v.Type())
}

func getKindHints(k string) (string, []Hint, error) {
	ck, hintString, _ := strings.Cut(k, ":")
	if hintString == "" {
		return k, nil, nil
	}

	hints := []Hint{}
	for _, hintRune := range hintString {
		switch Hint(hintRune) {
		case HintString:
			hints = append(hints, HintString)
		case HintBool:
			hints = append(hints, HintBool)
		case HintInt64:
			hints = append(hints, HintInt64)
		case HintUint64:
			hints = append(hints, HintUint64)
		case HintBytes:
			hints = append(hints, HintBytes)
		case HintRef:
			hints = append(hints, HintRef)
		case HintList:
			hints = append(hints, HintList)
		case HintMap:
			hints = append(hints, HintMap)
		default:
			return ck, nil, fmt.Errorf("unknown hint: %s", string(hintRune))
		}
	}

	return ck, hints, nil
}
