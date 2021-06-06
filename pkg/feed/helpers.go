package feed

import (
	"strings"

	"github.com/elliotchance/orderedmap"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	value "nimona.io/pkg/object/value"
)

var (
	feedObjectAddedType   = new(Added).Type()
	feedObjectRemovedType = new(Removed).Type()
)

func GetFeedCIDs(
	objectReader object.Reader,
) ([]value.CID, error) {
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
		case feedObjectAddedType:
			event := &Added{}
			// TODO should this error?
			if err := event.UnmarshalObject(obj); err != nil {
				return nil, err
			}
			for _, cid := range event.ObjectCID {
				objects.Set(cid, true)
			}
		case feedObjectRemovedType:
			event := &Removed{}
			// TODO should this error?
			if err := event.UnmarshalObject(obj); err != nil {
				return nil, err
			}
			for _, cid := range event.ObjectCID {
				objects.Set(cid, false)
			}
		}
	}
	cids := []value.CID{}
	for el := objects.Front(); el != nil; el = el.Next() {
		if !el.Value.(bool) {
			continue
		}
		cids = append(cids, el.Key.(value.CID))
	}
	return cids, nil
}

func GetFeedHypotheticalRoot(
	owner crypto.PublicKey,
	objectType string,
) *FeedStreamRoot {
	r := &FeedStreamRoot{
		ObjectType: getTypeForFeed(objectType),
		Metadata: object.Metadata{
			Owner: owner,
		},
	}
	return r
}

func GetFeedHypotheticalRootCID(
	owner crypto.PublicKey,
	objectType string,
) value.CID {
	return object.MustMarshal(
		GetFeedHypotheticalRoot(
			owner,
			objectType,
		),
	).CID()
}

func getTypeForFeed(objectType string) string {
	pt := object.ParseType(objectType)
	return strings.TrimLeft(pt.Namespace+"/"+pt.Object, "/")
}
