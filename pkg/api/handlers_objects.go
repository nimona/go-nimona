package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/dot"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/http/router"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func (api *API) HandleGetObjects(c *router.Context) {
	filters := []sqlobjectstore.LookupOption{}
	contentTypes := strings.Split(c.Query("type"), ",")
	for _, ct := range contentTypes {
		ct = strings.TrimSpace(ct)
		if ct == "" {
			continue
		}
		filters = append(filters, sqlobjectstore.FilterByObjectType(ct))
	}
	ms, err := api.objectStore.Filter(filters...)
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
	ps, err := api.resolver.Lookup(ctx, resolver.LookupByContentHash(h))
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	keys := []crypto.PublicKey{}
	for _, p := range gatherPeers(ps) {
		keys = append(keys, p.PublicKey())
	}
	nps, err := api.resolver.Lookup(ctx, resolver.LookupByOwner(keys...))
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	for p := range nps {
		go api.streammanager.Sync(ctx, h, p) // nolint: errcheck
	}

	graphObjects, err := api.streammanager.Get(ctx, h)
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
		d, err := dot.Dot(os)
		if err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
		c.Header("Content-Type", "text/vnd.graphviz")
		c.Text(http.StatusOK, d)
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
	k := api.keychain.GetPrimaryPeerKey()
	o = o.SetOwners(
		api.keychain.ListPublicKeys(keychain.IdentityKey),
	)

	sig, err := object.NewSignature(k, o)
	if err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.New("could not sign object"),
		)
		return
	}

	o = o.AddSignature(sig)
	if err := api.objectStore.Put(o); err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.Wrap(err, errors.New("could not store object")),
		)
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
			ps, err := api.resolver.Lookup(
				ctx,
				resolver.LookupByOwner(crypto.PublicKey(s)),
			)
			if err != nil {
				return
			}
			for p := range ps {
				err := api.exchange.Send(
					ctx,
					o,
					p,
				)
				if err != nil {
					logger := log.FromContext(ctx)
					logger.Error(
						"could not send to peer",
						log.String("s", s),
						log.Error(err),
					)
				}
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
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.New("could not sign object"),
		)
		return
	}

	// Find the leaves
	leaves := stream.GetStreamLeaves(objs)

	parents := []string{}
	for _, l := range leaves {
		parents = append(parents, l.Hash().String())
	}

	req["stream:s"] = rootObjectHash
	req["parents:as"] = parents
	req["@owners:as"] = []interface{}{
		api.keychain.List(keychain.IdentityKey)[0],
	}

	o := object.FromMap(req)

	sig, err := object.NewSignature(
		api.keychain.GetPrimaryPeerKey(),
		o,
	)
	if err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.New("could not sign object"),
		)
		return
	}

	o = o.AddSignature(sig)
	ctx := context.New(context.WithTimeout(time.Second))
	api.syncOut(ctx, o) // nolint: errcheck

	if err := api.streammanager.Put(o); err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.Wrap(err, errors.New("could not store object")),
		)
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

	opts := []resolver.LookupOption{
		resolver.LookupByContentType(o.GetType()),
		resolver.LookupByCertificateSigner(owners[0]),
	}

	ps, err := api.resolver.Lookup(ctx, opts...)
	if err != nil {
		return err
	}

	for p := range ps {
		// nolint: errcheck
		go api.exchange.Send(
			ctx,
			o,
			p,
		)
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
