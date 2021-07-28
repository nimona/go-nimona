package object

import (
	"nimona.io/pkg/chore"
	"nimona.io/pkg/did"
)

type (
	// Metadata for object
	// TODO: rename stream to root
	// TODO: add shape, sequence
	// TODO: add authors, contributors, license, copyright
	// TODO: add version
	// TODO: consider renaming datetime to timestamp
	Metadata struct {
		Owner     did.DID    `nimona:"owner:s"`
		Datetime  string     `nimona:"datetime:s"`
		Parents   Parents    `nimona:"parents:m"`
		Policies  Policies   `nimona:"policies:am"`
		Stream    chore.Hash `nimona:"stream:r"`
		Signature Signature  `nimona:"_signature:m"`
	}
)
