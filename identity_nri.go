package nimona

import (
	"fmt"
	"strings"
)

func ParseIdentityNRI(nri string) (*Identity, error) {
	t := string(ShorthandIdentity)
	if !strings.HasPrefix(nri, t) {
		return nil, fmt.Errorf("invalid identity nri")
	}

	nri = strings.TrimPrefix(nri, t)
	dh, err := ParseDocumentHash(nri)
	if err != nil {
		return nil, fmt.Errorf("unable to parse identity nri: %w", err)
	}
	return &Identity{
		KeyGraph: dh,
	}, nil
}
