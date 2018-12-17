package api

import (
	"context"

	"net/http"
	"os"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nimona.io/go/log"

	"nimona.io/go/crypto"
	"nimona.io/go/dht"
	"nimona.io/go/encoding"
	nnet "nimona.io/go/net"
	"nimona.io/go/storage"
)

//go:generate go run -tags=dev assets_generator.go

// API for HTTP
type API struct {
	router     *gin.Engine
	key        *crypto.Key
	net        nnet.Network
	exchange   nnet.Exchange
	dht        *dht.DHT
	blockStore storage.Storage
	localKey   string
	token      string

	version   string
	commit    string
	buildDate string
}

// New HTTP API
func New(k *crypto.Key, n nnet.Network, x nnet.Exchange, dht *dht.DHT,
	bls storage.Storage, version, commit, buildDate, token string) *API {
	router := gin.Default()
	router.Use(cors.Default())

	api := &API{
		router:     router,
		key:        k,
		net:        n,
		exchange:   x,
		dht:        dht,
		blockStore: bls,
		version:    version,
		commit:     commit,
		buildDate:  buildDate,
		token:      token,
	}

	router.Group("/api/v1/")
	router.GET("/version", api.HandleVersion)

	router.Use(api.TokenAuth())

	local := router.Group("/api/v1/local")
	local.GET("/", api.HandleGetLocal)

	peers := router.Group("/api/v1/peers")
	peers.GET("/", api.HandleGetPeers)
	peers.GET("/:peerID", api.HandleGetPeer)

	providers := router.Group("/api/v1/providers")
	providers.GET("/", api.HandleGetProviders)

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
	return api.router.Run(address)
}

func (api *API) Stop(c *gin.Context) {
	ctx := context.Background()
	logger := log.Logger(ctx).Named("api")

	c.Status(http.StatusOK)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		logger.Error("Failed to find process", zap.Error(err))
		return
	}

	if err := p.Signal(syscall.SIGTERM); err != nil {
		logger.Error("Failed kill process", zap.Error(err))

		return
	}

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
