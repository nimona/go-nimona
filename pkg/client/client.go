package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"nimona.io/pkg/peer"

	"nimona.io/pkg/object"
)

type (
	Client struct {
		baseURL    *url.URL
		httpClient *http.Client
	}
)

func New(baseURL string) (*Client, error) {
	bURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:    bURL,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) do(
	method string,
	path string,
	reqBody interface{},
	resBody interface{},
) error {
	rel := &url.URL{
		Path: path,
	}
	u := c.baseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if reqBody != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(reqBody)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent:
	default:
		return errors.New("unexpected status " + resp.Status)
	}
	defer resp.Body.Close() // nolint: errcheck
	return json.NewDecoder(resp.Body).Decode(&resBody)
}

func (c *Client) Info() (*peer.Peer, error) {
	resp := map[string]interface{}{}
	if err := c.do("GET", "/api/v1/local", nil, &resp); err != nil {
		return nil, err
	}
	p := &peer.Peer{}
	p.FromObject(object.FromMap(resp)) // nolint: errcheck
	return p, nil
}

func (c *Client) PostObject(o object.Object) error {
	return c.do("POST", "/api/v1/objects", o.ToMap(), nil)
}
