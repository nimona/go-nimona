// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package object

type (
	Request struct {
		Metadata   Metadata `nimona:"metadata:m,omitempty"`
		RequestID  string   `nimona:"requestID:s,omitempty"`
		ObjectHash Hash     `nimona:"objectHash:r,omitempty"`
	}
	Response struct {
		Metadata  Metadata `nimona:"metadata:m,omitempty"`
		RequestID string   `nimona:"requestID:s,omitempty"`
		Object    *Object  `nimona:"object:o,omitempty"`
	}
)

func (e *Request) Type() string {
	return "nimona.io/Request"
}

func (e Request) ToObject() *Object {
	r := &Object{
		Type:     "nimona.io/Request",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["requestID:s"] = e.RequestID
	r.Data["objectHash:r"] = e.ObjectHash
	return r
}

func (e Request) ToObjectMap() map[string]interface{} {
	d := map[string]interface{}{}
	d["requestID:s"] = e.RequestID
	d["objectHash:r"] = e.ObjectHash
	r := map[string]interface{}{
		"type:s":     "nimona.io/Request",
		"metadata:m": MetadataToMap(&e.Metadata),
		"data:m":     d,
	}
	return r
}

func (e *Request) FromObject(o *Object) error {
	return Decode(o, e)
}

func (e *Response) Type() string {
	return "nimona.io/Response"
}

func (e Response) ToObject() *Object {
	r := &Object{
		Type:     "nimona.io/Response",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["requestID:s"] = e.RequestID
	if e.Object != nil {
		r.Data["object:o"] = e.Object
	}
	return r
}

func (e Response) ToObjectMap() map[string]interface{} {
	d := map[string]interface{}{}
	d["requestID:s"] = e.RequestID
	if e.Object != nil {
		d["object:o"] = e.Object
	}
	r := map[string]interface{}{
		"type:s":     "nimona.io/Response",
		"metadata:m": MetadataToMap(&e.Metadata),
		"data:m":     d,
	}
	return r
}

func (e *Response) FromObject(o *Object) error {
	return Decode(o, e)
}
