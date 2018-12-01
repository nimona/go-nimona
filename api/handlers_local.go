package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleGetLocal(c *gin.Context) {
	v := api.addressBook.GetLocalPeerInfo()
	ms := api.mapBlock(v.ToObject())
	c.Render(http.StatusOK, Renderer(c, ms))
}
