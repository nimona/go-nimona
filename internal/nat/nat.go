package nat

import (
	"fmt"
	"os"

	"gitlab.com/NebulousLabs/go-upnp"

	"nimona.io/pkg/errors"
)

func MapExternalPort(port int) (address string, removeMap func(), err error) {
	if os.Getenv("NIMONA_SKIP_UPNP") != "" {
		return "", nil, errors.Error("skipped")
	}
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		return "", nil, err
	}

	// clear existing mappings
	d.Clear(uint16(port)) // nolint: errcheck

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		return "", nil, err
	}

	// add port mapping
	err = d.Forward(uint16(port), "nimona daemon")
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf(
			"tcps:%s:%d",
			ip,
			port,
		), func() {
			// clear mappings
			d.Clear(uint16(port)) // nolint: errcheck
		}, nil
}
