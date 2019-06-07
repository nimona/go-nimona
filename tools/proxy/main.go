package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"golang.org/x/net/http2"
)

type Config struct {
	BindAddress string   `env:"BIND_ADDRESS" envDefault:"0.0.0.0:443"`
	Targets     []string `env:"TARGETS" envSeparator:","`
}

type Health struct{}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func main() {
	c := &Config{}
	if err := env.Parse(c); err != nil {
		log.Fatal("env.Parse:", err)
	}

	// priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, publicKey(priv), priv)
	if err != nil {
		log.Fatal("x509.CreateCertificate:", err)
	}

	tlsCert := tls.Certificate{
		PrivateKey:  priv,
		Certificate: [][]byte{cert},
	}

	config := &tls.Config{
		NextProtos: []string{"h2"},
		Certificates: []tls.Certificate{
			tlsCert,
		},
	}

	targets := map[string]string{}
	for _, target := range c.Targets {
		ps := strings.Split(target, "=")
		if len(ps) != 2 {
			log.Println("* skipping target " + target)
			continue
		}
		log.Println("* adding target " + target)
		targets[ps[0]] = ps[1]
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			NextProtos:         []string{"h2"},
			InsecureSkipVerify: true, // nolint: gosec
		},
	}
	http2.ConfigureTransport(tr)

	handler := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			target := targets[req.Host]
			if target == "" {
				log.Println("unknown source " + req.Host)
				return
			}

			u, _ := url.Parse("https://" + target)
			targetQuery := u.RawQuery
			req.Method = "GET"
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.ContentLength = -1
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
		},
		FlushInterval: 250 * time.Millisecond,
		Transport:     tr,
	}

	srv := &http.Server{
		Addr:    c.BindAddress,
		Handler: handler,
	}

	netListener, err := net.Listen("tcp", c.BindAddress)
	if err != nil {
		log.Fatal("net.Listen:", err)
	}

	tlsListener := tls.NewListener(
		netListener,
		config,
	)

	go http.ListenAndServe("0.0.0.0:80", &Health{})

	log.Println("Starting proxy server on", c.BindAddress)
	if err := srv.Serve(tlsListener); err != nil {
		log.Fatal("srv.Serve:", err)
	}

	// domains := []string{}
	// for domain := range c.Targets {
	// 	domains = append(domains, domain)
	// }

	// certManager := autocert.Manager{
	// 	Email:      c.LetsEncryptEmail,
	// 	Prompt:     autocert.AcceptTOS,
	// 	HostPolicy: autocert.HostWhitelist(domains...),
	// 	Cache:      autocert.DirCache("certs"),
	// }

	// go http.ListenAndServe("0.0.0.0:80", certManager.HTTPHandler(nil))

	// server := &http.Server{
	// 	Addr:    c.BindAddress,
	// 	Handler: handler,
	// 	TLSConfig: &tls.Config{
	// 		GetCertificate: certManager.GetCertificate,
	// 	},
	// }

	// log.Println("Starting proxy server on", c.BindAddress)
	// if err := server.ListenAndServeTLS("", ""); err != nil {
	// 	log.Fatal("ListenAndServe:", err)
	// }
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
