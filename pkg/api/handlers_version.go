package api

import (
	"net/http"

	"nimona.io/pkg/http/router"
)

func (api *API) HandleVersion(c *router.Context) {
	d := map[string]interface{}{
		"version": api.version,
		"commit":  api.commit,
		"date":    api.buildDate,
	}
	c.JSON(http.StatusOK, d)
}
