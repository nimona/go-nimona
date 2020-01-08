package api

import (
	"net/http"
	"strconv"
	"time"

	"nimona.io/pkg/http/router"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/dot"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

func (api *API) HandleGetObjects(c *router.Context) {
	// TODO this will be replaced by manager.Subscribe()
	contentTypes := api.local.GetContentTypes()

	ms, err := api.objectStore.Filter(sqlobjectstore.FilterByObjectType(contentTypes...))
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	c.JSON(http.StatusOK, api.mapObjects(ms))
}

func (api *API) HandleGetObject(c *router.Context) {
	objectHash := c.Param("objectHash")
	returnDot, _ := strconv.ParseBool(c.Query("dot"))
	if objectHash == "" {
		c.AbortWithError(400, errors.New("missing object hash")) // nolint: errcheck
	}

	ctx := context.New()
	defer ctx.Cancel()

	h := object.Hash(objectHash)
	ps, err := api.discovery.Lookup(ctx, discovery.LookupByContentHash(h))
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	addrs := []string{}
	for _, p := range ps {
		addrs = append(addrs, p.Address())
	}
	api.orchestrator.Sync(ctx, h, addrs) // nolint: errcheck

	graphObjects, err := api.orchestrator.Get(ctx, h)
	if err != nil {
		if errors.CausedBy(err, sqlobjectstore.ErrNotFound) {
			c.AbortWithError(404, err) // nolint: errcheck
			return
		}
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	os := graphObjects.Objects
	if len(os) == 0 {
		c.AbortWithError(404, errors.New("no objects found")) // nolint: errcheck
		return
	}

	if returnDot {
		dot, err := dot.Dot(os)
		if err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
		c.Header("Content-Type", "text/vnd.graphviz")
		c.Text(http.StatusOK, dot)
		return
	}

	ms := []interface{}{}
	for _, graphObject := range os {
		ms = append(ms, api.mapObject(graphObject))
	}
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandlePostObjects(c *router.Context) {
	req := map[string]interface{}{}
	if err := c.BindBody(&req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	k := api.local.GetPeerPrivateKey()

	req["@identity:s"] = api.local.GetIdentityPublicKey().String()

	o := object.FromMap(req)
	op := stream.Policies(o)
	if len(op) == 0 {
		c.AbortWithError(400, errors.New("missing policy")) // nolint: errcheck
		return
	}

	if err := crypto.Sign(o, k); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	if err := api.orchestrator.Put(o); err != nil {
		c.AbortWithError(500, errors.Wrap(err, errors.New("could not store object"))) // nolint: errcheck
		return
	}

	ctx := context.New(context.WithTimeout(time.Second))
	api.syncOut(ctx, o) // nolint: errcheck

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

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}

func (api *API) HandlePostObject(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")

	req := map[string]interface{}{}
	if err := c.BindBody(&req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	// Get all the objects for a stream
	objs, err := api.objectStore.Filter(
		sqlobjectstore.FilterByStreamHash(
			object.Hash(rootObjectHash),
		),
	)
	if err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	// Find the leaves
	leaves := stream.GetStreamLeaves(objs)

	parents := []string{}
	for _, l := range leaves {
		parents = append(parents, hash.New(l).String())
	}

	req["stream:s"] = rootObjectHash
	req["parents:as"] = parents
	req["@identity:s"] = api.local.GetIdentityPublicKey().String()

	o := object.FromMap(req)

	if err := crypto.Sign(o, api.local.GetPeerPrivateKey()); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	ctx := context.New(context.WithTimeout(time.Second))
	api.syncOut(ctx, o) // nolint: errcheck

	if err := api.orchestrator.Put(o); err != nil {
		c.AbortWithError(500, errors.Wrap(err, errors.New("could not store object"))) // nolint: errcheck
		return
	}

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}

func (api *API) syncOut(ctx context.Context, o object.Object) error {
	if o.GetType() == "" {
		return nil
	}

	id, ok := o.Get("@identity:s").(string)
	if !ok || id == "" {
		return nil
	}

	opts := []discovery.LookupOption{
		discovery.LookupByContentType(o.GetType()),
		discovery.LookupByCertificateSigner(crypto.PublicKey(id)),
	}

	ps, err := api.discovery.Lookup(ctx, opts...)
	if err != nil {
		return err
	}

	for _, p := range ps {
		// nolint: errcheck
		api.exchange.Send(ctx, o, p.Address(), exchange.WithAsync())
	}

	return nil
}
