package feed

import (
	"strings"

	"github.com/elliotchance/orderedmap"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func GetFeedHashes(
	objectReader object.Reader,
) ([]chore.Hash, error) {
	objects := orderedmap.NewOrderedMap()
	for {
		obj, err := objectReader.Read()
		if err == object.ErrReaderDone {
			break
		}
		if err != nil {
			return nil, err
		}
		switch obj.Type {
		case AddedType:
			event := &Added{}
			// TODO should this error?
			if err := object.Unmarshal(obj, event); err != nil {
				return nil, err
			}
			for _, hash := range event.ObjectHash {
				objects.Set(hash, true)
			}
		case RemovedType:
			event := &Removed{}
			// TODO should this error?
			if err := object.Unmarshal(obj, event); err != nil {
				return nil, err
			}
			for _, hash := range event.ObjectHash {
				objects.Set(hash, false)
			}
		}
	}
	hashes := []chore.Hash{}
	for el := objects.Front(); el != nil; el = el.Next() {
		if !el.Value.(bool) {
			continue
		}
		hashes = append(hashes, el.Key.(chore.Hash))
	}
	return hashes, nil
}

func GetFeedHypotheticalRoot(
	owner crypto.PublicKey,
	objectType string,
) *FeedStreamRoot {
	r := &FeedStreamRoot{
		ObjectType: getTypeForFeed(objectType),
		Metadata: object.Metadata{
			Owner: owner.DID(),
		},
	}
	return r
}

func GetFeedHypotheticalRootHash(
	owner crypto.PublicKey,
	objectType string,
) chore.Hash {
	return object.MustMarshal(
		GetFeedHypotheticalRoot(
			owner,
			objectType,
		),
	).Hash()
}

func getTypeForFeed(objectType string) string {
	pt := object.ParseType(objectType)
	return strings.TrimLeft(pt.Namespace+"/"+pt.Object, "/")
}
