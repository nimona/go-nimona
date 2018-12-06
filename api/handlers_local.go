package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleGetLocal(c *gin.Context) {
	ms := api.mapBlock(api.net.GetPeerInfo().ToObject())
	c.Render(http.StatusOK, Renderer(c, ms))
}
