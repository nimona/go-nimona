package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"nimona.io/go/dht"
	"nimona.io/go/encoding"
	nnet "nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/storage"
)

//go:generate go run -tags=dev assets_generator.go

// API for HTTP
type API struct {
	router      *gin.Engine
	addressBook *peers.AddressBook
	dht         *dht.DHT
	blockStore  storage.Storage
	exchange    nnet.Exchange
	localKey    string
}

// New HTTP API
func New(addressBook *peers.AddressBook, dht *dht.DHT, exchange nnet.Exchange, bls storage.Storage) *API {
	router := gin.Default()
	router.Use(cors.Default())

	api := &API{
		router:      router,
		addressBook: addressBook,
		dht:         dht,
		blockStore:  bls,
		exchange:    exchange,
		localKey:    addressBook.GetLocalPeerInfo().Thumbprint(),
	}

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
	return o.Map()

	// 	"type":        v.Type,
	// 	"payload":     v.Payload,
	// 	"annotations": v.Annotations,
	// }
	// if s := v.Signature; s != nil {
	// 	m["signature"] = v.Signature
	// 	m["owner"] = v.Signature.Key.Thumbprint()
	// 	if v.Signature.Key.Thumbprint() == api.localKey {
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
