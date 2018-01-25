package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.uber.org/zap"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

// Ping is our example client, it simply sends a PING string and expects a PONG
type Ping struct{}

// Name of our negotiator
func (p *Ping) Name() string {
	return "ping"
}

// Negotiate will be called after all the other middleware have been processed
func (p *Ping) Negotiate(ctx context.Context, conn fabric.Conn) (context.Context, fabric.Conn, error) {
	lgr := fabric.Logger(ctx).With(
		zap.Namespace("ping"),
	)

	// close conection when done
	defer conn.Close()

	if rp, ok := ctx.Value(fabric.ContextKeyRemoteIdentity).(string); ok {
		lgr.Info("Context contains remote id", zap.String("remote.id", rp))
	}

	// send ping
	if err := fabric.WriteToken(conn, []byte("PING")); err != nil {
		lgr.Error("Could not write token", zap.Error(err))
		return nil, nil, err
	}

	lgr.Info("Wrote token")

	// get pong
	token, err := fabric.ReadToken(conn)
	if err != nil {
		lgr.Error("Could not read token", zap.Error(err))
		return nil, nil, err
	}

	lgr.Info("Read token", zap.String("token", string(token)))

	return nil, nil, nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	p := &Ping{}

	yamux := &fabric.YamuxMiddleware{}
	router := &fabric.RouterMiddleware{}
	identity := &fabric.IdentityMiddleware{Local: "CLIENT"}
	tls := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0:3001"))
	f.AddMiddleware(yamux)
	f.AddMiddleware(router)
	f.AddMiddleware(identity)
	f.AddMiddleware(tls)
	f.AddNegotiator(p)

	// make a new connection to the the server's ping handler
	if _, _, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/tls/router/identity/ping"); err != nil {
		fmt.Println("Dial error", err)
	}
}
