package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleGetProviders(c *gin.Context) {
	providers, err := api.dht.GetAllProviders()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(http.StatusOK, providers)
}
