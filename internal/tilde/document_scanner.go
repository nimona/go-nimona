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

func unmarshalValue(v *fastjson.Value) (Value, error) {
	switch v.Type() {
	case fastjson.TypeString:
		s, err := v.StringBytes()
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling string: %w", err)
		}
		return String(s), nil
	case fastjson.TypeNumber:
		f, err := v.Float64()
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling float64: %w", err)
		}
		if f < 0 {
			return Int64(int64(f)), nil
		}
		return Uint64(uint64(f)), nil
	case fastjson.TypeTrue:
		return Bool(true), nil
	case fastjson.TypeFalse:
		return Bool(false), nil
	case fastjson.TypeNull:
		// do nothing
		return nil, nil
	case fastjson.TypeObject:
		o, err := v.Object()
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling object: %w", err)
		}
		nm := Map{}
		var errs error
		o.Visit(func(k []byte, v *fastjson.Value) {
			var nv Value
			nv, err = unmarshalValue(v)
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
			// Check hints
			ck, kind := getKindHint(string(k))
			switch kind {
			case KindRef:
				nb, err := v.StringBytes()
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nb, err = base64.StdEncoding.DecodeString(string(nb))
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nv = Ref(nb)
			case KindBytes:
				nb, err := v.StringBytes()
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nb, err = base64.StdEncoding.DecodeString(string(nb))
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nv = Bytes(nb)
			case KindInt64:
				nb, err := v.Float64()
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nv = Int64(int64(nb))
			case KindUint64:
				nb, err := v.Float64()
				if err != nil {
					errs = multierror.Append(
						errs,
						fmt.Errorf("error unmarshaling object value '%s': %w", string(k), err),
					)
					return
				}
				nv = Uint64(uint64(nb))
			}
			nm[ck] = nv
			fmt.Printf("key: %s, value: %v\n", ck, nv)
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
			ne, err := unmarshalValue(v)
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

func getKindHint(k string) (string, ValueKind) {
	ck, hint, _ := strings.Cut(k, ":")
	if hint == "" {
		return k, KindInvalid
	}

	switch hint {
	case "s":
		return ck, KindString
	case "b":
		return ck, KindBool
	case "i":
		return ck, KindInt64
	case "u":
		return ck, KindUint64
	case "d":
		return ck, KindBytes
	case "r":
		return ck, KindRef
	case "a":
		// TODO: support nested arrays
		return ck, KindList
	case "m":
		return ck, KindMap
	default:
		return ck, KindInvalid
	}
}
