package main

import (
	"fmt"

	cbg "github.com/whyrusleeping/cbor-gen"

	"nimona.io"
)

type mapping struct {
	file  string
	pkg   string
	types []any
}

var mappings = []mapping{{
	file: "node_addr_gen.go",
	pkg:  "nimona",
	types: []any{
		nimona.NodeAddr{},
	},
}, {
	file: "node_info_gen.go",
	pkg:  "nimona",
	types: []any{
		nimona.NodeInfo{},
	},
}, {
	file: "message_wrapper_gen.go",
	pkg:  "nimona",
	types: []any{
		nimona.MessageWrapper{},
	},
}, {
	file: "message_ping_gen.go",
	pkg:  "nimona",
	types: []any{
		nimona.Ping{},
		nimona.Pong{},
	},
}, {
	file: "rpc_peer_capabilities_gen.go",
	pkg:  "nimona",
	types: []any{
		nimona.PeerCapabilitiesRequest{},
		nimona.PeerCapabilitiesResponse{},
	},
}}

func main() {
	for _, m := range mappings {
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
