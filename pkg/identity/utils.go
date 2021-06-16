package identity

import (
	"fmt"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/resolver"
	object "nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
)

// TODO Implement
func Lookup(
	ctx context.Context,
	idKey crypto.PublicKey,
	res resolver.Resolver,
	man objectmanager.ObjectManager,
) (*Profile, error) {
	// TODO check key usage is identity
	streamRoot := &ProfileStreamRoot{
		Metadata: object.Metadata{
			Owner: idKey,
		},
	}
	streamRootObj, err := streamRoot.MarshalObject()
	if err != nil {
		return nil, err
	}
	streamRootHash := streamRootObj.Hash()

	recipients, err := res.Lookup(
		ctx,
		resolver.LookupByPeerKey(
			idKey,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error looking up stream providers, %w", err)
	}

	_, err = man.RequestStream(
		ctx,
		streamRootHash,
		recipients...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to request stream, %w", err)
	}

	return nil, fmt.Errorf("not implemented")
}
