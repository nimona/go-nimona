package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	IdentityAlias struct {
		_        string `nimona:"$type,type=core/identity.alias"`
		Hostname string `nimona:"hostname,omitempty"`
		Path     string `nimona:"path,omitempty"`
	}
	IdentityInfo struct {
		_             Metadata      `nimona:"$metadata,omitempty,type=core/identity.info"`
		Alias         IdentityAlias `nimona:"alias,omitempty"`
		KeygraphID    KeygraphID    `nimona:"keygraph,omitempty"`
		PeerAddresses []PeerAddr    `nimona:"peerAddresses"`
	}
)

func (i IdentityAlias) String() string {
	r := string(ShorthandIdentityAlias) + i.Hostname
	if i.Path != "" {
		r += "/" + i.Path
	}
	return r
}

func ParseIdentityAlias(alias string) (*IdentityAlias, error) {
	t := string(ShorthandIdentityAlias)
	if !strings.HasPrefix(alias, t) {
		return nil, fmt.Errorf("invalid resource id")
	}

	alias = strings.TrimPrefix(alias, t)
	hostname, path, _ := strings.Cut(alias, "/")
	return &IdentityAlias{
		Hostname: hostname,
		Path:     path,
	}, nil
}

func (i IdentityAlias) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *IdentityAlias) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if idString, ok := value.(string); ok {
		id, err := ParseIdentityAlias(idString)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityAlias: %w", err)
		}
		i.Hostname = id.Hostname
		i.Path = id.Path
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityAlias")
}
