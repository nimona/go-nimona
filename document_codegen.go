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

		SkipUnmarshal bool

		Type      reflect.Type
		Pkg       string
		IsPointer bool
		IsStruct  bool
		IsSlice   bool

		ElemType      reflect.Type
		IsElemPointer bool
		IsElemStruct  bool
		IsElemSlice   bool
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

func GenerateDocumentMapMethods(fname, pkg string, types ...interface{}) error {
	buf := new(bytes.Buffer)

	// Gather document info
	docs := []*DocumentInfo{}
	for _, t := range types {
		gti, err := documentType(t)
		docType := ""
		if err != nil {
			return fmt.Errorf("failed to document type: %w", err)
		}
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
					if t.PkgPath() == pkg {
						return strings.TrimPrefix(t.String(), pkg+".")
					}
					return t.Name()
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
)

var _ = zero.IsZeroVal

{{- range .Types }}
func (t *{{ .Name }}) DocumentMap() DocumentMap {
	m := DocumentMap{}
	{{ range .Fields }}
		// # t.{{ .Name }}
		//
		// Type: {{ .Type }}, Kind: {{ .Type.Kind }}
		// IsSlice: {{ .IsSlice }}, IsStruct: {{ .IsStruct }}, IsPointer: {{ .IsPointer }}
		{{- if .ElemType }}
		//
		// ElemType: {{ .ElemType }}, ElemKind: {{ .ElemType.Kind }}
		// IsElemSlice: {{ .IsElemSlice }}, IsElemStruct: {{ .IsElemStruct }}, IsElemPointer: {{ .IsElemPointer }}
		{{- end }}
		{{- if .Tag.Const }}
			{
				m["{{ .Tag.Name }}"] = {{ .Tag.Const | printf "%q" }}
			}
			{{ continue }}
		{{- end }}
		{
		{{- if .Tag.OmitEmpty }}
		if !zero.IsZeroVal(t.{{ .Name }}) {
		{{- end }}
		{{- if and .IsSlice .IsElemStruct }}
			sm := []any{}
			for _, v := range t.{{ .Name }} {
				if !zero.IsZeroVal(t.{{ .Name }}) {
					sm = append(sm, v.DocumentMap())
				}
			}
			m["{{ .Tag.Name }}"] = sm
		{{- else if .IsStruct }}
			m["{{ .Tag.Name }}"] = t.{{ .Name }}.DocumentMap()
		{{- else if and (.IsSlice) (eq .ElemType.String "uint8") }}
			m["{{ .Tag.Name }}"] = []byte(t.{{ .Name }})
		{{- else if .IsSlice }}
			s := make([]any, len(t.{{ .Name }}))
			for i, v := range t.{{ .Name }} {
				s[i] = v
			}
			m["{{ .Tag.Name }}"] = s
		{{- else }}
			m["{{ .Tag.Name }}"] = t.{{ .Name }}
		{{- end }}
		{{- if .Tag.OmitEmpty }}
		}
		{{- end }}
		}
	{{ end }}
	return m
}

func (t *{{ .Name }}) FromDocumentMap(m DocumentMap) {
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
		// Type: {{ .Type }}, Kind: {{ .Type.Kind }}
		// IsSlice: {{ .IsSlice }}, IsStruct: {{ .IsStruct }}, IsPointer: {{ .IsPointer }}
		{{- if .ElemType }}
		//
		// ElemType: {{ .ElemType }}, ElemKind: {{ .ElemType.Kind }}
		// IsElemSlice: {{ .IsElemSlice }}, IsElemStruct: {{ .IsElemStruct }}, IsElemPointer: {{ .IsElemPointer }}
		{{- end }}
		{
		{{- if and .IsSlice .IsElemStruct }}
			{{- if .IsElemPointer }}
			sm := []*{{ typeName .ElemType }}{} // {{ typeName .ElemType }}
			{{- else }}
			sm := []{{ typeName .ElemType }}{}
			{{- end }}
			if vs, ok := m["{{ .Tag.Name }}"].([]any); ok {
				for _, vi := range vs {
					v, ok := vi.(DocumentMap)
					if ok {
						{{- if .IsElemPointer }}
							e := &{{ typeName .ElemType }}{}
						{{- else }}
							e := {{ typeName .ElemType }}{}
						{{- end }}
						e.FromDocumentMap(v)
						sm = append(sm, e)
					}
				}
			}
			if len(sm) > 0 {
				t.{{ .Name }} = sm
			}
		{{- else if .IsStruct }}
			if v, ok := m["{{ .Tag.Name }}"].(DocumentMap); ok {
				e := {{ typeName .Type }}{}
				e.FromDocumentMap(v)
				{{- if .IsPointer }}
					t.{{ .Name }} = &e
				{{- else }}
					t.{{ .Name }} = e
				{{- end }}
			}
		{{- else if eq .Type.String "string" }}
			if v, ok := m["{{ .Tag.Name }}"].(string); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "int" }}
			if v, ok := m["{{ .Tag.Name }}"].(int); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "int64" }}
			if v, ok := m["{{ .Tag.Name }}"].(int64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "uint64" }}
			if v, ok := m["{{ .Tag.Name }}"].(uint64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "float64" }}
			if v, ok := m["{{ .Tag.Name }}"].(float64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "bool" }}
			if v, ok := m["{{ .Tag.Name }}"].(bool); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]byte" }}
			if v, ok := m["{{ .Tag.Name }}"].([]byte); ok {
				t.{{ .Name }} = v
			}
		{{/*- else if eq .Type.String "[]uint8" */}}
		{{- else if and (.IsSlice) (eq .ElemType.String "uint8") }}
			if v, ok := m["{{ .Tag.Name }}"].([]byte); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[][]uint8" }}
			if v, ok := m["{{ .Tag.Name }}"].([]any); ok {
				s := make([][]byte, len(v))
				for i, vi := range v {
					s[i] = vi.([]byte)
				}
				t.{{ .Name }} = s
			}
		{{- else if .IsSlice }}
			if v, ok := m["{{ .Tag.Name }}"].([]any); ok {
				s := make([]{{ typeName .ElemType }}, len(v))
				for i, vi := range v {
					s[i] = vi.({{ typeName .ElemType }})
				}
				t.{{ .Name }} = s
			}
		{{- else if eq .Type.String "[]string" }}
			if v, ok := m["{{ .Tag.Name }}"].([]string); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]int" }}
			if v, ok := m["{{ .Tag.Name }}"].([]int); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]int64" }}
			if v, ok := m["{{ .Tag.Name }}"].([]int64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]uint64" }}
			if v, ok := m["{{ .Tag.Name }}"].([]uint64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]float64" }}
			if v, ok := m["{{ .Tag.Name }}"].([]float64); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[]bool" }}
			if v, ok := m["{{ .Tag.Name }}"].([]bool); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[][]byte" }}
			if v, ok := m["{{ .Tag.Name }}"].([][]byte); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "[][]uint8" }}
			if v, ok := m["{{ .Tag.Name }}"].([][]uint8); ok {
				t.{{ .Name }} = [][]byte(v)
			}
		{{- else if and (.IsSlice) (eq .ElemType.String "uint8") }}
			if v, ok := m["{{ .Tag.Name }}"].([]byte); ok {
				t.{{ .Name }} = v
			}
		{{- else if .IsSlice }}
			if v, ok := m["{{ .Tag.Name }}"].([]{{ .ElemType }}); ok {
				t.{{ .Name }} = v
			}
		{{- else if eq .Type.String "nimona.DocumentHash" }}
			if v, ok := m["{{ .Tag.Name }}"].([]uint8); ok {
				copy(t.DocumentHash[:], v)
			}
		{{- else if eq .Type.String "nimona.DocumentMap" }}
			if v, ok := m["{{ .Tag.Name }}"].(DocumentMap); ok {
				t.{{ .Name }} = v
			} else if v, ok := m["{{ .Tag.Name }}"].(map[string]any); ok {
				t.{{ .Name }} = DocumentMap(v)
			}
		{{- else }}
			// TODO: Unsupported type {{ .Type.String }}
		{{- end }}
		}
	{{ end }}
}
{{- end }}

`
