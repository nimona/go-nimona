package net

import (
	"crypto/rand"
	"crypto/tls"
	"net"
	"strings"
	"time"

	igd "github.com/emersion/go-upnp-igd"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
)

type tcpTransport struct {
	local   *peer.LocalPeer
	address string
}

func NewTCPTransport(
	local *peer.LocalPeer,
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

	conn := newConnection(tcpConn, false)
	conn.remoteAddress = address
	conn.localAddress = tcpConn.LocalAddr().String()

	return conn, nil
}

func (tt *tcpTransport) Listen(ctx context.Context) (
	chan *Connection, error) {
	logger := log.FromContext(ctx).Named("network")
	cert, err := crypto.GenerateCertificate(tt.local.GetPeerPrivateKey())
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
	devices := make(chan igd.Device, 10)

	useIPs := true
	addresses := []string{}

	if tt.local.GetHostname() != "" {
		useIPs = false
		addresses = append(addresses, fmtAddress(
			"tcps",
			tt.local.GetHostname(),
			port,
		))
	}

	if useIPs {
		addresses = append(addresses, GetAddresses("tcps", tcpListener)...)
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
			} else if useIPs {
				addresses = append(addresses, fmtAddress(
					"tcps",
					externalAddress.String(),
					port,
				))
			}
		}
	}

	logger.Info("Started listening", log.Strings("addresses", addresses))
	tt.local.AddAddress("tcps", addresses)

	cconn := make(chan *Connection, 10)
	go func() {
		for {
			tcpConn, err := tcpListener.Accept()
			if err != nil {
				log.DefaultLogger.Warn(
					"could not accept connection", log.Error(err))
				continue
			}

			conn := newConnection(tcpConn, true)
			conn.remoteAddress = tcpConn.RemoteAddr().String()
			conn.localAddress = tcpConn.LocalAddr().String()
			cconn <- conn
		}
	}()

	return cconn, nil
}
