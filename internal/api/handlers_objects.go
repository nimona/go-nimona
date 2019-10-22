package api

import (
	"net/http"
	"strconv"

	"nimona.io/internal/http/router"
	"nimona.io/internal/store/kv"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

func (api *API) HandleGetObjects(c *router.Context) {
	// TODO this will be replaced by manager.Subscribe()
	ms, err := api.objectStore.Heads()
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
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
	c.JSON(http.StatusOK, api.mapObjects(ms))
	// c.JSON(http.StatusNotImplemented, nil)
}

func (api *API) HandleGetObject(c *router.Context) {
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
		c.JSON(http.StatusOK, ms)
		return
	}

	ctx := context.New()
	defer ctx.Cancel()

	h, _ := object.HashFromCompact(objectHash)
	ps, err := api.discovery.FindByContent(ctx, h)
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	addrs := []string{}
	for _, p := range ps {
		addrs = append(addrs, p.Address())
	}
	os, err := api.orchestrator.Sync(ctx, h, addrs)
	if err != nil {
		if err == kv.ErrNotFound {
			c.AbortWithError(404, err) // nolint: errcheck
			return
		}
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	if len(os.Objects) == 0 {
		c.AbortWithError(404, err) // nolint: errcheck
		return
	}
	ms := api.mapObject(os.Objects[0])
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandlePostObject(c *router.Context) {
	req := map[string]interface{}{}
	if err := c.BindBody(&req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	o := object.FromMap(req)
	op := stream.Policies(o)
	if len(op) == 0 {
		c.AbortWithError(400, errors.New("missing policy")) // nolint: errcheck
		return
	}

	if err := crypto.Sign(o, api.local.GetPeerPrivateKey()); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	for _, p := range op {
		for i, s := range p.Subjects {
			go func(i int, s string) {
				ctx := context.New(
					context.WithCorrelationID("XPOST" + strconv.Itoa(i)),
				)
				err := api.exchange.Send(ctx, o, "peer:"+s)
				if err != nil {
					logger := log.FromContext(ctx)
					logger.Error("could not send to peer", log.String("s", s), log.Error(err))
				}
			}(i, s)
		}
	}

	c.JSON(http.StatusOK, nil)
}
