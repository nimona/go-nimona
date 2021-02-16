package main

import (
	"bytes"
	"strings"
	"text/template"

	"nimona.io/pkg/object"
)

var primitives = map[string]struct {
	Hint      string
	Type      string
	IsObject  bool
	IsPrimary bool
}{
	"nimona.io/crypto.PrivateKey": {
		Hint:      "s",
		Type:      "crypto.PrivateKey",
		IsObject:  false,
		IsPrimary: true,
	},
	"nimona.io/crypto.PublicKey": {
		Hint:      "s",
		Type:      "crypto.PublicKey",
		IsObject:  false,
		IsPrimary: true,
	},
	"nimona.io/object.CID": {
		Hint:      "s",
		Type:      "object.CID",
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

type (
	{{- range $object := .Objects }}
	{{ structName $object.Name }} struct {
		Metadata object.Metadata {{ tagMetadata }}
		{{- range $member := $object.Members }}
			{{- if $member.IsRepeated }}
				{{- if $member.IsObject }}
					{{ $member.Name }} []*{{ memberType $member.Type }} {{ tag $member }}
				{{- else }}
					{{ $member.Name }} []{{ memberType $member.Type }} {{ tag $member }}
				{{- end }}
			{{- else if $member.IsPrimitive }}
				{{ $member.Name }} {{ memberType $member.Type }} {{ tag $member }}
			{{- else if $member.IsObject }}
				{{ $member.Name }} *{{ memberType $member.Type }} {{ tag $member }}
			{{- else }}
				{{ $member.Name }} {{ memberType $member.Type }} {{ tag $member }}
			{{- end }}
		{{- end }}
	}
	{{- end }}
)

{{ range $object := .Objects }}
func (e *{{ structName $object.Name }}) Type() string {
	return "{{ $object.Name }}"
}

func (e {{ structName $object.Name }}) ToObject() *object.Object {
	r := &object.Object{
		Type: "{{ $object.Name }}",
		Metadata: e.Metadata,
		Data: object.Map{},
	}
	{{- range $member := $object.Members }}
		{{- if $member.IsRepeated }}
			// if $member.IsRepeated
			if len(e.{{ $member.Name }}) > 0 {
			{{- if $member.IsObject }}
				// if $member.IsObject
				rv := make(object.ObjectArray, len(e.{{ $member.Name }}))
				for i, v := range e.{{ $member.Name }} {
					{{- if eq $member.Type "nimona.io/object.Object" }}
					rv[i] = v
					{{- else }}
					rv[i] = v.ToObject()
					{{- end }}
				}
				r.Data["{{ $member.Tag }}"] = rv
			{{- else }}
				// else
				// r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
				{{- if eq $member.Hint "s" }}
					rv := make(object.StringArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.String(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else if eq $member.Hint "i" }}
					rv := make(object.IntArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.Int(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else if eq $member.Hint "d" }}
					rv := make(object.DataArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.Data(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else if eq $member.Hint "u" }}
					rv := make(object.UintArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.Uint(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else if eq $member.Hint "f" }}
					rv := make(object.FloatArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.Float(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else if eq $member.Hint "b" }}
					rv := make(object.BoolArray, len(e.{{ $member.Name }}))
					for i, iv := range e.{{ $member.Name }} {
						rv[i] = object.Bool(iv)
					}
					r.Data["{{ $member.Tag }}"] = rv
				{{- else }}
					// e.{{ $member.Name }} = object.FromMap(t)
				{{- end }}
			{{- end }}
			}
		{{- else if $member.IsPrimitive }}
			// else if $member.IsPrimitive
			r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
		{{- else if $member.IsObject }}
			// else if $member.IsObject
			if e.{{ $member.Name }} != nil {
				{{- if eq $member.Type "nimona.io/object.Object" }}
					r.Data["{{ $member.Tag }}"] = e.{{ $member.Name }}
				{{- else }}
					r.Data["{{ $member.Tag }}"] = e.{{ $member.Name }}.ToObject()
				{{- end }}
			}
		{{- else }}
			// else
			// r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- if eq $member.Hint "s" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else if eq $member.Hint "i" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else if eq $member.Hint "d" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else if eq $member.Hint "u" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else if eq $member.Hint "f" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else if eq $member.Hint "b" }}
				r.Data["{{ $member.Tag }}"] = {{ fromPrimitive $member }}(e.{{ $member.Name }})
			{{- else }}
				// e.{{ $member.Name }} = object.FromMap(t)
			{{- end }}
		{{- end }}
	{{- end }}
	return r
}

func (e *{{ structName $object.Name }}) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	{{- range $member := $object.Members }}
	{{- if $member.IsObject }}
		{{- if $member.IsRepeated }}
		if v, ok := o.Data["{{ $member.Tag }}"]; ok {
			if t, ok := v.(object.MapArray); ok {
				e.{{ $member.Name }} = make([]*{{ memberType $member.Type }}, len(t))
				for i, iv := range t {
					{{- if eq $member.Type "nimona.io/object.Object" }}
						eo := object.FromMap(iv)
						e.{{ $member.Name }}[i] = &eo
					{{- else }}
						es := &{{ memberType $member.Type }}{}
						eo := object.FromMap(iv)
						es.FromObject(eo)
						e.{{ $member.Name }}[i] = es
					{{- end }}
				}
			} else if t, ok := v.(object.ObjectArray); ok {
				e.{{ $member.Name }} = make([]*{{ memberType $member.Type }}, len(t))
				for i, iv := range t {
					{{- if eq $member.Type "nimona.io/object.Object" }}
						e.{{ $member.Name }}[i] = iv
					{{- else }}
						es := &{{ memberType $member.Type }}{}
						es.FromObject(iv)
						e.{{ $member.Name }}[i] = es
					{{- end }}
				}
			}
		}
		{{- else }}
		if v, ok := o.Data["{{ $member.Tag }}"]; ok {
			if t, ok := v.(object.Map); ok {
				{{- if eq $member.Type "nimona.io/object.Object" }}
					e.{{ $member.Name }} = object.FromMap(t)
				{{- else }}
					es := &{{ memberType $member.Type }}{}
					eo := object.FromMap(t)
					es.FromObject(eo)
					e.{{ $member.Name }} = es
				{{- end }}
			} else if t, ok := v.(*object.Object); ok {
				{{- if eq $member.Type "nimona.io/object.Object" }}
					e.{{ $member.Name }} = t
				{{- else }}
					es := &{{ memberType $member.Type }}{}
					es.FromObject(t)
					e.{{ $member.Name }} = es
				{{- end }}
			}
		}
		{{- end }}
	{{- else }}
		{{- if $member.IsRepeated }}
		if v, ok := o.Data["{{ $member.Tag }}"]; ok {
			{{- if eq $member.Hint "s" }}
				if t, ok := v.(object.StringArray); ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else if eq $member.Hint "i" }}
				if t, ok := v.(object.Int);Array ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else if eq $member.Hint "d" }}
				if t, ok := v.(object.DataArray); ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else if eq $member.Hint "u" }}
				if t, ok := v.(object.UintArray); ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else if eq $member.Hint "f" }}
				if t, ok := v.(object.FloatArray); ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else if eq $member.Hint "b" }}
				if t, ok := v.(object.BoolArray); ok {
					rv := make([]{{ $member.Type }}, len(t))
					for i, iv := range t {
						rv[i] = {{ $member.Type }}(iv)
					}
					e.{{ $member.Name }} = rv
				}
			{{- else }}
				// e.{{ $member.Name }} = object.FromMap(t)
			{{- end }}
		}
		{{- else }}
			if v, ok := o.Data["{{ $member.Tag }}"]; ok {
				{{- if eq $member.Hint "s" }}
					if t, ok := v.(object.String); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else if eq $member.Hint "i" }}
					if t, ok := v.(object.Int); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else if eq $member.Hint "d" }}
					if t, ok := v.(object.Data); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else if eq $member.Hint "u" }}
					if t, ok := v.(object.Uint); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else if eq $member.Hint "f" }}
					if t, ok := v.(object.Float); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else if eq $member.Hint "b" }}
					if t, ok := v.(object.Bool); ok {
						e.{{ $member.Name }} = {{ memberType $member.Type }}(t)
					}
				{{- else }}
					// e.{{ $member.Name }} = v.PrimitiveHinted().({{ memberType $member.Type }})
				{{- end }}
			}
		{{- end }}
	{{- end }}
{{- end }}
	return nil
}
{{ end }}
`

func Generate(doc *Document, output string) ([]byte, error) {
	originalImports := map[string]string{}
	t, err := template.New("tpl").Funcs(template.FuncMap{
		"tag": func(m Member) string {
			// NOTE(geoah): removed until we re-introduce encode/decode
			return ""
		},
		"key": func(m Member) string {
			h := m.Hint
			if m.Type == "nimona.io/object.Object" {
				h = "o"
			}
			if m.IsRepeated {
				h = "a" + h
			}
			return m.Tag + ":" + h
		},
		"fromPrimitive": func(m Member) string {
			h := m.Hint
			if m.Type == "nimona.io/object.Object" {
				h = "o"
			}
			if m.IsRepeated {
				h = "a" + h
			}
			switch object.Hint(h) {
			case object.BoolHint:
				return "object.Bool"
			case object.DataHint:
				return "object.Data"
			case object.FloatHint:
				return "object.Float"
			case object.IntHint:
				return "object.Int"
			case object.MapHint:
				return "object.Map"
			case object.StringHint:
				return "object.String"
			case object.UintHint:
				return "object.Uint"
			case object.BoolArrayHint:
				return "object.ToBoolArray"
			case object.DataArrayHint:
				return "object.ToDataArray"
			case object.FloatArrayHint:
				return "object.ToFloatArray"
			case object.IntArrayHint:
				return "object.ToIntArray"
			case object.MapArrayHint:
				return "object.ToMapArray"
			case object.StringArrayHint:
				return "object.ToStringArray"
			case object.UintArrayHint:
				return "object.ToUintArray"
			}
			panic("unknown primitive " + m.Hint)
		},
		"toPrimitive": func(m Member) string {
			h := m.Hint
			if m.Type == "nimona.io/object.Object" {
				h = "o"
			}
			if m.IsRepeated {
				h = "a" + h
			}
			switch object.Hint(h) {
			case object.BoolHint:
				return "bool"
			case object.DataHint:
				return "[]byte"
			case object.FloatHint:
				return "float64"
			case object.IntHint:
				return "int64"
			case object.StringHint:
				return "string"
			case object.UintHint:
				return "uint64"
			case object.BoolArrayHint:
				return "object.FromBoolArray"
			case object.DataArrayHint:
				return "object.FromDataArray"
			case object.FloatArrayHint:
				return "object.FromFloatArray"
			case object.IntArrayHint:
				return "object.FromIntArray"
			case object.MapArrayHint:
				return "object.FromMapArray"
			case object.StringArrayHint:
				return "object.FromStringArray"
			case object.UintArrayHint:
				return "object.FromUintArray"
			}
			panic("unknown primitive " + m.Hint)
		},
		"tagMetadata": func() string {
			return "`nimona:\"metadata:m,omitempty\"`"
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
		"memberType": func(name string) string {
			for alias, pkg := range originalImports {
				name = strings.Replace(name, pkg, alias, 1)
			}
			ps := strings.Split(name, "/")
			return strings.TrimPrefix(ps[len(ps)-1], doc.PackageAlias+".")
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
				if strings.HasSuffix(mv.Type, pk) {
					mv.Hint = pv.Hint
					mv.Type = pv.Type
					mv.IsObject = pv.IsObject
					mv.IsPrimitive = pv.IsPrimary
					break
				}
			}
		}
	}

	doc.Imports["json"] = "encoding/json"
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
