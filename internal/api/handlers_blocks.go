package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/storage"
)

func (api *API) HandleGetBlocks(c *gin.Context) {
	blockIDs, err := api.blockStore.List()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := []interface{}{}
	for _, blockID := range blockIDs {
		b, err := api.blockStore.Get(blockID)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		m, err := object.Unmarshal(b)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		ms = append(ms, api.mapBlock(m))
	}
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandleGetBlock(c *gin.Context) {
	blockID := c.Param("blockID")
	if blockID == "" {
		c.AbortWithError(400, errors.New("missing block id"))
	}
	b, err := api.blockStore.Get(blockID)
	if err != nil {
		if err == storage.ErrNotFound {
			c.AbortWithError(404, err)
			return
		}
		c.AbortWithError(500, err)
		return
	}
	m, err := object.Unmarshal(b)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := api.mapBlock(m)
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostBlock(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	o := object.NewObjectFromMap(req)
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
