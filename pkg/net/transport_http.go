package net

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"golang.org/x/net/http2"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
)

type httpTransport struct {
	local   *LocalInfo
	address string
}

func NewHTTPTransport(
	local *LocalInfo,
	address string,
) Transport {
	return &httpTransport{
		local:   local,
		address: address,
	}
}

func (tt *httpTransport) Dial(
	ctx context.Context,
	address string,
) (
	*Connection,
	error,
) {
	address = strings.Replace(address, "https:", "https://", 1)
	address = strings.Replace(address, ":443", "", 1)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint: gosec
		},
	}
	if err := http2.ConfigureTransport(tr); err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: tr,
	}
	pr, pw := io.Pipe()
	req, _ := http.NewRequest("GET", address, ioutil.NopCloser(pr))
	req.ContentLength = -1
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	rw := &connWrapper{
		w: pw,
		r: res.Body,
		c: res.Body,
	}

	conn := &Connection{
		Conn:          rw,
		RemotePeerKey: nil, // we don't really know who the other side is
	}

	return conn, nil
}

func (tt *httpTransport) Listen(
	ctx context.Context,
) (
	chan *Connection,
	error,
) {

	logger := log.FromContext(ctx).Named("transport/https")

	cert, err := crypto.GenerateCertificate(tt.local.GetPeerKey())
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		NextProtos:   []string{"h2"},
		Certificates: []tls.Certificate{*cert},
	}

	cconn := make(chan *Connection, 10)

	handler := func(w http.ResponseWriter, r *http.Request) {
		wf, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "only h2 is supported", 400)
			return
		}

		wf.Flush()

		rw := &connWrapper{
			w: w,
			r: r.Body,
			c: r.Body,
		}

		conn := &Connection{
			Conn:          rw,
			RemotePeerKey: nil,
			IsIncoming:    true,
		}

		cconn <- conn

		<-r.Cancel // TODO is this the right way to wait here?
	}

	srv := &http.Server{
		Addr:    tt.address,
		Handler: http.HandlerFunc(handler),
	}

	netListener, err := net.Listen("tcp", tt.address)
	if err != nil {
		return nil, err
	}

	tlsListener := tls.NewListener(
		tcpKeepAliveListener{
			netListener.(*net.TCPListener),
		},
		config,
	)

	go func() {
		err := srv.Serve(tlsListener)
		if err != nil {
			logger.Error("http transport stopped", log.Error(err))
		}
	}()

	port := netListener.Addr().(*net.TCPAddr).Port
	logger.Info("HTTP tranport listening", log.Int("port", port))
	devices := make(chan igd.Device, 10)

	useIPs := true

	addresses := []string{}

	if tt.local.GetHostname() != "" {
		useIPs = false
		addresses = append(addresses, fmtAddress("https", tt.local.GetHostname(), port))
	}

	if useIPs {
		addresses = append(addresses, GetAddresses("https", netListener)...)

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
			desc := "nimona-http"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				logger.Error("could not add port mapping", log.Error(err))
			} else if useIPs {
				addresses = append(addresses, fmtAddress("https", externalAddress.String(), port))
			}
		}
	}

	tt.local.AddAddress(addresses...)

	logger.Info(
		"Started listening",
		log.Strings("addresses", addresses),
		log.Int("port", port),
	)

	return cconn, nil
}

type connWrapper struct {
	r io.Reader
	w io.Writer
	c io.Closer
	l sync.Mutex
}

func (fw connWrapper) Write(p []byte) (int, error) {
	fw.l.Lock()
	defer fw.l.Unlock()

	n, err := fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return n, err
}

func (fw connWrapper) Read(b []byte) (int, error) {
	return fw.r.Read(b)
}

func (fw connWrapper) Close() error {
	return fw.c.Close()
}

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	if err := tc.SetKeepAlive(true); err != nil {
		return nil, err
	}
	if err := tc.SetKeepAlivePeriod(time.Minute); err != nil {
		return nil, err
	}
	return tc, nil
}
