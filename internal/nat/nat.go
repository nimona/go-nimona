package nat

import (
	"fmt"

	"gitlab.com/NebulousLabs/go-upnp"

	"nimona.io/pkg/eventbus"
)

func MapExternalPort(port int) (func(), error) {
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		return nil, err
	}

	// clear existing mappings
	d.Clear(uint16(port)) // nolint: errcheck

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		return nil, err
	}

	// add port mapping
	err = d.Forward(uint16(port), "nimona daemon")
	if err != nil {
		return nil, err
	}

	eventbus.DefaultEventbus.Publish(
		eventbus.NetworkAddressAdded{
			Address: fmt.Sprintf(
				"tcps:%s:%d",
				ip,
				port,
			),
		},
	)

	return func() {
		// clear mappings
		d.Clear(uint16(port)) // nolint: errcheck
	}, nil
}
