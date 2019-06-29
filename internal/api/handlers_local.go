package api

import (
	"net/http"

	"nimona.io/internal/http/router"
)

func (api *API) HandleGetLocal(c *router.Context) {
	ms := api.mapObject(api.local.GetPeerInfo().ToObject())
	c.JSON(http.StatusOK, ms)
}
