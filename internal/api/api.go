package api

import (
	"net/http"

	"nimona.io/internal/http/router"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/peer"
)

// API for HTTP
type API struct {
	router    *router.Router
	net       net.Network
	discovery discovery.Discoverer
	exchange  exchange.Exchange

	objectStore  graph.Store
	orchestrator orchestrator.Orchestrator
	local        *peer.LocalPeer

	localFingerprint crypto.Fingerprint

	token string

	version      string
	commit       string
	buildDate    string
	gracefulStop chan bool
	srv          *http.Server
}

// New HTTP API
func New(
	k *crypto.PrivateKey,
	n net.Network,
	d discovery.Discoverer,
	x exchange.Exchange,
	linf *peer.LocalPeer,
	bls graph.Store,
	orchestrator orchestrator.Orchestrator,
	version string,
	commit string,
	buildDate string,
	token string,
) *API {
	r := router.New()

	api := &API{
		router:      r,
		net:         n,
		discovery:   d,
		exchange:    x,
		objectStore: bls,

		orchestrator: orchestrator,

		localFingerprint: linf.GetFingerprint(),

		local: linf,

		version:      version,
		commit:       commit,
		buildDate:    buildDate,
		token:        token,
		gracefulStop: make(chan bool),
	}

	r.Use(api.Cors())
	r.Use(api.TokenAuth())

	r.Handle("GET", "/api/v1/version$", api.HandleVersion)
	r.Handle("GET", "/api/v1/local$", api.HandleGetLocal)
	r.Handle("GET", "/api/v1/dump$", api.HandleGetDump)

	r.Handle("GET", "/api/v1/identities$", api.HandleGetIdentities)
	r.Handle("GET", "/api/v1/identities/(?P<fingerprint>.+)$", api.HandleGetIdentity)
	r.Handle("POST", "/api/v1/identities$", api.HandlePostIdentities)

	r.Handle("GET", "/api/v1/peers$", api.HandleGetPeers)
	r.Handle("GET", "/api/v1/peers/(?P<fingerprint>.+)$", api.HandleGetPeer)

	r.Handle("GET", "/api/v1/objects$", api.HandleGetObjects)
	r.Handle("GET", "/api/v1/objects/(?P<objectHash>.+)$", api.HandleGetObject)
	r.Handle("POST", "/api/v1/objects$", api.HandlePostObject)

	r.Handle("GET", "/api/v1/graphs$", api.HandleGetGraphs)
	r.Handle("POST", "/api/v1/graphs$", api.HandlePostGraphs)
	r.Handle("GET", "/api/v1/graphs/(?P<rootObjectHash>.+)$", api.HandleGetGraph)
	r.Handle("POST", "/api/v1/graphs/(?P<rootObjectHash>.+)$", api.HandlePostGraph)

	r.Handle("GET", "/api/v1/streams/(?P<ns>.+)/(?P<pattern>.*)$", api.HandleGetStreams)

	r.Handle("POST", "/api/v1/stop$", api.Stop)

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
	return
}

func (api *API) mapObject(o object.Object) map[string]interface{} {
	m := o.ToMap()
	m["_hash.compact"] = hash.New(o).String()
	m["_hash"] = hash.New(o.ToObject()).ToObject()
	if o.GetType() == "nimona.io/crypto/PublicKey" {
		p := &crypto.PublicKey{}
		p.FromObject(o) // nolint: errcheck
		m["_fingerprint.compact"] = p.Fingerprint().String()
	}
	return m
}

func (api *API) mapObjects(os []object.Object) []map[string]interface{} {
	ms := []map[string]interface{}{}
	for _, o := range os {
		ms = append(ms, api.mapObject(o))
	}
	return ms
}
