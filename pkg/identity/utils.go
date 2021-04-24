package identity

import (
	"fmt"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/resolver"
	object "nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
)

func Lookup(
	ctx context.Context,
	idKey crypto.PublicKey,
	res resolver.Resolver,
	man objectmanager.ObjectManager,
) (*Profile, error) {
	// TODO check key usage is identity
	streamRootCID := ProfileStreamRoot{
		Metadata: object.Metadata{
			Owner: idKey,
		},
	}.ToObject().CID()

	recipients, err := res.Lookup(
		ctx,
		resolver.LookupByPeerKey(
			idKey,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error looking up stream providers, %w", err)
	}

	man.RequestStream(
		ctx,
		streamRootCID,
		recipients...,
	)

	return nil, fmt.Errorf("not implemented")
}
