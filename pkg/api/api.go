package api

import (
	"expvar"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"

	"nimona.io/pkg/keychain"

	"github.com/zserge/metric"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/http/router"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/sqlobjectstore"
)

// API for HTTP
type API struct {
	config *config.Config

	router    *router.Router
	keychain  keychain.Keychain
	net       net.Network
	discovery discovery.Discoverer
	exchange  exchange.Exchange

	objectStore  *sqlobjectstore.Store
	orchestrator orchestrator.Orchestrator

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
	d discovery.Discoverer,
	x exchange.Exchange,
	sst *sqlobjectstore.Store,
	or orchestrator.Orchestrator,
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
		discovery:   d,
		exchange:    x,
		objectStore: sst,

		orchestrator: or,

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
		"/api/v1/peers",
		api.HandleGetLookup,
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
		"GET",
		"/api/v1/objects/(?P<objectHash>.+)$",
		api.HandleGetObject,
	)
	r.Handle(
		"POST",
		"/api/v1/objects$",
		api.HandlePostObjects,
	)
	r.Handle(
		"POST",
		"/api/v1/objects/(?P<rootObjectHash>.+)$",
		api.HandlePostObject,
	)

	r.Handle(
		"GET",
		"/api/v1/streams/(?P<ns>.+)/(?P<pattern>.*)$",
		api.HandleGetStreams,
	)

	r.Handle(
		"POST",
		"/api/v1/stop$",
		api.Stop,
	)

	if y, _ := strconv.ParseBool(os.Getenv("NIMONA_EXPVAR")); y {
		expvar.Publish(
			"go:goroutine",
			metric.NewGauge("2m1s", "15m30s", "1h1m"),
		)
		expvar.Publish(
			"go:cgocall",
			metric.NewGauge("2m1s", "15m30s", "1h1m"),
		)
		expvar.Publish(
			"go:alloc",
			metric.NewGauge("2m1s", "15m30s", "1h1m"),
		)
		expvar.Publish(
			"go:alloc.total",
			metric.NewGauge("2m1s", "15m30s", "1h1m"),
		)
		go func() {
			for range time.Tick(100 * time.Millisecond) {
				m := &runtime.MemStats{}
				runtime.ReadMemStats(m)
				expvar.Get("go:goroutine").(metric.Metric).Add(
					float64(runtime.NumGoroutine()),
				)
				expvar.Get("go:cgocall").(metric.Metric).Add(
					float64(runtime.NumCgoCall()),
				)
				expvar.Get("go:alloc").(metric.Metric).Add(
					float64(m.Alloc) / 1000000,
				)
				expvar.Get("go:alloc.total").(metric.Metric).Add(
					float64(m.TotalAlloc) / 1000000,
				)
			}
		}()
		r.Handle(
			"GET",
			"/debug/metrics",
			func(c *router.Context) {
				metric.Handler(metric.Exposed).ServeHTTP(c.Writer, c.Request)
			},
		)
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
		r.Handle(
			"GET",
			"/debug/expvar",
			func(c *router.Context) {
				expvar.Handler().ServeHTTP(c.Writer, c.Request)
			},
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
	m["_hash:s"] = object.NewHash(o).String()
	return m
}

func (api *API) mapObjects(objects []object.Object) []map[string]interface{} {
	ms := []map[string]interface{}{}
	for _, o := range objects {
		ms = append(ms, api.mapObject(o))
	}
	return ms
}
