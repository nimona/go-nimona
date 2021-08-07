package main

import (
	"bytes"
	"strings"
	"text/template"

	"nimona.io/pkg/tilde"
)

var primitives = map[string]struct {
	Hint      string
	Type      string
	IsObject  bool
	IsPrimary bool
}{
	"nimona.io/tilde.Digest": {
		Hint:      "s",
		Type:      "tilde.Digest",
		IsObject:  false,
		IsPrimary: true,
	},
}

// nolint
var tpl = `// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package {{ .PackageAlias }}

import (
	"fmt"
	{{ range $alias, $pkg := .Imports }}
	{{ $alias }} "{{ $pkg }}"
	{{- end }}
)

{{- range $object := .Objects }}
const {{ structName $object.Name }}Type = "{{ $object.Name }}"

type {{ structName $object.Name }} struct {
	Metadata object.Metadata {{ tagMetadata $object.Name }}
	{{- range $member := $object.Members }}
		{{- if $member.IsRepeated }}
			{{ $member.Name }} []{{ memberType $member true }} {{ tag $member }}
		{{- else if $member.IsPrimitive }}
			{{ $member.Name }} {{ memberType $member true }} {{ tag $member }}
		{{- else }}
			{{ $member.Name }} {{ memberType $member true}} {{ tag $member }}
		{{- end }}
	{{- end }}
}

{{ end }}
`

