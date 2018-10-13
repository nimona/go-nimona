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
	// TODO return actuall provider blocks, or remove endpoint
	c.Render(http.StatusOK, Renderer(c, providers))
}
