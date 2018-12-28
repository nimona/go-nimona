package api

import (
	"context"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nimona.io/go/log"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	nnet "nimona.io/go/net"
	"nimona.io/go/storage"
)

//go:generate go run vfsgendev -source="nimona.io/go/api".Assets

// API for HTTP
type API struct {
	router     *gin.Engine
	key        *crypto.Key
	net        nnet.Network
	exchange   nnet.Exchange
	blockStore storage.Storage
	localKey   string
	token      string

	version      string
	commit       string
	buildDate    string
	gracefulStop chan bool
	srv          *http.Server
}

// New HTTP API
func New(k *crypto.Key, n nnet.Network, x nnet.Exchange,
	bls storage.Storage, version, commit, buildDate, token string) *API {
	router := gin.Default()
	router.Use(cors.Default())

	api := &API{
		router:       router,
		key:          k,
		net:          n,
		exchange:     x,
		blockStore:   bls,
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

	blocksEnd := router.Group("/api/v1/blocks")
	blocksEnd.GET("/", api.HandleGetBlocks)
	blocksEnd.GET("/:blockID", api.HandleGetBlock)
	blocksEnd.POST("/", api.HandlePostBlock)

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

func (api *API) mapBlock(o *encoding.Object) map[string]interface{} {
	m := o.ToMap()

	m["_hash"] = o.HashBase58()

	if signer := o.GetSignerKey(); signer != nil {
		if api.localKey == signer.HashBase58() {
			m["_direction"] = "outgoing"
		} else {
			m["_direction"] = "incoming"
		}
	}

	recipients := []string{}
	if op := o.GetPolicy(); op != nil {
		p := &encoding.Policy{}
		p.FromObject(op)
		recipients = append(recipients, p.Subjects...)
	}
	m["_recipients"] = recipients

	um, err := encoding.UntypeMap(m)
	if err != nil {
		panic(err)
	}

	return um
}
