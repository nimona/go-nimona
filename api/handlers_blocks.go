package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/storage"
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
		m, err := encoding.Unmarshal(b)
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
	m, err := encoding.Unmarshal(b)
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

	// TODO(geoah) better way to require recipients?
	// TODO(geoah) helper function for getting subjects
	subjects := []string{}
	if ps, ok := req["@ann.policy.subjects"]; ok {
		if subs, ok := ps.([]string); ok {
			subjects = subs
		}
	}
	if len(subjects) == 0 {
		c.AbortWithError(400, errors.New("missing recipients"))
		return
	}

	signer := api.addressBook.GetLocalPeerKey()
	o := encoding.NewObjectFromMap(req)
	so, err := crypto.Sign(o, signer)
	if err != nil {
		c.AbortWithError(500, errors.New("could not sign object"))
		return
	}

	ctx := context.Background()
	for _, recipient := range subjects {
		addr := "peer:" + recipient
		if err := api.exchange.Send(ctx, so, addr); err != nil {
			c.AbortWithError(500, err)
			return
		}
	}

	c.Render(http.StatusOK, Renderer(c, nil))
}
