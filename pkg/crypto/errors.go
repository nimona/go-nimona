package crypto

import "nimona.io/pkg/errors"

const (
	ErrOnlyEd25519KeysSupported = errors.Error("only ed25519 keys supported")
)
