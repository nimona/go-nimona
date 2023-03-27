package nimona

import (
	"testing"
)

func Test_Profile(t *testing.T) {
	p := &Profile{
		Metadata: Metadata{
			Owner: NewTestIdentity(t),
		},
		DisplayName: "test",
	}

	d := p.Document()

	DumpDocument(d)
}
