package identity

import (
	"fmt"

	"nimona.io/pkg/context"
	"nimona.io/pkg/did"
	"nimona.io/pkg/hyperspace/resolver"
	object "nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
)

// TODO Implement
func Lookup(
	ctx context.Context,
	id did.DID,
	res resolver.Resolver,
	man objectmanager.ObjectManager,
) (*Profile, error) {
	// TODO check key usage is identity
	streamRoot := &ProfileStreamRoot{
		Metadata: object.Metadata{
			Owner: id,
		},
	}
	streamRootObj, err := object.Marshal(streamRoot)
	if err != nil {
		return nil, err
	}
	streamRootHash := streamRootObj.Hash()

	_, err = man.RequestStream(
		ctx,
		streamRootHash,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to request stream, %w", err)
	}

	return nil, fmt.Errorf("not implemented")
}
