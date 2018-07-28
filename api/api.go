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

func New(addressBook net.AddressBooker, dht *dht.DHT, bls net.Storage) *API {
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
	providers := router.Group("/api/v1/providers")
	providers.GET("/", func(c *gin.Context) {
		providers, err := dht.GetAllProviders()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, providers)
	})
	blocks := router.Group("/api/v1/blocks")
	blocks.GET("/", func(c *gin.Context) {
		blockIDs, err := bls.List()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		blocks := []*net.Block{}
		for _, blockID := range blockIDs {
			block, err := bls.Get(blockID)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			blocks = append(blocks, block)
		}
		c.JSON(http.StatusOK, blocks)
	})
	blocks.GET("/:blockID", func(c *gin.Context) {
		blockID := c.Param("blockID")
		block, err := bls.Get(blockID)
		if err != nil {
			if err == net.ErrNotFound {
				c.AbortWithError(404, err)
				return
			}
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, block)
	})
	return &API{
		router: router,
	}
}

func (api *API) Serve(address string) error {
	return api.router.Run(address)
}
