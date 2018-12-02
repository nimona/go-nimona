package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if api.token != c.Request.Header.Get("token") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
