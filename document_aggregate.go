package nimona

import (
	"encoding/json"
	"fmt"
	"reflect"

	"nimona.io/internal/ckv"
)

type AggregateStore struct {
	Store ckv.Store
}

func (s *AggregateStore) Apply(doc *Document) error {
	doc = doc.Copy()

	if doc.Type() != "core/stream/patch" {
		body, err := doc.MarshalJSON()
		if err != nil {
			return err
		}
		err = s.Store.Put(
			fmt.Sprintf(
				"%s/",
				NewDocumentID(doc).String(),
			),
			body,
		)
		if err != nil {
			return fmt.Errorf("error storing root doc: %w", err)
		}
		return nil
	}

	patch := &DocumentPatch{}
	err := patch.FromDocument(doc)
	if err != nil {
		return fmt.Errorf("error unmarshaling patch: %w", err)
	}

	for _, operation := range patch.Operations {
		switch operation.Op {
		case "replace":
			// TODO: implement
		case "append":
			body, err := json.Marshal(operation.Value)
			if err != nil {
				return fmt.Errorf("error marshaling op value: %w", err)
			}
			err = s.Store.Put(
				fmt.Sprintf(
					"%s/%s/%s/",
					patch.Metadata.Root.String(),
					operation.Path,
					operation.Key,
				),
				body,
			)
			if err != nil {
				return fmt.Errorf("error storing op value: %w", err)
			}
		default:
			return fmt.Errorf("unsupported operation: %s", operation.Op)
		}
	}
	return nil
}

func (s *AggregateStore) GetAggregateNested(
	rootHash DocumentID,
	path string,
	target any, // *[]DocumentMapper,
) error {
	key := fmt.Sprintf("%s/%s/", rootHash.String(), path)
	bodies, err := s.Store.Get(key)
	if err != nil {
		return fmt.Errorf("error getting aggregate: %w", err)
	}

	// check if target is a pointer to a slice
	res := reflect.ValueOf(target).Elem()
	if res.Kind() != reflect.Slice {
		return fmt.Errorf("target is not a slice")
	}

	// check if target is a pointer to a slice of pointers
	if res.Type().Elem().Kind() != reflect.Ptr {
		return fmt.Errorf("target is not a slice of pointers")
	}

	// get the type of the slice elements
	typ := reflect.TypeOf(target).Elem().Elem().Elem()

	for _, body := range bodies {
		doc := &Document{}
		err := doc.UnmarshalJSON(body)
		if err != nil {
			return fmt.Errorf("error unmarshaling aggregate: %w", err)
		}
		val := reflect.New(typ).Interface().(DocumentMapper)
		err = (val).FromDocument(doc)
		if err != nil {
			return fmt.Errorf("error unmarshaling aggregate: %w", err)
		}
		res.Set(reflect.Append(res, reflect.ValueOf(val)))
	}

	return nil
}
