package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.uber.org/zap"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

func handler(ctx context.Context, c fabric.Conn) (context.Context, fabric.Conn, error) {
	lgr := fabric.Logger(ctx).With(
		zap.Namespace("ping"),
	)

	lgr.Info("Handling new request")

	// close connection when done
	defer c.Close()

	if rp, ok := ctx.Value(fabric.ContextKeyRemoteIdentity).(string); ok {
		lgr.Info("Context contains remote id", zap.String("remote.id", rp))
	}

	// // client pings
	// fmt.Println("Ping: Reading ping from", rp)
	token, err := fabric.ReadToken(c)
	if err != nil {
		lgr.Error("Could not read token", zap.Error(err))
		return nil, nil, err
	}

	lgr.Info("Read token", zap.String("token", string(token)))

	// we pong back
	if err := fabric.WriteToken(c, []byte("PONG")); err != nil {
		lgr.Error("Could not write token", zap.Error(err))
		return nil, nil, err
	}

	lgr.Info("Wrote token")

	// TODO return connection as it was?
	return nil, nil, nil
}

func main() {
	ctx := context.Background()
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	yamux := &fabric.YamuxMiddleware{}
	router := &fabric.RouterMiddleware{}
	identity := &fabric.IdentityMiddleware{Local: "SERVER"}
	tls := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0:3000"))

	f.AddMiddleware(yamux)
	f.AddMiddleware(identity)

	f.AddHandlerFunc("ping", handler)
	f.AddHandlerFunc("identity/ping", handler)

	fmt.Println("Listening...")

	f.Listen(ctx)
}
