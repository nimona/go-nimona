package nimona

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"nimona.io/internal/tilde"
)

// Notes on the code generation:
// - CBOR and JSON unmarshaling will create []interface{} for slices, so we are
//   doing the same for consistency.

type (
	DocumentInfo struct {
		Name   string
		Fields []*DocumentField
	}
	DocumentField struct {
		Tag  Tag
		Name string

		SkipUnmarshal            bool
		ImplementsDocumentValuer bool

		Type      reflect.Type
		Pkg       string
		IsPointer bool
		IsStruct  bool
		IsSlice   bool
		TildeKind tilde.ValueKind

		ElemType      reflect.Type
		IsElemPointer bool
		IsElemStruct  bool
		IsElemSlice   bool
		ElemTildeKind tilde.ValueKind
	}
	Tag struct {
		Name      string
		OmitEmpty bool
		Omit      bool
		Const     string

		// Nimona specific attributes
		DocumentType string
	}
)

func GenerateDocumentMethods(fname, pkg string, types ...interface{}) error {
	buf := new(bytes.Buffer)

	// Gather document info
	docs := []*DocumentInfo{}
	for _, t := range types {
		gti, err := documentType(t)
		if err != nil {
			return fmt.Errorf("failed to get document type: %w", err)
		}
		docType := ""
		for _, tf := range gti.Fields {
			if tf.Tag.Name == "$metadata" && tf.Tag.DocumentType != "" {
				docType = tf.Tag.DocumentType
				break
			}
		}
		if docType != "" {
			gti.Fields = append(gti.Fields, &DocumentField{
				Type: reflect.TypeOf(""), // string
				Name: "$type",
				Tag: Tag{
					Name:  "$type",
					Const: docType,
				},
				SkipUnmarshal: true,
			})
		}
		// sort fields by name
		sort.Slice(gti.Fields, func(i, j int) bool {
			return gti.Fields[i].Name < gti.Fields[j].Name
		})
		docs = append(docs, gti)
	}

	// Gather imports
	imports := map[string]struct{}{}

	// Construct the values
	values := struct {
		Package string
		Imports map[string]struct{} // pkgpath -> alias
		Types   []*DocumentInfo
	}{
		Package: pkg,
		Imports: imports,
		Types:   docs,
	}

	// Render the template
	tpl := template.
		Must(template.New("map").
			Funcs(template.FuncMap{
				"typeName": func(t reflect.Type) string {
					return strings.TrimPrefix(t.String(), pkg+".")
				},
			}).
			Parse(tpl))
	if err := tpl.Execute(buf, values); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the code
	data, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		return fmt.Errorf("failed to format source: %w", err)
	}

	// Create the file
	fi, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Write the file
	_, err = fi.Write(data)
	defer fi.Close()
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func nameIsExported(name string) bool {
	if name == "" || name == "_" {
		return false
	}
	return strings.ToUpper(name[0:1]) == name[0:1]
}

