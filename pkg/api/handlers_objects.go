package api

import (
	"net/http"
	"strconv"
	"time"

	"nimona.io/pkg/peer"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/dot"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/http/router"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
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

	ctx := context.New(
		context.WithTimeout(15 * time.Second),
	)
	defer ctx.Cancel()

	h := object.Hash(objectHash)
	ps, err := api.discovery.Lookup(ctx, peer.LookupByContentHash(h))
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	keys := []crypto.PublicKey{}
	for _, p := range gatherPeers(ps) {
		keys = append(keys, p.PublicKey())
	}
	api.orchestrator.Sync(ctx, h, peer.LookupByOwner(keys...)) // nolint: errcheck

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

	o := object.FromMap(req)
	k := api.local.GetPeerPrivateKey()
	o = o.SetOwners([]crypto.PublicKey{
		api.local.GetIdentityPublicKey(),
	})

	sig, err := object.NewSignature(k, o)
	if err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	o = o.SetSignature(sig)
	if err := api.objectStore.Put(o); err != nil {
		c.AbortWithError(500, errors.Wrap(err, errors.New("could not store object"))) // nolint: errcheck
		return
	}

	ctx := context.New(context.WithTimeout(time.Second))
	api.syncOut(ctx, o) // nolint: errcheck

	p := o.GetPolicy()

	for i, s := range p.Subjects {
		go func(i int, s string) {
			ctx := context.New(
				context.WithCorrelationID("XPOST" + strconv.Itoa(i)),
			)
			err := api.exchange.Send(ctx, o, peer.LookupByOwner(crypto.PublicKey(s)), exchange.WithAsync())
			if err != nil {
				logger := log.FromContext(ctx)
				logger.Error("could not send to peer", log.String("s", s), log.Error(err))
			}
		}(i, s)
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
		parents = append(parents, object.NewHash(l).String())
	}

	req["stream:s"] = rootObjectHash
	req["parents:as"] = parents
	req["@owners:as"] = []interface{}{api.local.GetIdentityPublicKey().String()}

	o := object.FromMap(req)

	sig, err := object.NewSignature(api.local.GetPeerPrivateKey(), o)
	if err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	o = o.SetSignature(sig)
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

	owners := o.GetOwners()
	if len(owners) == 0 {
		return nil
	}

	opts := []peer.LookupOption{
		peer.LookupByContentType(o.GetType()),
		peer.LookupByCertificateSigner(crypto.PublicKey(owners[0])),
	}

	ps, err := api.discovery.Lookup(ctx, opts...)
	if err != nil {
		return err
	}

	for _, p := range gatherPeers(ps) {
		// nolint: errcheck
		api.exchange.Send(ctx, o, peer.LookupByOwner(p.PublicKey()), exchange.WithAsync())
	}

	return nil
}

func gatherPeers(p <-chan *peer.Peer) []*peer.Peer {
	ps := []*peer.Peer{}
	for p := range p {
		p := p
		ps = append(ps, p)
	}
	return peer.Unique(ps)
}
