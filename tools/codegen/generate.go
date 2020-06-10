package main

import (
	"bytes"
	"strings"
	"text/template"
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
	"nimona.io/object.Hash": {
		Hint:      "s",
		Type:      "object.Hash",
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
		raw object.Object
		Stream object.Hash
		Parents []object.Hash
		Owners []crypto.PublicKey
		Policy object.Policy
		Signatures []object.Signature
		{{- range $member := $object.Members }}
			{{- if $member.IsRepeated }}
				{{- if $member.IsObject }}
					{{ $member.Name }} []*{{ memberType $member.Type }}
				{{- else }}
					{{ $member.Name }} []{{ memberType $member.Type }}
				{{- end }}
			{{- else if $member.IsPrimitive }}
				{{ $member.Name }} {{ memberType $member.Type }}
			{{- else if $member.IsObject }}
				{{ $member.Name }} *{{ memberType $member.Type }}
			{{- else }}
				{{ $member.Name }} {{ memberType $member.Type }}
			{{- end }}
		{{- end }}
	}
	{{- end }}
)

{{ range $object := .Objects }}
func (e {{ structName $object.Name }}) GetType() string {
	return "{{ $object.Name }}"
}

{{ if hnp $object.Name "nimona.io/object.Schema" }}
func (e {{ structName $object.Name }}) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
		{{- range $member := $object.Members }}
			&object.SchemaProperty{
				Name: "{{ $member.Tag }}",
				Type: "{{ $member.SimpleType }}",
				Hint: "{{ $member.Hint }}",
				IsRepeated: {{ if $member.IsRepeated }} true {{ else }} false {{ end }},
				IsOptional: {{ if $member.IsOptional }} true {{ else }} false {{ end }},
			},
		{{- end }}
		},
	}
}
{{ end }}

func (e {{ structName $object.Name }}) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("{{ $object.Name }}")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.AddSignature(e.Signatures...)
	o = o.SetPolicy(e.Policy)
	{{- range $member := $object.Members }}
		{{- if $member.IsObject }}
			{{- if $member.IsRepeated }}
			if len(e.{{ $member.Name }}) > 0 {
				v := object.List{}
				for _, iv := range e.{{ $member.Name }} {
					v = v.Append(iv.ToObject().Raw())
				}
				o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", v)
			}
			{{- else }}
			if e.{{ $member.Name }} != nil {
				o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }}.ToObject().Raw())
			}
			{{- end }}
		{{- else }}
			{{- if $member.IsRepeated }}
				if len(e.{{ $member.Name }}) > 0 {
					v := object.List{}
					for _, iv := range e.{{ $member.Name }} {
						{{- if eq $member.Hint "s" }}
							v = v.Append(object.String(iv))
						{{- else if eq $member.Hint "b" }}
							v = v.Append(object.Bool(iv))
						{{- else if eq $member.Hint "d" }}
							v = v.Append(object.Bytes(iv))
						{{- else if eq $member.Hint "i" }}
							v = v.Append(object.Int(iv))
						{{- else if eq $member.Hint "u" }}
							// TODO(geoah) uints not implemented
						{{- else if eq $member.Hint "f" }}
							v = v.Append(object.Float(iv))
						{{- else }}
							// TODO missing type hint {{ $member.Hint }}, for repeated {{ $member.Name }}
						{{- end }}
					}
					o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", v)
				}
			{{- else }}
				{{- if eq $member.Hint "s" }}
				if e.{{ $member.Name }} != "" {
					o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }})
				}
				{{- else if eq $member.Hint "b" }}
				o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }})
				{{- else if eq $member.Hint "d" }}
				if len(e.{{ $member.Name }}) != 0 {
					o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }})
				}
				{{- else if eq $member.Hint "i" }}
					o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }})
				{{- else if eq $member.Hint "u" }}
					// TODO(geoah) uints not implemented
				{{- else if eq $member.Hint "f" }}
					o = o.Set("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}", e.{{ $member.Name }})
				{{- else }}
					// TODO missing type hint {{ $member.Hint }}, for {{ $member.Name }}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if hnp $object.Name "nimona.io/object.Schema" }}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	{{- end }}
	return o
}

