package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleGetPeers(c *gin.Context) {
	peers, err := api.addressBook.GetAllPeerInfo()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := []map[string]interface{}{}
	for _, v := range peers {
		ms = append(ms, api.mapBlock(v.ToObject()))
	}
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandleGetPeer(c *gin.Context) {
	peerID := c.Param("peerID")
	m, err := api.addressBook.GetPeerInfo(peerID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := api.mapBlock(m.ToObject())
	c.Render(http.StatusOK, Renderer(c, ms))
}
