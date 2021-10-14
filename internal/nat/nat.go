package nat

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/libp2p/go-nat"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
)

func MapExternalPort(
	ctx context.Context,
	localPort int,
) (address string, rm func(), err error) {
	if skip, _ := strconv.ParseBool(os.Getenv("NIMONA_UPNP_DISABLE")); skip {
		return "", nil, errors.Error("skipped")
	}
	// connect to router
	gw, err := nat.DiscoverGateway(ctx)
	if err != nil {
		return "", nil, err
	}

	// discover external IP
	externalIP, err := gw.GetExternalAddress()
	if err != nil {
		return "", nil, err
	}

	// TODO make service name configurable, maybe using the `version` pkg?
	serviceName := "nimona"

	// add port mapping
	externalPort, err := gw.AddPortMapping(
		"tcp",
		localPort,
		serviceName,
		time.Minute,
	)
	if err != nil {
		return "", nil, err
	}

	go func() {
		for {
			<-time.After(30 * time.Second)
			if _, err := gw.AddPortMapping(
				"tcp",
				localPort,
				serviceName,
				time.Minute,
			); err != nil {
				// TODO should we exist after X failed attempts?
				log.DefaultLogger.Error(
					"error adding port mapping",
					log.String("method", "MapExternalPort"),
					log.Error(err),
				)
			}
		}
	}()

	return fmt.Sprintf(
			"tcps:%s:%d",
			externalIP,
			externalPort,
		), func() {
			// clear mappings
			gw.DeletePortMapping("tcp", localPort) // nolint: errcheck
		}, nil
}
