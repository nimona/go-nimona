package net

import (
	"crypto/rand"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
)

type tcpTransport struct {
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

func (tt *tcpTransport) Listen(
	ctx context.Context,
	bindAddress string,
	key crypto.PrivateKey,
) (net.Listener, error) {
	cert, err := crypto.GenerateTLSCertificate(key)
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
	tcpListener, err := tls.Listen("tcp", bindAddress, &config)
	if err != nil {
		return nil, err
	}

	return tcpListener, nil
}
