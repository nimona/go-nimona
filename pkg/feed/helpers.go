package feed

import (
	"strings"

	"github.com/elliotchance/orderedmap"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

var (
	feedObjectAddedType   = Added{}.GetType()
	feedObjectRemovedType = Removed{}.GetType()
)

func GetFeedHashes(
	objectReader object.Reader,
) ([]object.Hash, error) {
	objects := orderedmap.NewOrderedMap()
	for {
		obj, err := objectReader.Read()
		if err == object.ErrReaderDone {
			break
		}
		if err != nil {
			return nil, err
		}
		switch obj.GetType() {
		case feedObjectAddedType:
			event := &Added{}
			// TODO should this error?
			if err := event.FromObject(*obj); err != nil {
				return nil, err
			}
			for _, hash := range event.ObjectHash {
				objects.Set(hash, true)
			}
		case feedObjectRemovedType:
			event := &Removed{}
			// TODO should this error?
			if err := event.FromObject(*obj); err != nil {
				return nil, err
			}
			for _, hash := range event.ObjectHash {
				objects.Set(hash, false)
			}
		}
	}
	hashes := []object.Hash{}
	for el := objects.Front(); el != nil; el = el.Next() {
		if !el.Value.(bool) {
			continue
		}
		hashes = append(hashes, el.Key.(object.Hash))
	}
	return hashes, nil
}

func GetFeedHypotheticalRoot(
	owners []crypto.PublicKey,
	objectType string,
) FeedStreamRoot {
	r := FeedStreamRoot{
		Type:   getTypeForFeed(objectType),
		Owners: owners,
	}
	return r
}

func GetFeedHypotheticalRootHash(
	owners []crypto.PublicKey,
	objectType string,
) object.Hash {
	return GetFeedHypotheticalRoot(owners, objectType).ToObject().Hash()
}

func getTypeForFeed(objectType string) string {
	pt := object.ParseType(objectType)
	return strings.TrimLeft(pt.Namespace+"/"+pt.Object, "/")
}