func (e *{{ structName $object.Name }}) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signatures = o.GetSignatures()
	e.Policy = o.GetPolicy()
	{{- range $member := $object.Members }}
	{{- if $member.IsObject }}
		{{- if $member.IsRepeated }}
		if v := data.Value("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}"); v != nil && v.IsList() {
			m := v.PrimitiveHinted().([]interface{})
			e.{{ $member.Name }} = make([]*{{ memberType $member.Type }}, len(m))
			for i, iv := range m {
				{{- if eq $member.Type "nimona.io/object.Object" }}
					eo := object.FromMap(iv.(map[string]interface{}))
					e.{{ $member.Name }}[i] = &eo
				{{- else }}
					es := &{{ memberType $member.Type }}{}
					eo := object.FromMap(iv.(map[string]interface{}))
					es.FromObject(eo)
					e.{{ $member.Name }}[i] = es
				{{- end }}
			}
		}
		{{- else }}
		if v := data.Value("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}"); v != nil {
			es := &{{ memberType $member.Type }}{}
			eo := object.FromMap(v.PrimitiveHinted().(map[string]interface{}))
			es.FromObject(eo)
			e.{{ $member.Name }} = es
		}
		{{- end }}
	{{- else }}
		{{- if $member.IsRepeated }}
		if v := data.Value("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}"); v != nil && v.IsList() {
			{{- if eq $member.Hint "s" }}
				m := v.PrimitiveHinted().([]string)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else if eq $member.Hint "i" }}
				m := v.PrimitiveHinted().([]int64)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else if eq $member.Hint "d" }}
				m := v.PrimitiveHinted().([]byte)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else if eq $member.Hint "u" }}
				m := v.PrimitiveHinted().([]uint64)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else if eq $member.Hint "f" }}
				m := v.PrimitiveHinted().([]float64)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else if eq $member.Hint "b" }}
				m := v.PrimitiveHinted().([]bool)
				e.{{ $member.Name }} = make([]{{ memberType $member.Type }}, len(m))
				for i, iv := range m {
					e.{{ $member.Name }}[i] = {{ memberType $member.Type }}(iv)
				}
			{{- else }}
				// TODO missing implementation for repeated type hint {{ $member.Hint }}
			{{- end }}
		}
		{{- else }}
			if v := data.Value("{{ memberTag $member.Tag $member.Hint $member.IsRepeated }}"); v != nil {
				{{- if eq $member.Hint "s" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().(string))
				{{- else if eq $member.Hint "i" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().(int64))
				{{- else if eq $member.Hint "d" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().([]byte))
				{{- else if eq $member.Hint "u" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().(uint64))
				{{- else if eq $member.Hint "f" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().(float64))
				{{- else if eq $member.Hint "b" }}
					e.{{ $member.Name }} = {{ memberType $member.Type }}(v.PrimitiveHinted().(bool))
				{{- else }}
					e.{{ $member.Name }} = v.PrimitiveHinted().({{ memberType $member.Type }})
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
		"structName": func(name string) string {
			ps := strings.Split(name, "/")
			ps = strings.Split(ps[len(ps)-1], ".")
			if len(ps) == 1 {
				return ucFirst(ps[0])
			}
			if strings.ToLower(ps[len(ps)-2]) == strings.ToLower(doc.PackageAlias) {
				return ucFirst(ps[len(ps)-1])
			}
			return ucFirst(ps[len(ps)-2]) + ucFirst(ps[len(ps)-1])
		},
		"memberType": func(name string) string {
			for alias, pkg := range originalImports {
				name = strings.Replace(name, pkg, alias, 1)
			}
			ps := strings.Split(name, "/")
			return strings.TrimPrefix(ps[len(ps)-1], doc.PackageAlias+".")
		},
		"memberTag": func(tag, hint string, isRepeated bool) string {
			if isRepeated {
				return tag + ":a" + hint
			}
			return tag + ":" + hint
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
			o.Name = s.Name + "." + o.Name
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
		// res = strings.ReplaceAll(res, "object.Hash", "Hash")
		// res = strings.ReplaceAll(res, "object.Schema", "Schema")
		// res = strings.ReplaceAll(res, "object.Signature", "Signature")
		// res = strings.ReplaceAll(res, "object.Policy", "Policy")
		// res = strings.ReplaceAll(res, "object.FromMap", "FromMap")
	}

	return []byte(res), nil
}

// lastSegment returns the last part of a namespace,
// ie lastSegment(nimona.io/stream) returns stream
func lastSegment(s string) string {
	ps := strings.Split(s, "/")
	return ps[len(ps)-1]
}
