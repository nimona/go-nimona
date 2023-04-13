package nimona

import (
	"testing"
)

func Test_Profile(t *testing.T) {
	p := &Profile{
		Metadata: Metadata{
			Owner: NewTestKeygraphID(t),
		},
		DisplayName: "test",
	}

	d := p.Document()

	DumpDocument(d)
}
