package main

import (
	"fmt"

	cbg "github.com/whyrusleeping/cbor-gen"

	"nimona.io"
)

type mapping struct {
	file  string
	types []any
	pkg   string
}

var mappings = []mapping{{
	file: "session_request_gen.go",
	types: []any{
		nimona.Request{},
	},
}, {
	file: "session_response_gen.go",
	types: []any{
		nimona.Response{},
	},
}, {
	file: "peer_addr_gen.go",
	types: []any{
		nimona.PeerAddr{},
	},
}, {
	file: "peer_info_gen.go",
	types: []any{
		nimona.PeerInfo{},
	},
}, {
	file: "message_wrapper_gen.go",
	types: []any{
		nimona.MessageWrapper{},
	},
}, {
	file: "message_ping_gen.go",
	types: []any{
		nimona.Ping{},
		nimona.Pong{},
	},
}, {
	file: "handler_peer_gen.go",
	types: []any{
		nimona.PeerCapabilitiesRequest{},
		nimona.PeerCapabilitiesResponse{},
	},
}, {
	file: "handler_network_gen.go",
	types: []any{
		nimona.NetworkInfoRequest{},
		nimona.NetworkInfo{},
	},
}, {
	file: "identifier_network_gen.go",
	types: []any{
		nimona.NetworkID{},
	},
}, {
	file: "message_metadata_gen.go",
	types: []any{
		nimona.Signature{},
		nimona.Metadata{},
	},
}, {
	file: "keystream_gen.go",
	types: []any{
		nimona.KeyStreamPermissions{},
		nimona.KeyStreamDelegatorSeal{},
		nimona.KeyStream{},
	},
}}

func main() {
	for _, m := range mappings {
		if m.pkg == "" {
			m.pkg = "nimona"
		}
		err := cbg.WriteMapEncodersToFile(
			m.file,
			m.pkg,
			m.types...,
		)
		if err != nil {
			panic(
				fmt.Sprintf("error writing %s, err: %s", m.file, err),
			)
		}
	}
}
