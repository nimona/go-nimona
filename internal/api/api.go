package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/aggregate"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/object/exchange"
)

//go:generate go run -tags=dev nimona.io/tools/nmake vfsgen

// API for HTTP
type API struct {
	router    *gin.Engine
	key       *crypto.PrivateKey
	net       net.Network
	discovery discovery.Discoverer
	exchange  exchange.Exchange

	objectStore graph.Store
	dag         dag.Manager
	agg         aggregate.Manager
	local       *net.LocalInfo

	localKey string
	token    string

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
	linf *net.LocalInfo,
	bls graph.Store,
	dag dag.Manager,
	agg aggregate.Manager,
	version string,
	commit string,
	buildDate string,
	token string,
) *API {
	router := gin.Default()
	router.Use(cors.Default())

	api := &API{
		router:      router,
		key:         k,
		net:         n,
		discovery:   d,
		exchange:    x,
		objectStore: bls,

		dag: dag,
		agg: agg,

		localKey: linf.GetPeerInfo().Fingerprint(),
		local:    linf,

		version:      version,
		commit:       commit,
		buildDate:    buildDate,
		token:        token,
		gracefulStop: make(chan bool),
	}

	router.Group("/api/v1/")
	router.GET("/version", api.HandleVersion)

	router.Use(api.TokenAuth())

	local := router.Group("/api/v1/local")
	local.GET("/", api.HandleGetLocal)

	peers := router.Group("/api/v1/peers")
	peers.GET("/", api.HandleGetPeers)
	peers.GET("/:fingerprint", api.HandleGetPeer)

	objectsEnd := router.Group("/api/v1/objects")
	objectsEnd.GET("/", api.HandleGetObjects)
	objectsEnd.GET("/:objectHash", api.HandleGetObject)
	objectsEnd.POST("/", api.HandlePostObject)

	graphsEnd := router.Group("/api/v1/graphs")
	graphsEnd.GET("/", api.HandleGetGraphs)
	graphsEnd.POST("/", api.HandlePostGraphs)
	graphsEnd.GET("/:rootObjectHash", api.HandleGetGraph)
	graphsEnd.POST("/:rootObjectHash", api.HandlePostGraph)

	aggregatesEnd := router.Group("/api/v1/aggregates")
	aggregatesEnd.GET("/", api.HandleGetAggregates)
	aggregatesEnd.POST("/", api.HandlePostAggregates)
	aggregatesEnd.GET("/:rootObjectHash", api.HandleGetAggregate)
	aggregatesEnd.POST("/:rootObjectHash", api.HandlePostAggregate)

	streamsEnd := router.Group("/api/v1/streams")
	streamsEnd.GET("/:ns/*pattern", api.HandleGetStreams)

	router.POST("/api/v1/stop", api.Stop)

	router.Use(ServeFs("/", Assets))

	return api
}

// Serve HTTP API
func (api *API) Serve(address string) error {
	ctx := context.Background()
	logger := log.Logger(ctx).Named("api")

	api.srv = &http.Server{
		Addr:    address,
		Handler: api.router,
	}

	go func() {
		if err := api.srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			logger.Error("Error serving", zap.Error(err))
		}
	}()

	<-api.gracefulStop

	if err := api.srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown", zap.Error(err))
	}

	return nil
}

func (api *API) Stop(c *gin.Context) {
	c.Status(http.StatusOK)

	go func() {
		api.gracefulStop <- true
	}()
	return
}

func (api *API) mapObject(o *object.Object) map[string]interface{} {
	m := o.ToPlainMap()
	m["_hash"] = o.HashBase58()
	return m
}
