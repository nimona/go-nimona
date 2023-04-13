package nimona

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/copystructure"

	"nimona.io/tilde"
)

type Document struct {
	Metadata Metadata
	data     tilde.Map
	Schema   *tilde.Schema
}

type DocumentMapper interface {
	FromDocument(doc *Document) error
	Document() *Document
}

func NewDocument(m tilde.Map) *Document {
	doc := &Document{
		Metadata: Metadata{},
		data:     m,
	}
	// pull metadata out of map
	metaMap := m.Fluent().Get("$metadata").Map()
	if metaMap != nil {
		err := doc.Metadata.FromMap(metaMap)
		if err != nil {
			panic(fmt.Errorf("error parsing metadata: %w", err))
		}
	}

	return doc
}

type DocumentValuer interface {
	DocumentValue(v any) any
}

func (doc *Document) Set(path string, value tilde.Value) error {
	if strings.Contains(path, "$metadata") {
		return fmt.Errorf("cannot set $metadata")
	}

	return doc.data.Set(path, value)
}

func (doc *Document) Get(path string) (tilde.Value, error) {
	if strings.Contains(path, "$metadata") {
		return nil, fmt.Errorf("cannot get $metadata")
	}

	return doc.data.Get(path)
}

func (doc *Document) Map() tilde.Map {
	if doc == nil {
		return nil
	}
	docMap := tilde.Copy(doc.data)
	metaMap := doc.Metadata.Map()
	if len(metaMap) > 0 {
		err := docMap.Set("$metadata", metaMap)
		if err != nil {
			// TODO: handle this better
			panic(fmt.Errorf("error setting metadata: %w", err))
		}
	}
	return docMap
}

func (doc *Document) Type() string {
	if doc.data == nil {
		return ""
	}
	vi, err := doc.data.Get("$type")
	if err != nil {
		return ""
	}
	v, ok := vi.(tilde.String)
	if !ok {
		return ""
	}
	return string(v)
}

func (doc *Document) Context() string {
	if doc.data == nil {
		return ""
	}
	vi, err := doc.data.Get("$ctx")
	if err != nil {
		return ""
	}
	v, ok := vi.(tilde.String)
	if !ok {
		return ""
	}
	return string(v)
}

func (doc *Document) Copy() *Document {
	newMeta, err := copystructure.Copy(doc.Metadata)
	if err != nil {
		panic(fmt.Errorf("error copying metadata: %w", err))
	}
	newDoc := &Document{
		Metadata: newMeta.(Metadata),
		data:     tilde.Copy(doc.data),
	}
	return newDoc
}

func (doc *Document) Document() *Document {
	return doc
}

func (doc *Document) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(doc.Map())
	if err != nil {
		return nil, fmt.Errorf("error marshaling into json: %w", err)
	}
	return b, nil
}

func (doc *Document) UnmarshalJSON(b []byte) error {
	if doc == nil {
		*doc = Document{}
	}
	mm := &tilde.Map{}
	err := json.Unmarshal(b, mm)
	if err != nil {
		return fmt.Errorf("error unmarshaling from json: %w", err)
	}
	doc.data = *mm
	return nil
}

func DumpDocumentBytes(b []byte) {
	fmt.Printf("%x\n", b)
}

func DumpDocument(doc *Document) {
	yb, err := json.MarshalIndent(doc.Map(), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(yb))
}
