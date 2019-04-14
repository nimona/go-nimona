package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"nimona.io/internal/store/kv"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func (api *API) HandleGetObjects(c *gin.Context) {
	// TODO this will be replaced by manager.Subscribe()
	// objectHashs, err := api.objectStore.List()
	// if err != nil {
	// 	c.AbortWithError(500, err)
	// 	return
	// }
	// ms := []interface{}{}
	// for _, objectHash := range objectHashs {
	// 	b, err := api.objectStore.Get(objectHash)
	// 	if err != nil {
	// 		c.AbortWithError(500, err)
	// 		return
	// 	}
	// 	m, err := object.Unmarshal(b)
	// 	if err != nil {
	// 		c.AbortWithError(500, err)
	// 		return
	// 	}
	// 	ms = append(ms, api.mapObject(m))
	// }
	// c.Render(http.StatusOK, Renderer(c, ms))
	c.Render(http.StatusNotImplemented, nil)
}

func (api *API) HandleGetObject(c *gin.Context) {
	objectHash := c.Param("objectHash")
	if objectHash == "" {
		c.AbortWithError(400, errors.New("missing object hash"))
	}
	o, err := api.objectStore.Get(objectHash)
	if err != nil {
		if err == kv.ErrNotFound {
			c.AbortWithError(404, err)
			return
		}
		c.AbortWithError(500, err)
		return
	}
	ms := api.mapObject(o)
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostObject(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	o := object.FromMap(req)
	op := o.GetPolicy()
	if op == nil {
		c.AbortWithError(400, errors.New("missing policy"))
		return
	}

	p := &object.Policy{}
	if err := p.FromObject(op); err != nil {
		c.AbortWithError(400, errors.New("invalid policy"))
		return
	}

	if len(p.Subjects) == 0 {
		c.AbortWithError(400, errors.New("missing recipients"))
		return
	}

	if err := crypto.Sign(o, api.key); err != nil {
		c.AbortWithError(500, errors.New("could not sign object"))
		return
	}

	ctx := context.Background()
	for _, recipient := range p.Subjects {
		addr := "peer:" + recipient
		if err := api.exchange.Send(ctx, o, addr); err != nil {
			c.AbortWithError(500, err)
			return
		}
	}

	c.Render(http.StatusOK, Renderer(c, nil))
}
