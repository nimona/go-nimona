package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/store/kv"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func (api *API) HandleGetObjects(c *gin.Context) {
	// TODO this will be replaced by manager.Subscribe()
	// objectHashs, err := api.objectStore.List()
	// if err != nil {
	// 	c.AbortWithError(500, err) // nolint: errcheck
	// 	return
	// }
	// ms := []interface{}{}
	// for _, objectHash := range objectHashs {
	// 	b, err := api.objectStore.Get(objectHash)
	// 	if err != nil {
	// 		c.AbortWithError(500, err) // nolint: errcheck
	// 		return
	// 	}
	// 	m, err := object.Unmarshal(b)
	// 	if err != nil {
	// 		c.AbortWithError(500, err) // nolint: errcheck
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
		c.AbortWithError(400, errors.New("missing object hash")) // nolint: errcheck
	}
	o, err := api.objectStore.Get(objectHash)
	if err != nil && err != kv.ErrNotFound {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	} else if err == nil {
		ms := api.mapObject(o)
		c.Render(http.StatusOK, Renderer(c, ms))
		return
	}

	ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
	defer cf()
	ps, err := api.discovery.FindByContent(ctx, objectHash)
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	addrs := []string{}
	for _, p := range ps {
		addrs = append(addrs, p.Address())
	}
	os, err := api.dag.Sync(ctx, []string{objectHash}, addrs)
	if err != nil {
		if err == kv.ErrNotFound {
			c.AbortWithError(404, err) // nolint: errcheck
			return
		}
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	if len(os) == 0 {
		c.AbortWithError(404, err) // nolint: errcheck
		return
	}
	ms := api.mapObject(os[0])
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostObject(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	o := object.FromMap(req)
	op := o.GetPolicy()
	if op == nil {
		c.AbortWithError(400, errors.New("missing policy")) // nolint: errcheck
		return
	}

	p := &object.Policy{}
	if err := p.FromObject(op); err != nil {
		c.AbortWithError(400, errors.New("invalid policy")) // nolint: errcheck
		return
	}

	if len(p.Subjects) == 0 {
		c.AbortWithError(400, errors.New("missing recipients")) // nolint: errcheck
		return
	}

	if err := crypto.Sign(o, api.key); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	ctx := context.Background()
	for _, recipient := range p.Subjects {
		addr := "peer:" + recipient
		if err := api.exchange.Send(ctx, o, addr); err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
	}

	c.Render(http.StatusOK, Renderer(c, nil))
}
