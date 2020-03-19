package net

import (
	"crypto/rand"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"gitlab.com/NebulousLabs/go-upnp"

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
		InsecureSkipVerify: true, // nolint: gosec
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
	cert, err := crypto.GenerateTLSCertificate(tt.local.GetPeerPrivateKey())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	// TODO(geoah) is this of any actual use?
	config.NextProtos = []string{"nimona/1"}
	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader
	tcpListener, err := tls.Listen("tcp", tt.address, &config)
	if err != nil {
		return nil, err
	}

	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", log.Int("port", port))

	useIPs := true

	if tt.local.GetHostname() != "" {
		useIPs = false
		tt.local.AddAddress("tcps-hostname", []string{
			fmtAddress(
				"tcps",
				tt.local.GetHostname(),
				port,
			),
		})
	}

	if useIPs {
		tt.local.AddAddress("tcps-local", GetAddresses("tcps", tcpListener))
	}

	if UseUPNP && useIPs {
		go func() {
			logger.Info("Trying to find external IP and open port")

			// connect to router
			d, err := upnp.Discover()
			if err != nil {
				logger.Error("could not discover devices", log.Error(err))
				return
			}

			// discover external IP
			ip, err := d.ExternalIP()
			if err != nil {
				logger.Error("could not discover external ip", log.Error(err))
				return
			}

			// add port mapping
			err = d.Forward(uint16(port), "nimona daemon")
			if err != nil {
				logger.Error("could not forward port", log.Error(err))
				return
			}

			tt.local.AddAddress("tcps-upnp", []string{
				fmtAddress(
					"tcps",
					ip,
					port,
				),
			})

			logger.Info(
				"created port mapping",
				log.String("externalAddress", ip),
				log.Int("port", port),
				log.Strings("addresses", tt.local.GetAddresses()),
			)
		}()
	}

	logger.Info("Started listening")

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
