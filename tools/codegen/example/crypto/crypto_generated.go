// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package crypto

import (
	"nimona.io/pkg/immutable"
	object "nimona.io/pkg/object"
)

type (
	Hash struct {
		Header   object.Header
		HashType string
		Digest   []byte
	}
	HeaderSignature struct {
		Header    object.Header
		PublicKey *PublicKey
		Algorithm string
		R         []byte
		S         []byte
	}
	PrivateKey struct {
		Header    object.Header
		PublicKey *PublicKey
		KeyType   string
		Algorithm string
		Curve     string
		X         []byte
		Y         []byte
		D         []byte
	}
	PublicKey struct {
		Header    object.Header
		KeyType   string
		Algorithm string
		Curve     string
		X         []byte
		Y         []byte
	}
)

func (e Hash) GetType() string {
	return "example/crypto.Hash"
}

func (e Hash) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "hashType",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "digest",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e Hash) ToObject() object.Object {
	d := map[string]interface{}{}
	if e.HashType != "" {
		d["hashType:s"] = e.HashType
	}
	if len(e.Digest) != 0 {
		d["digest:d"] = e.Digest
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	o := object.Object{
		Header: e.Header,
		Data:   immutable.AnyToValue(":o", d).(immutable.Map),
	}
	o.SetType("example/crypto.Hash")
	return o
}

func (e *Hash) FromObject(o object.Object) error {
	e.Header = o.Header
	if v := o.Data.Value("hashType:s"); v != nil {
		e.HashType = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("digest:d"); v != nil {
		e.Digest = []byte(v.PrimitiveHinted().([]byte))
	}
	return nil
}

func (e HeaderSignature) GetType() string {
	return "example/object.Header.Signature"
}

func (e HeaderSignature) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "publicKey",
				Type:       "PublicKey",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "algorithm",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "r",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "s",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e HeaderSignature) ToObject() object.Object {
	d := map[string]interface{}{}
	if e.PublicKey != nil {
		d["publicKey:o"] = e.PublicKey.ToObject().ToMap()
	}
	if e.Algorithm != "" {
		d["algorithm:s"] = e.Algorithm
	}
	if len(e.R) != 0 {
		d["r:d"] = e.R
	}
	if len(e.S) != 0 {
		d["s:d"] = e.S
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	o := object.Object{
		Header: e.Header,
		Data:   immutable.AnyToValue(":o", d).(immutable.Map),
	}
	o.SetType("example/object.Header.Signature")
	return o
}

func (e *HeaderSignature) FromObject(o object.Object) error {
	e.Header = o.Header
	if v := o.Data.Value("publicKey:o"); v != nil {
		es := &PublicKey{}
		eo := object.FromMap(v.PrimitiveHinted().(map[string]interface{}))
		es.FromObject(eo)
		e.PublicKey = es
	}
	if v := o.Data.Value("algorithm:s"); v != nil {
		e.Algorithm = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("r:d"); v != nil {
		e.R = []byte(v.PrimitiveHinted().([]byte))
	}
	if v := o.Data.Value("s:d"); v != nil {
		e.S = []byte(v.PrimitiveHinted().([]byte))
	}
	return nil
}

func (e PrivateKey) GetType() string {
	return "example/crypto.PrivateKey"
}

func (e PrivateKey) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "publicKey",
				Type:       "PublicKey",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "keyType",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "algorithm",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "curve",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "x",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "y",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "d",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e PrivateKey) ToObject() object.Object {
	d := map[string]interface{}{}
	if e.PublicKey != nil {
		d["publicKey:o"] = e.PublicKey.ToObject().ToMap()
	}
	if e.KeyType != "" {
		d["keyType:s"] = e.KeyType
	}
	if e.Algorithm != "" {
		d["algorithm:s"] = e.Algorithm
	}
	if e.Curve != "" {
		d["curve:s"] = e.Curve
	}
	if len(e.X) != 0 {
		d["x:d"] = e.X
	}
	if len(e.Y) != 0 {
		d["y:d"] = e.Y
	}
	if len(e.D) != 0 {
		d["d:d"] = e.D
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	o := object.Object{
		Header: e.Header,
		Data:   immutable.AnyToValue(":o", d).(immutable.Map),
	}
	o.SetType("example/crypto.PrivateKey")
	return o
}

func (e *PrivateKey) FromObject(o object.Object) error {
	e.Header = o.Header
	if v := o.Data.Value("publicKey:o"); v != nil {
		es := &PublicKey{}
		eo := object.FromMap(v.PrimitiveHinted().(map[string]interface{}))
		es.FromObject(eo)
		e.PublicKey = es
	}
	if v := o.Data.Value("keyType:s"); v != nil {
		e.KeyType = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("algorithm:s"); v != nil {
		e.Algorithm = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("curve:s"); v != nil {
		e.Curve = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("x:d"); v != nil {
		e.X = []byte(v.PrimitiveHinted().([]byte))
	}
	if v := o.Data.Value("y:d"); v != nil {
		e.Y = []byte(v.PrimitiveHinted().([]byte))
	}
	if v := o.Data.Value("d:d"); v != nil {
		e.D = []byte(v.PrimitiveHinted().([]byte))
	}
	return nil
}

func (e PublicKey) GetType() string {
	return "example/crypto.PublicKey"
}

func (e PublicKey) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "keyType",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "algorithm",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "curve",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "x",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "y",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e PublicKey) ToObject() object.Object {
	d := map[string]interface{}{}
	if e.KeyType != "" {
		d["keyType:s"] = e.KeyType
	}
	if e.Algorithm != "" {
		d["algorithm:s"] = e.Algorithm
	}
	if e.Curve != "" {
		d["curve:s"] = e.Curve
	}
	if len(e.X) != 0 {
		d["x:d"] = e.X
	}
	if len(e.Y) != 0 {
		d["y:d"] = e.Y
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	o := object.Object{
		Header: e.Header,
		Data:   immutable.AnyToValue(":o", d).(immutable.Map),
	}
	o.SetType("example/crypto.PublicKey")
	return o
}

func (e *PublicKey) FromObject(o object.Object) error {
	e.Header = o.Header
	if v := o.Data.Value("keyType:s"); v != nil {
		e.KeyType = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("algorithm:s"); v != nil {
		e.Algorithm = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("curve:s"); v != nil {
		e.Curve = string(v.PrimitiveHinted().(string))
	}
	if v := o.Data.Value("x:d"); v != nil {
		e.X = []byte(v.PrimitiveHinted().([]byte))
	}
	if v := o.Data.Value("y:d"); v != nil {
		e.Y = []byte(v.PrimitiveHinted().([]byte))
	}
	return nil
}
