package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if api.token != c.Request.Header.Get("Authorization") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
