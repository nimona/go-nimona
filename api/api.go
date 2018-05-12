package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
)

type API struct {
	Router *gin.Engine
}

func New(reg mesh.Registry, dht *dht.DHT) *API {
	router := gin.Default()
	router.Use(cors.Default())
	local := router.Group("/api/v1/local")
	local.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, reg.GetLocalPeerInfo())
	})
	peers := router.Group("/api/v1/peers")
	peers.GET("/", func(c *gin.Context) {
		peers, err := reg.GetAllPeerInfo()
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
		Router: router,
	}
}

func (api *API) Serve(address string) error {
	return api.Router.Run(address)
}
