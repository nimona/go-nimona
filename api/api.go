package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

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

	version   string
	commit    string
	buildDate string
}

// New HTTP API
func New(k *crypto.Key, n nnet.Network, x nnet.Exchange, dht *dht.DHT,
	bls storage.Storage, version, commit, buildDate string) *API {
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
	}

	router.Group("/api/v1/")
	router.GET("/version", api.HandleVersion)

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

	router.Use(ServeFs("/", Assets))

	return api
}

// Serve HTTP API
func (api *API) Serve(address string) error {
	return api.router.Run(address)
}

func (api *API) mapBlock(o *encoding.Object) map[string]interface{} {
	return o.ToMap()

	// 	"type":        v.Type,
	// 	"payload":     v.Payload,
	// 	"annotations": v.Annotations,
	// }
	// if s := v.Signature; s != nil {
	// 	m["signature"] = v.Signature
	// 	m["owner"] = v.Signature.Key.HashBase58()
	// 	if v.Signature.Key.HashBase58() == api.localKey {
	// 		m["direction"] = "outgoing"
	// 	} else {
	// 		m["direction"] = "incoming"
	// 	}
	// }
	// recipients := []string{}
	// if v.Annotations != nil {
	// 	for _, policy := range v.Annotations.Policies {
	// 		recipients = append(recipients, policy.Subjects...)
	// 	}
	// }
	// m["id"] = crypto.ID(v)
	// m["recipients"] = recipients
	// return m
}
