package api

import (
	"net/http"

	"nimona.io/pkg/http/router"
)

func (api *API) TokenAuth() router.Handler {
	return func(c *router.Context) {
		if api.token != c.Request.Header.Get("Authorization") {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
	}
}
