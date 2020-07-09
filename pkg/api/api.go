package api

import (
	"net/http"
	"runtime/debug"
	"runtime/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/http/router"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/streammanager"
)

// API for HTTP
type API struct {
	config *config.Config

	router   *router.Router
	keychain keychain.Keychain
	net      net.Network
	resolver resolver.Resolver
	exchange exchange.Exchange

	objectStore   *sqlobjectstore.Store
	streammanager streammanager.StreamManager

	token string

	version      string
	commit       string
	buildDate    string
	gracefulStop chan bool
	srv          *http.Server
}

// New HTTP API
func New(
	cfg *config.Config,
	k crypto.PrivateKey,
	kc keychain.Keychain,
	n net.Network,
	d resolver.Resolver,
	x exchange.Exchange,
	sst *sqlobjectstore.Store,
	or streammanager.StreamManager,
	version string,
	commit string,
	buildDate string,
	token string,
) *API {
	r := router.New()

	api := &API{
		config: cfg,

		router:      r,
		keychain:    kc,
		net:         n,
		resolver:    d,
		exchange:    x,
		objectStore: sst,

		streammanager: or,

		version:      version,
		commit:       commit,
		buildDate:    buildDate,
		token:        token,
		gracefulStop: make(chan bool),
	}

	r.Use(api.Cors())
	r.Use(api.TokenAuth())

	r.Handle(
		"GET",
		"/api/v1/version$",
		api.HandleVersion,
	)
	r.Handle(
		"GET",
		"/api/v1/local$",
		api.HandleGetLocal,
	)

	r.Handle(
		"GET",
		"/api/v1/identities$",
		api.HandleGetIdentities,
	)
	r.Handle(
		"GET",
		"/api/v1/identities/(?P<fingerprint>.+)$",
		api.HandleGetIdentity,
	)
	r.Handle(
		"POST",
		"/api/v1/identities$",
		api.HandlePostIdentities,
	)

	r.Handle(
		"GET",
		"/api/v1/peers$",
		api.HandleGetPeers,
	)
	r.Handle(
		"GET",
		"/api/v1/peers/(?P<fingerprint>.+)$",
		api.HandleGetPeer,
	)

	r.Handle(
		"GET",
		"/api/v1/objects$",
		api.HandleGetObjects,
	)
	r.Handle(
		"POST",
		"/api/v1/objects$",
		api.HandlePostObjects,
	)
	r.Handle(
		"POST",
		"/api/v1/stop$",
		api.Stop,
	)
	if cfg.Peer.EnableMetrics {
		r.Handle(
			"GET",
			"/metrics",
			func(c *router.Context) {
				promhttp.Handler().ServeHTTP(c.Writer, c.Request)
			},
		)
	}
	if cfg.Peer.EnableDebug {
		goroutineStackHandler := func(c *router.Context) {
			stack := debug.Stack()
			c.Writer.Write(stack)                          // nolint: errcheck
			pprof.Lookup("goroutine").WriteTo(c.Writer, 2) // nolint: errcheck
		}
		r.Handle(
			"GET",
			"/debug/stack/goroutine",
			goroutineStackHandler,
		)
	}

	return api
}

// Serve HTTP API
func (api *API) Serve(address string) error {
	ctx := context.Background()
	logger := log.FromContext(ctx).Named("api")

	api.srv = &http.Server{
		Addr:    address,
		Handler: api.router,
	}

	go func() {
		if err := api.srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			logger.Error("Error serving", log.Error(err))
		}
	}()

	<-api.gracefulStop

	if err := api.srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown", log.Error(err))
	}

	return nil
}

func (api *API) Stop(c *router.Context) {
	c.Status(http.StatusOK)
	go func() {
		api.gracefulStop <- true
	}()
}

func (api *API) mapObject(o object.Object) map[string]interface{} {
	m := o.ToMap()
	m["_hash:s"] = o.Hash().String()
	return m
}

func (api *API) mapObjects(objects []object.Object) []map[string]interface{} {
	ms := []map[string]interface{}{}
	for _, o := range objects {
		ms = append(ms, api.mapObject(o))
	}
	return ms
}
