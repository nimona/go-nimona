package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) HandleVersion(c *gin.Context) {
	d := map[string]interface{}{
		"version": api.version,
		"commit":  api.commit,
		"date":    api.buildDate,
	}
	c.Render(http.StatusOK, Renderer(c, d))
}
