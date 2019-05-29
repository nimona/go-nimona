package net

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"net"
	"strings"
	"time"

	igd "github.com/emersion/go-upnp-igd"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
)

type tcpTransport struct {
	local   *LocalInfo
	address string
}

func NewTCPTransport(
	local *LocalInfo,
	address string,
) Transport {
	return &tcpTransport{
		local:   local,
		address: address,
	}
}

func (tt *tcpTransport) Dial(ctx context.Context, address string) (
	*Connection, error) {

	config := tls.Config{
		InsecureSkipVerify: true,
	}
	addr := strings.Replace(address, "tcps:", "", 1)
	dialer := net.Dialer{Timeout: time.Second}

	tcpConn, err := tls.DialWithDialer(&dialer, "tcp", addr, &config)
	if err != nil {
		return nil, err
	}

	if tcpConn == nil {
		return nil, ErrAllAddressesFailed
	}

	conn := &Connection{
		Conn:          tcpConn,
		RemotePeerKey: nil, // we don't really know who the other side is
	}

	return conn, nil
}

func (tt *tcpTransport) Listen(ctx context.Context) (
	chan *Connection, error) {

	logger := log.FromContext(ctx).Named("network")
	cert, err := crypto.GenerateCertificate(tt.local.GetPeerKey())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	config.NextProtos = []string{"nimona/1"} // TODO(geoah) is this of any actual use?
	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader
	tcpListener, err := tls.Listen("tcp", tt.address, &config)
	if err != nil {
		return nil, err
	}

	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", log.Int("port", port))
	addresses := GetAddresses("tcps", tcpListener)
	devices := make(chan igd.Device, 10)

	if tt.local.GetHostname() != "" {
		addresses = append(addresses, fmtAddress(
			"tcps",
			tt.local.GetHostname(),
			port,
		))
	}

	if UseUPNP {
		logger.Info("Trying to find external IP and open port")
		go func() {
			if err := igd.Discover(devices, 2*time.Second); err != nil {
				logger.Error("could not discover devices", log.Error(err))
			}
		}()
		for device := range devices {
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				logger.Error("could not get external ip", log.Error(err))
				continue
			}
			desc := "nimona-tcp"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				logger.Error("could not add port mapping", log.Error(err))
			} else {
				addresses = append(addresses, fmtAddress(
					"tcps",
					externalAddress.String(),
					port,
				))
			}
		}
	}

	logger.Info("Started listening", log.Strings("addresses", addresses))
	tt.local.AddAddress(addresses...)

	cconn := make(chan *Connection, 10)
	go func() {

		for {
			tcpConn, err := tcpListener.Accept()
			if err != nil {
				log.DefaultLogger.Warn(
					"could not accept connection", log.Error(err))
				continue
			}

			conn := &Connection{
				Conn:          tcpConn,
				RemotePeerKey: nil,
				IsIncoming:    true,
			}

			cconn <- conn
		}
	}()

	return cconn, nil
}
