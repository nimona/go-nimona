package api // import "nimona.io/go/api"

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServeFs returns a middleware handler that serves static files from an http fs
func ServeFs(urlPrefix string, fs http.FileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		fileserver.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
