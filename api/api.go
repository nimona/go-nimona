package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/net"
)

type API struct {
	router *gin.Engine
}

func New(addressBook net.PeerManager, dht *dht.DHT) *API {
	router := gin.Default()
	router.Use(cors.Default())
	local := router.Group("/api/v1/local")
	local.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, addressBook.GetLocalPeerInfo())
	})
	peers := router.Group("/api/v1/peers")
	peers.GET("/", func(c *gin.Context) {
		peers, err := addressBook.GetAllPeerInfo()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, peers)
	})
	values := router.Group("/api/v1/values")
	values.GET("/", func(c *gin.Context) {
		values, err := dht.GetAllValues()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, values)
	})
	providers := router.Group("/api/v1/providers")
	providers.GET("/", func(c *gin.Context) {
		providers, err := dht.GetAllProviders()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, providers)
	})
	return &API{
		router: router,
	}
}

func (api *API) Serve(address string) error {
	return api.router.Run(address)
}
