package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/peers"
	"github.com/nimona/go-nimona/storage"
)

type API struct {
	router *gin.Engine
}

func New(addressBook peers.AddressBooker, dht *dht.DHT, bls storage.Storage) *API {
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
	blocksEnd := router.Group("/api/v1/blocks")
	blocksEnd.GET("/", func(c *gin.Context) {
		blockIDs, err := bls.List()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		bs := []*blocks.Block{}
		for _, blockID := range blockIDs {
			block, err := bls.Get(blockID)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			bs = append(bs, block)
		}
		c.JSON(http.StatusOK, bs)
	})
	blocksEnd.GET("/:blockID", func(c *gin.Context) {
		blockID := c.Param("blockID")
		block, err := bls.Get(blockID)
		if err != nil {
			if err == storage.ErrNotFound {
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
