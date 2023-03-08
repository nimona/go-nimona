package nimona

import (
	"fmt"
	"io"
	"net/http"
)

const (
	resolverWellKnownProvider = "https://%s/.well-known/nimona.json"
	resolverWellKnownUser     = "https://%s/.well-known/nimona-%s.json"
)

type ResolverHTTP struct{}

func (r *ResolverHTTP) ResolveIdentityAlias(alias *IdentityAlias) (*NetworkInfo, error) {
	url := ""
	if alias.Handle != "" {
		url = fmt.Sprintf(resolverWellKnownUser, alias.Network.Hostname, alias.Handle)
	} else {
		url = fmt.Sprintf(resolverWellKnownProvider, alias.Network.Hostname)
	}

	// make http request to url
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
	id := &NetworkInfo{}
	err = id.FromDocument(idDoc)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve identity: %w", err)
	}

	return id, nil
}
