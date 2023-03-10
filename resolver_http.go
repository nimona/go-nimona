package nimona

import (
	"fmt"
	"io"
	"net/http"
)

const (
	resolverWellKnownAlias = "https://%s/.well-known/nimona.json"
)

type ResolverHTTP struct{}

func (r *ResolverHTTP) ResolveIdentityAlias(alias IdentityAlias) (*IdentityInfo, error) {
	// make http request to url
	url := fmt.Sprintf(resolverWellKnownAlias, alias.Hostname)
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unable to resolve identity: %s", res.Status)
	}

	// get body
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}

	// parse response
	idDoc := &Document{}
	err = idDoc.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}

	// decode document
	id := &IdentityInfo{}
	err = id.FromDocument(idDoc)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}

	return id, nil
}
