package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"nimona.io/internal/log"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/exchange"
	"nimona.io/pkg/storage"
)

//go:generate go run -tags=dev nimona.io/tools/nmake vfsgen

// API for HTTP
type API struct {
	router      *gin.Engine
	key         *crypto.Key
	net         net.Network
	exchange    exchange.Exchange
	objectStore storage.Storage
	localKey    string
	token       string

	version      string
	commit       string
	buildDate    string
	gracefulStop chan bool
	srv          *http.Server
}

// New HTTP API
func New(k *crypto.Key, n net.Network, x exchange.Exchange,
	bls storage.Storage, version, commit, buildDate, token string) *API {
	router := gin.Default()
	router.Use(cors.Default())

	api := &API{
		router:       router,
		key:          k,
		net:          n,
		exchange:     x,
		objectStore:  bls,
		localKey:     n.GetPeerInfo().HashBase58(),
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
	peers.GET("/:peerID", api.HandleGetPeer)

	objectsEnd := router.Group("/api/v1/objects")
	objectsEnd.GET("/", api.HandleGetObjects)
	objectsEnd.GET("/:objectID", api.HandleGetObject)
	objectsEnd.POST("/", api.HandlePostObject)

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
	m := o.ToMap()

	m["_hash"] = o.HashBase58()

	if signer := o.GetSignerKey(); signer != nil {
		m["_signer"] = o.GetSignerKey().HashBase58()
		if api.localKey == signer.HashBase58() {
			m["_direction"] = "outgoing"
		} else {
			m["_direction"] = "incoming"
		}
	}

	if mandateObj := o.GetMandate(); mandateObj != nil {
		mandate := &crypto.Mandate{}
		if err := mandate.FromObject(mandateObj); err == nil {
			m["_authority"] = "identity:" + mandate.Signer.HashBase58()
		}
	}

	recipients := []string{}
	if op := o.GetPolicy(); op != nil {
		p := &object.Policy{}
		p.FromObject(op)
		recipients = append(recipients, p.Subjects...)
	}
	m["_recipients"] = recipients

	um, err := object.UntypeMap(m)
	if err != nil {
		panic(err)
	}

	return um
}
