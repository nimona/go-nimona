package main

import (
	"fmt"

	"github.com/ugorji/go/codec"

	"github.com/nimona/go-nimona/net"
)

type NestedStruct struct {
	Bool   bool
	String string
	Int    int
}

type Dummy struct {
	Bool         bool
	String       string
	Int          int
	NestedStruct []*NestedStruct
}

func main() {
	net.GlobalRegistry.Register("dummy", Dummy{})

	a := &net.Message{
		Headers: net.Headers{
			ContentType: "dummy",
		},
		Payload: &Dummy{
			Bool:   true,
			String: "string",
			Int:    1,
			NestedStruct: []*NestedStruct{
				&NestedStruct{
					Bool:   true,
					String: "string",
					Int:    1,
				},
				&NestedStruct{
					Bool:   true,
					String: "string",
					Int:    2,
				},
			},
		},
	}
	bytes, _ := net.Marshal(a)

	// b
	b := &net.Message{
		Payload: &Dummy{},
	}
	dec := codec.NewDecoderBytes(bytes, &codec.CborHandle{})
	err := dec.Decode(&b)
	if err != nil {
		panic(err)
	}
	fmt.Println(b.Payload.(*Dummy).String)
	fmt.Println(b.Payload.(*Dummy).NestedStruct[1].Int)

	// c
	c := &net.Message{}
	dec = codec.NewDecoderBytes(bytes, &codec.CborHandle{})
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.Payload.(*Dummy).String)
	fmt.Println(b.Payload.(*Dummy).NestedStruct[1].Int)

	// d
	d := &Dummy{}
	err = c.DecodePayload(d)
	if err != nil {
		panic(err)
	}

	fmt.Println(d.String)
	fmt.Println(d.NestedStruct[1].Int)

}
