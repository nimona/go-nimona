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
	"github.com/nimona/go-nimona/log"
	"go.uber.org/zap"
)

// Networker interface for mocking Network
type Networker interface {
	Dial(ctx context.Context, peerID string) (net.Conn, error)
	Listen(ctx context.Context, addrress string) (net.Listener, error)
}

// NewNetwork creates a new p2p network using an address book
func NewNetwork(AddressBook AddressBooker) (*Network, error) {
	return &Network{
		AddressBook: AddressBook,
	}, nil
}

// Network allows dialing and listening for p2p connections
type Network struct {
	AddressBook AddressBooker
}

// Dial to a peer and return a net.Conn or error
func (n *Network) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	peerInfo, err := n.AddressBook.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second * 5}
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
	newAddresses := make(chan string, 100)
	devices := make(chan igd.Device)
	go func() {
		for device := range devices {
			upnp := true
			upnpFlag := os.Getenv("UPNP")
			if upnpFlag != "" {
				upnp, _ = strconv.ParseBool(upnpFlag)
			}
			if !upnp {
				continue
			}
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
				newAddresses <- fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port)
			}
		}
		close(newAddresses)
	}()

	// go func() {
	if err := igd.Discover(devices, 2*time.Second); err != nil {
		close(newAddresses)
		logger.Error("could not discover devices", zap.Error(err))
	}

	addresses := GetAddresses(tcpListener)
	for newAddress := range newAddresses {
		addresses = append(addresses, newAddress)
	}

	// TODO Replace with actual relay peer ids
	addresses = append(addresses, "relay:01x2Adrt7msBM2ZBW16s9SbJcnnqwG8UQme9VTcka5s7T9Z1")

	localPeerInfo := n.AddressBook.GetLocalPeerInfo()
	localPeerInfo.Addresses = addresses
	n.AddressBook.PutLocalPeerInfo(localPeerInfo)
	// }()

	return tcpListener, nil
}
