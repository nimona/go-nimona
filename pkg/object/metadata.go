package object

import (
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

type (
	// Metadata for object
	// TODO: add shape
	// TODO: add authors, contributors, license, copyright
	// TODO: add version
	Metadata struct {
		Owner     peer.ID      `nimona:"owner:s"`
		Parents   Parents      `nimona:"parents:m"`
		Policies  Policies     `nimona:"policies:am"`
		Root      tilde.Digest `nimona:"root:r"`
		Sequence  uint64       `nimona:"sequence:u,omitzero"`
		Signature Signature    `nimona:"_signature:m"`
		Timestamp string       `nimona:"timestamp:s"`
	}
)
