package net

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"go.uber.org/zap"

	"nimona.io/go/log"
	"nimona.io/go/peers"
)

// Networker interface for mocking Network
type Networker interface {
	Dial(ctx context.Context, peerID string) (net.Conn, error)
	Listen(ctx context.Context, addrress string) (net.Listener, error)
}

// NewNetwork creates a new p2p network using an address book
func NewNetwork(addressBook *peers.AddressBook) (*Network, error) {
	return &Network{
		addressBook: addressBook,
	}, nil
}

// Network allows dialing and listening for p2p connections
type Network struct {
	addressBook *peers.AddressBook
}

// Dial to a peer and return a net.Conn or error
func (n *Network) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	logger := log.Logger(ctx)
	peerInfo, err := n.addressBook.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	if len(peerInfo.Addresses) == 0 {
		return nil, ErrNoAddresses
	}

	var conn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second}
		logger.Debug("dialing", zap.String("address", addr))
		newConn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			continue
		}
		conn = newConn
		break
	}

	if conn == nil {
		return nil, ErrAllAddressesFailed
	}

	return conn, nil

}

// Listen on an address
// TODO do we need to return a listener?
func (n *Network) Listen(ctx context.Context, addr string) (net.Listener, error) {
	logger := log.Logger(ctx).Named("network")
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", zap.Int("port", port))
	addresses := GetAddresses(tcpListener)
	devices := make(chan igd.Device, 10)

	upnp := true
	upnpFlag := os.Getenv("UPNP")
	if upnpFlag != "" {
		upnp, _ = strconv.ParseBool(upnpFlag)
	}
	if upnp {
		logger.Info("Trying to find external IP and open port")
		go func() {
			if err := igd.Discover(devices, 2*time.Second); err != nil {
				logger.Error("could not discover devices", zap.Error(err))
			}
		}()
		for device := range devices {
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				logger.Error("could not get external ip", zap.Error(err))
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				logger.Error("could not add port mapping", zap.Error(err))
			} else {
				addresses = append(addresses, fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port))
			}
		}
	}

	logger.Info("Started listening", zap.Strings("addresses", addresses))
	n.addressBook.AddAddress(addresses...)

	return tcpListener, nil
}
