package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleGetLocal(c *gin.Context) {
	v := api.addressBook.GetLocalPeerInfo()
	c.JSON(http.StatusOK, api.mapBlock(v.Block()))
}