func documentType(i interface{}) (*DocumentInfo, error) {
	t := reflect.TypeOf(i)

	pkg := t.PkgPath()

	out := DocumentInfo{
		Name: t.Name(),
	}

	typeDocumentValuer := reflect.TypeOf((*DocumentValuer)(nil)).Elem()
	typeTildeValue := reflect.TypeOf((*tilde.Value)(nil)).Elem()

	typeToTildeKind := func(t reflect.Type) (tilde.ValueKind, error) {
		switch t.Kind() {
		case reflect.Map, reflect.Struct:
			return tilde.KindMap, nil
		case reflect.String:
			return tilde.KindString, nil
		case reflect.Bool:
			return tilde.KindBool, nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return tilde.KindInt64, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return tilde.KindUint64, nil
		case reflect.Float32, reflect.Float64:
			return tilde.KindInvalid, fmt.Errorf("floats are not supported")
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 {
				return tilde.KindBytes, nil
			}
			return tilde.KindList, nil
		}

		if t.Implements(typeTildeValue) {
			return tilde.KindAny, nil
		}

		if t.Name() == "DocumentHash" {
			return tilde.KindRef, nil
		}

		return tilde.KindInvalid, fmt.Errorf("unsupported type %s", t)
	}

	for i := 0; i < t.NumField(); i++ {
		refField := t.Field(i)
		if !nameIsExported(refField.Name) {
			// allow unexported fields to be used for setting the document type
			if refField.Name != "_" {
				continue
			}
			tag, _ := ParseTag(refField.Tag.Get("nimona"))
			if tag.DocumentType == "" {
				continue
			}
			out.Fields = append(out.Fields, &DocumentField{
				Type: reflect.TypeOf(""), // string
				Name: "$type",
				Tag: Tag{
					Name:  "$type",
					Const: tag.DocumentType,
				},
				SkipUnmarshal: true,
			})
			continue
		}

		docField := DocumentField{
			Name: refField.Name,
			Type: refField.Type,
			Pkg:  pkg,
		}

		if docField.Type.Kind() == reflect.Ptr {
			docField.Type = docField.Type.Elem()
			docField.IsPointer = true
		}

		tag, err := ParseTag(refField.Tag.Get("nimona"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag: %w", err)
		}

		if tag.Omit {
			continue
		}

		docField.Tag = tag

		docField.ImplementsDocumentValuer = docField.Type.Implements(
			typeDocumentValuer,
		)

		switch docField.Type.Kind() {
		case reflect.Struct:
			docField.IsStruct = true
		case reflect.Slice:
			docField.IsSlice = true
			docField.ElemType = docField.Type.Elem()
			if docField.ElemType.Kind() == reflect.Ptr {
				docField.ElemType = docField.ElemType.Elem()
				docField.IsElemPointer = true
			}
			if docField.ElemType.Kind() == reflect.Struct {
				docField.IsElemStruct = true
			}
			if docField.ElemType.Kind() == reflect.Slice {
				docField.IsElemSlice = true
			}
		}

		tildeKind, err := typeToTildeKind(docField.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert type to tilde kind: %w", err)
		}
		docField.TildeKind = tildeKind

		if docField.TildeKind == tilde.KindList {
			tildeKind, err := typeToTildeKind(docField.ElemType)
			if err != nil {
				return nil, fmt.Errorf("failed to convert type to tilde kind: %w", err)
			}
			docField.ElemTildeKind = tildeKind
		}

		out.Fields = append(out.Fields, &docField)
	}

	return &out, nil
}

func ParseTag(tagString string) (Tag, error) {
	tag := Tag{}

	tagString = strings.TrimSpace(tagString)
	tagParts := strings.Split(tagString, ",")

	// Parse the tag name
	tag.Name = tagParts[0]
	if len(tagParts) == 1 {
		return tag, nil
	}

	// Parse the tag params
	for _, tagPart := range tagParts[1:] {
		tagPart = strings.TrimSpace(tagPart)
		if tagPart == "" {
			continue
		}
		switch tagPart {
		case "omitempty":
			tag.OmitEmpty = true
			continue
		case "omit":
			tag.Omit = true
			continue
		}
		tagPartKey, tagPartValue, isKV := strings.Cut(tagPart, "=")
		if !isKV {
			continue
		}
		switch tagPartKey {
		case "const":
			tag.Const = tagPartValue
			continue
		case "name":
			tag.Name = tagPartValue
			continue
		case "type":
			tag.DocumentType = tagPartValue
			continue
		}
		return tag, fmt.Errorf("unknown tag param: %s", tagPart)
	}

	return tag, nil
}

var tpl = `

// Code generated by nimona.io. DO NOT EDIT.

package {{ .Package }}

import (
	"github.com/vikyd/zero"
	{{ range $pkgPath, $pkgAlias := .Imports }}
	"{{ $pkgPath }}"
	{{ end }}

	"nimona.io/internal/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

{{- range .Types }}
func (t *{{ .Name }}) Document() *Document {
	return NewDocument(t.Map())
}

func (t *{{ .Name }}) Map() tilde.Map {
	m := tilde.Map{}
	{{ range .Fields }}
		// # t.{{ .Name }}
		//
		// Type: {{ .Type }}, Kind: {{ .Type.Kind }}, TildeKind: {{ .TildeKind.Name }}
		// IsSlice: {{ .IsSlice }}, IsStruct: {{ .IsStruct }}, IsPointer: {{ .IsPointer }}
		{{- if .ElemType }}
		//
		// ElemType: {{ .ElemType }}, ElemKind: {{ .ElemType.Kind }}
		// IsElemSlice: {{ .IsElemSlice }}, IsElemStruct: {{ .IsElemStruct }}, IsElemPointer: {{ .IsElemPointer }}
		{{- end }}
		{{- if .Tag.Const }}
			{
				m.Set("{{ .Tag.Name }}", tilde.String({{ .Tag.Const | printf "%q" }}))
			}
			{{ continue }}
		{{- end }}
		{
		{{- if .Tag.OmitEmpty }}
		if !zero.IsZeroVal(t.{{ .Name }}) {
		{{- end }}
		{{- if eq .Type.String "nimona.DocumentHash" }}
			m.Set("{{ .Tag.Name }}", tilde.Ref(t.{{ .Name }}[:]))
		{{- else if and .IsSlice .IsElemStruct }}
			sm := tilde.List{}
			for _, v := range t.{{ .Name }} {
				if !zero.IsZeroVal(t.{{ .Name }}) {
					sm = append(sm, v.Map())
				}
			}
			m.Set("{{ .Tag.Name }}", sm)
		{{- else if .IsStruct }}
			m.Set("{{ .Tag.Name }}", t.{{ .Name }}.Map())
		{{- else if and (.IsSlice) (eq .ElemType.String "uint8") }}
			m.Set("{{ .Tag.Name }}", tilde.{{ .TildeKind.Name }}(t.{{ .Name }}))
		{{- else if .IsSlice }}
			s := make(tilde.List, len(t.{{ .Name }}))
			for i, v := range t.{{ .Name }} {
				s[i] = tilde.{{ .ElemTildeKind.Name }}(v)
			}
			m.Set("{{ .Tag.Name }}", s)
		{{- else }}
			m.Set("{{ .Tag.Name }}", tilde.{{ .TildeKind.Name }}(t.{{ .Name }}))
		{{- end }}
		{{- if .Tag.OmitEmpty }}
		}
		{{- end }}
		}
	{{ end }}
	return m
}

func (t *{{ .Name }}) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *{{ .Name }}) FromMap(d tilde.Map) error {
	*t = {{ .Name }}{}
	{{ range .Fields }}
		{{- if .Tag.Const }}
			{{ continue }}
		{{- end }}
		{{- if .SkipUnmarshal }}
			{{ continue }}
		{{- end }}
		// # t.{{ .Name }}
		//
		// Type: {{ .Type }}, Kind: {{ .Type.Kind }}, TildeKind: {{ .TildeKind.Name }}
		// IsSlice: {{ .IsSlice }}, IsStruct: {{ .IsStruct }}, IsPointer: {{ .IsPointer }}
		{{- if .ElemType }}
		//
		// ElemType: {{ .ElemType }}, ElemKind: {{ .ElemType.Kind }}, ElemTildeKind: {{ .ElemTildeKind.Name }}
		// IsElemSlice: {{ .IsElemSlice }}, IsElemStruct: {{ .IsElemStruct }}, IsElemPointer: {{ .IsElemPointer }}
		{{- end }}
		{
		{{- if and .IsSlice .IsElemStruct }}
			{{- if .IsElemPointer }}
			sm := []*{{ typeName .ElemType }}{} // {{ typeName .ElemType }}
			{{- else }}
			sm := []{{ typeName .ElemType }}{}
			{{- end }}
			if vs, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if vs, ok := vs.(tilde.List); ok {
					for _, vi := range vs {
						if v, ok := vi.(tilde.Map); ok {
							{{- if .IsElemPointer }}
								e := &{{ typeName .ElemType }}{}
							{{- else }}
								e := {{ typeName .ElemType }}{}
							{{- end }}
							d := NewDocument(v)
							e.FromDocument(d)
							sm = append(sm, e)
						}
					}
				}
			}
			if len(sm) > 0 {
				t.{{ .Name }} = sm
			}
		{{- else if eq .Type.String "nimona.Document" }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.Map); ok {
					t.{{ .Name }} = NewDocument(v)
				}
			}
		{{- else if .IsStruct }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.Map); ok {
					e := {{ typeName .Type }}{}
					d := NewDocument(v)
					e.FromDocument(d)
					{{- if .IsPointer }}
						t.{{ .Name }} = &e
					{{- else }}
						t.{{ .Name }} = e
					{{- end }}
				}
			}
		{{- else if eq .Type.String "nimona.DocumentHash" }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.Ref); ok {
					copy(t.{{ .Name }}[:], v)
				}
			}
		{{- else if eq .Type.String "[]uint8" }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.Bytes); ok {
					t.{{ .Name }} = []byte(v)
				}
			}
		{{- else if eq .TildeKind.Name "List" }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.{{ .TildeKind.Name }}); ok {
					s := make({{ .Type }}, len(v))
					for i, vi := range v {
						if vi, ok := vi.(tilde.{{ .ElemTildeKind.Name }}); ok {
							s[i] = {{ .ElemType }}(vi)
						}
					}
					t.{{ .Name }} = s
				}
			}
		{{- else }}
			if v, err := d.Get("{{ .Tag.Name }}"); err == nil {
				if v, ok := v.(tilde.{{ .TildeKind.Name }}); ok {
					t.{{ .Name }} = {{ typeName .Type }}(v)
				}
			}
		{{- end }}
		}
	{{ end }}

	return nil
}
{{- end }}

`
