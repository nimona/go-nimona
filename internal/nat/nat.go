package nat

import (
	"fmt"

	"gitlab.com/NebulousLabs/go-upnp"
)

func MapExternalPort(port int) (string, func(), error) {
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