func Generate(doc *Document, output string) ([]byte, error) {
	originalImports := map[string]string{}
	t, err := template.New("tpl").Funcs(template.FuncMap{
		"tag": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			return "`nimona:\"" + m.Tag + ":" + h + "\"`"
		},
		"tagMetadata": func(name string) string {
			return "`nimona:\"@metadata:m,type=" + name + "\"`"
		},
		"key": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			return m.Tag + ":" + h
		},
		"fromPrimitive": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			switch tilde.Hint(h) {
			case tilde.BoolHint:
				return "object.Bool"
			case tilde.DataHint:
				return "object.Data"
			case tilde.FloatHint:
				return "object.Float"
			case tilde.IntHint:
				return "object.Int"
			case tilde.MapHint:
				return "tilde.Map"
			case tilde.StringHint:
				return "tilde.String"
			case tilde.UintHint:
				return "object.Uint"
			case tilde.BoolArrayHint:
				return "object.ToBoolArray"
			case tilde.DataArrayHint:
				return "object.ToDataArray"
			case tilde.FloatArrayHint:
				return "object.ToFloatArray"
			case tilde.IntArrayHint:
				return "object.ToIntArray"
			case tilde.MapArrayHint:
				return "object.ToMapArray"
			case tilde.StringArrayHint:
				return "object.ToStringArray"
			case tilde.UintArrayHint:
				return "object.ToUintArray"
			}
			panic("unknown hint in fromPrimitive " + m.Hint)
		},
		"toPrimitive": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			switch tilde.Hint(h) {
			case tilde.BoolHint:
				return "bool"
			case tilde.DataHint:
				return "[]byte"
			case tilde.FloatHint:
				return "float64"
			case tilde.IntHint:
				return "int64"
			case tilde.StringHint:
				return "string"
			case tilde.UintHint:
				return "uint64"
			case tilde.MapHint:
				return "tilde.Map"
			case tilde.BoolArrayHint:
				return "object.FromBoolArray"
			case tilde.DataArrayHint:
				return "object.FromDataArray"
			case tilde.FloatArrayHint:
				return "object.FromFloatArray"
			case tilde.IntArrayHint:
				return "object.FromIntArray"
			case tilde.MapArrayHint:
				return "object.FromMapArray"
			case tilde.StringArrayHint:
				return "object.FromStringArray"
			case tilde.UintArrayHint:
				return "object.FromUintArray"
			}
			panic("unknown hint in toPrimitive " + m.Hint)
		},
		"primitive": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			switch tilde.Hint(h) {
			case tilde.BoolHint:
				return "object.Bool"
			case tilde.DataHint:
				return "object.Data"
			case tilde.FloatHint:
				return "object.Float"
			case tilde.IntHint:
				return "object.Int"
			case tilde.StringHint:
				return "tilde.String"
			case tilde.MapHint:
				return "tilde.Map"
			case tilde.UintHint:
				return "object.Uint"
			case tilde.BoolArrayHint:
				return "object.BoolArray"
			case tilde.DataArrayHint:
				return "object.DataArray"
			case tilde.FloatArrayHint:
				return "object.FloatArray"
			case tilde.IntArrayHint:
				return "object.IntArray"
			case tilde.ObjectArrayHint:
				return "object.ObjectArray"
			case tilde.MapArrayHint:
				return "object.MapArray"
			case tilde.StringArrayHint:
				return "object.StringArray"
			case tilde.UintArrayHint:
				return "object.UintArray"
			}
			panic("unknown hint in primitive " + m.Hint)
		},
		"primitiveSingular": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			switch tilde.Hint(h) {
			case tilde.BoolArrayHint:
				return "object.Bool"
			case tilde.DataArrayHint:
				return "object.Data"
			case tilde.FloatArrayHint:
				return "object.Float"
			case tilde.IntArrayHint:
				return "object.Int"
			case tilde.MapArrayHint:
				return "tilde.Map"
			case tilde.ObjectArrayHint:
				return ""
			case tilde.StringArrayHint:
				return "tilde.String"
			case tilde.UintArrayHint:
				return "object.Uint"
			}
			panic("unknown hint in primitiveSingular " + m.Hint)
		},
		"marshalFunc": func(m Member) string {
			switch m.SimpleType {
			case "data":
				return "MarshalBytes"
			default:
				return "Marshal" + strings.Title(m.SimpleType)
			}
		},
		"unmarshalFunc": func(m Member) string {
			switch m.SimpleType {
			case "data":
				return "UnmarshalBytes"
			default:
				return "Unmarshal" + strings.Title(m.SimpleType)
			}
		},
		"unmarshalArg": func(m Member) string {
			h := m.Hint
			if m.IsRepeated {
				h = "a" + h
			}
			switch tilde.Hint(h) {
			case tilde.BoolHint:
				return "bool"
			case tilde.DataHint:
				return "[]byte"
			case tilde.FloatHint:
				return "float64"
			case tilde.IntHint:
				return "int64"
			case tilde.StringHint:
				return "string"
			case tilde.UintHint:
				return "uint64"
			case tilde.MapHint:
				return "tilde.Map"
			case tilde.BoolArrayHint:
				return "bool"
			case tilde.DataArrayHint:
				return "[]byte"
			case tilde.FloatArrayHint:
				return "float64"
			case tilde.IntArrayHint:
				return "int64"
			case tilde.MapArrayHint:
				return "tilde.Map"
			case tilde.ObjectArrayHint:
				return ""
			case tilde.StringArrayHint:
				return "string"
			case tilde.UintArrayHint:
				return "uint64"
			}
			panic("unknown primitive " + m.Hint)
		},
		"structName": func(name string) string {
			ps := strings.Split(name, "/")
			ps = strings.Split(ps[len(ps)-1], ".")
			nn := ""
			if len(ps) == 1 {
				nn = ucFirst(ps[0])
			} else if strings.ToLower(ps[len(ps)-2]) == strings.ToLower(doc.PackageAlias) {
				nn = ucFirst(ps[len(ps)-1])
			} else {
				nn = ucFirst(ps[len(ps)-2]) + ucFirst(ps[len(ps)-1])
			}
			if strings.HasPrefix(name, "stream:") {
				nn += "StreamRoot"
			}
			return nn
		},
		"memberType": func(m Member, dec bool) string {
			name := m.GoFullType
			for alias, pkg := range originalImports {
				name = strings.Replace(name, pkg, alias, 1)
			}
			ps := strings.Split(name, "/")
			name = strings.TrimPrefix(ps[len(ps)-1], doc.PackageAlias+".")
			if m.IsObject && m.IsOptional {
				if dec {
					name = "*" + name
				} else {
					name = "&" + name
				}
			}
			return name
		},
		"neq": func(a, b string) bool {
			return a != b
		},
		"hp": func(a, b string) bool {
			return strings.HasPrefix(a, b)
		},
		"hnp": func(a, b string) bool {
			return !strings.HasPrefix(a, b)
		},
	}).Parse(tpl)
	if err != nil {
		return nil, err
	}

	// instead of doing the same work for both top-level and stream objects, we
	// convert stream objects into top-level ones

	for _, s := range doc.Streams {
		for _, o := range s.Objects {
			switch {
			case o.IsRoot:
				o.Name = "stream:" + s.Name
			case o.IsEvent:
				o.Name = "event:" + s.Name + "." + o.Name
			default:
				o.Name = s.Name + "." + o.Name
			}
			doc.Objects = append(doc.Objects, o)
		}
	}

	for _, e := range doc.Objects {
		for _, mv := range e.Members {
			for pk, pv := range primitives {
				if strings.HasSuffix(mv.GoFullType, pk) {
					mv.Hint = pv.Hint
					mv.GoFullType = pv.Type
					mv.IsObject = pv.IsObject
					mv.IsPrimitive = pv.IsPrimary
					break
				}
			}
		}
	}

	doc.Imports["json"] = "encoding/json"
	doc.Imports["tilde"] = "nimona.io/tilde"
	doc.Imports["hint"] = "nimona.io/object/hint"

	if doc.Package != "nimona.io/object" {
		doc.Imports["object"] = "nimona.io/object"
	}
	if doc.Package != "nimona.io/stream" {
		doc.Imports["stream"] = "nimona.io/stream"
	}
	if doc.Package != "nimona.io/crypto" {
		doc.Imports["crypto"] = "nimona.io/crypto"
	}
	if doc.Package != "nimona.io/schema" {
		doc.Imports["schema"] = "nimona.io/schema"
	}

	for alias, pkg := range doc.Imports {
		originalImports[alias] = pkg
	}

	for i, pkg := range doc.Imports {
		doc.Imports[i] = strings.Replace(pkg, "nimona.io/", "nimona.io/pkg/", 1)
	}

	out := bytes.NewBuffer([]byte{})
	if err := t.Execute(out, doc); err != nil {
		return nil, err
	}

	res := out.String()
	if doc.Package == "nimona.io/object" {
		res = strings.ReplaceAll(res, "object.", "")
	}

	return []byte(res), nil
}

// lastSegment returns the last part of a namespace,
// ie lastSegment(nimona.io/stream) returns stream
func lastSegment(s string) string {
	ps := strings.Split(s, "/")
	return ps[len(ps)-1]
}
