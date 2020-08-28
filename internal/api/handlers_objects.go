package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"nimona.io/pkg/objectmanager"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/dot"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/http/router"
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
	or, err := api.objectStore.Filter(filters...)
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	ms, err := object.ReadAll(or)
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
	os := []object.Object{}
	for p := range nps {
		os = []object.Object{}
		res, err := api.objectmanager.RequestStream(ctx, h, p)
		if err != nil {
			continue
		}
		done := false
		for {
			obj, err := res.Read()
			if err == objectmanager.ErrDone {
				done = true
				break
			}
			os = append(os, *obj)
		}
		if done {
			break
		}
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
	k := api.localpeer.GetPrimaryPeerKey()
	o = o.SetOwner(
		api.localpeer.GetPrimaryIdentityKey().PublicKey(),
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

	o = o.SetSignature(sig)
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
	or, err := api.objectStore.Filter(
		sqlobjectstore.FilterByStreamHash(
			object.Hash(rootObjectHash),
		),
	)
	if err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.New("could not filter objects"),
		)
		return
	}
	objs, err := object.ReadAll(or)
	if err != nil {
		// nolint: errcheck
		c.AbortWithError(
			500,
			errors.New("could not read all objects"),
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
	req["owner:s"] = api.localpeer.GetPrimaryIdentityKey().PublicKey()

	o := object.FromMap(req)

	sig, err := object.NewSignature(
		api.localpeer.GetPrimaryPeerKey(),
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

	o = o.SetSignature(sig)
	ctx := context.New(context.WithTimeout(time.Second))
	api.syncOut(ctx, o) // nolint: errcheck

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}

func (api *API) syncOut(ctx context.Context, o object.Object) error {
	if o.GetType() == "" {
		return nil
	}

	owner := o.GetOwner()
	if owner.IsEmpty() {
		return nil
	}

	opts := []resolver.LookupOption{
		resolver.LookupByContentType(o.GetType()),
		resolver.LookupByCertificateSigner(owner),
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
