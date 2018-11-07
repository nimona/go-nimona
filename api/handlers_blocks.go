package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"nimona.io/go/primitives"
	"nimona.io/go/storage"
)

type BlockRequest struct {
	Type        string                 `json:"type,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Recipients  []string               `json:"recipients"`
}

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
		m, err := primitives.Unmarshal(b)
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
	m, err := primitives.Unmarshal(b)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := api.mapBlock(m)
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostBlock(c *gin.Context) {
	req := &BlockRequest{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	if len(req.Recipients) == 0 {
		c.AbortWithError(400, errors.New("missing recipients"))
		return
	}

	block := &primitives.Block{
		Type: req.Type,
		Annotations: &primitives.Annotations{
			Policies: []primitives.Policy{
				primitives.Policy{
					Subjects: req.Recipients,
					Actions:  []string{"read"},
					Effect:   "allow",
				},
			},
		},
		Payload: req.Payload,
	}

	signer := api.addressBook.GetLocalPeerKey()
	if err := primitives.Sign(block, signer); err != nil {
		c.AbortWithError(500, err)
		return
	}

	ctx := context.Background()
	for _, recipient := range req.Recipients {
		addr := "peer:" + recipient
		if err := api.exchange.Send(ctx, block, addr); err != nil {
			c.AbortWithError(500, err)
			return
		}
	}

	c.Render(http.StatusOK, Renderer(c, nil))
}
