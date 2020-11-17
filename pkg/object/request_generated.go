// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package object

type (
	Request struct {
		Metadata              Metadata `nimona:"metadata:m,omitempty"`
		RequestID             string   `nimona:"requestID:s,omitempty"`
		ObjectHash            Hash     `nimona:"objectHash:r,omitempty"`
		ExcludedNestedObjects bool     `nimona:"excludedNestedObjects:b,omitempty"`
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
	o, err := Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *Request) FromObject(o *Object) error {
	return Decode(o, e)
}

func (e *Response) Type() string {
	return "nimona.io/Response"
}

func (e Response) ToObject() *Object {
	o, err := Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *Response) FromObject(o *Object) error {
	return Decode(o, e)
}
