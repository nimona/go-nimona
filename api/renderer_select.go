package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/render"
)

// Renderer chooses the correct renderer based on the
// contet-type header
func Renderer(c *gin.Context, data interface{}) render.Render {
	contentType := c.ContentType()
	if strings.Contains(contentType, "cbor") {
		return &Cbor{
			Data: data,
		}
	}

	return &render.JSON{
		Data: data,
	}
}
