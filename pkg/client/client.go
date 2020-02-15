package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type (
	Client struct {
		baseURL    *url.URL
		httpClient *http.Client
	}
	InfoResponse struct {
		Signature   *crypto.Signature `json:"_signature:o"`
		Hash        string            `json:"_hash"`
		Fingerprint string            `json:"_fingerprint"`
		Addresses   []string          `json:"addresses:as"`
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
	if resBody == nil {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(b))
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(&resBody)
}

func (c *Client) Info() (*InfoResponse, error) {
	resp := &InfoResponse{}
	if err := c.do("GET", "/api/v1/local", nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) PostObject(o object.Object) error {
	return c.do("POST", "/api/v1/objects", o.ToMap(), nil)
}
