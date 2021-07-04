package net

import (
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
)

type tcpTransport struct {
	localpeer localpeer.LocalPeer
}

func (tt *tcpTransport) Dial(
	ctx context.Context,
	address string,
) (*Connection, error) {
	// TODO we probably should not be generating the certificate every time
	// but at this point it's kind of annoying to cache the primary peer key
	// TODO consider storing ready made certificated in the localpeer
	cert, err := crypto.GenerateTLSCertificate(
		tt.localpeer.GetPeerKey(),
	)
	if err != nil {
		return nil, err
	}

	config := tls.Config{
		Certificates:       []tls.Certificate{*cert},
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

	if err := tcpConn.Handshake(); err != nil {
		// not currently supported
		// TODO find a way to surface this error
		conn.Close() // nolint: errcheck
		return nil, fmt.Errorf("could not complete handshake")
	}

	state := tcpConn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) != 1 {
		conn.Close() // nolint: errcheck
		return nil, fmt.Errorf("only single certs are currently supported")
	}

	pubKey, ok := certs[0].PublicKey.(ed25519.PublicKey)
	if !ok {
		conn.Close() // nolint: errcheck
		return nil, fmt.Errorf("only ed25519 keys are currently supported")
	}

	conn.RemotePeerKey = crypto.NewEd25519PublicKeyFromRaw(
		pubKey,
		,
	)

	return conn, nil
}

func (tt *tcpTransport) Listen(
	ctx context.Context,
	bindAddress string,
	key crypto.PrivateKey,
) (net.Listener, error) {
	cert, err := crypto.GenerateTLSCertificate(key)
	if err != nil {
		return nil, err
	}

	config := tls.Config{
		Certificates:       []tls.Certificate{*cert},
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true, // nolint: gosec
	}
	// TODO(geoah) is this of any actual use?
	config.NextProtos = []string{"nimona/1"}
	tcpListener, err := tls.Listen("tcp", bindAddress, &config)
	if err != nil {
		return nil, err
	}

	return tcpListener, nil
}
