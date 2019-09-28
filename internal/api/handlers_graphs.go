package api

import (
	"net/http"
	"strconv"
	"time"

	"nimona.io/internal/http/router"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
)

func (api *API) HandleGetGraphs(c *router.Context) {
	// TODO this will be replaced by manager.Subscribe()
	graphRoots, err := api.objectStore.Heads()
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	ms := []interface{}{}
	for _, graphRoot := range graphRoots {
		ms = append(ms, api.mapObject(graphRoot))
	}
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandlePostGraphs(c *router.Context) {
	req := map[string]interface{}{}
	if err := c.BindBody(&req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	o := object.FromMap(req)

	if err := crypto.Sign(o, api.local.GetPeerKey()); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	if err := api.orchestrator.Put(o); err != nil {
		c.AbortWithError(500, errors.Wrap(err, errors.New("could not store object"))) // nolint: errcheck
		return
	}

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}

func (api *API) HandleGetGraph(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")
	returnDot, _ := strconv.ParseBool(c.Query("dot"))
	sync, _ := strconv.ParseBool(c.Query("sync"))

	if rootObjectHash == "" {
		c.AbortWithError(400, errors.New("missing root object hash")) // nolint: errcheck
		return
	}

	ctx := context.New(context.WithTimeout(time.Second * 10))
	defer ctx.Cancel()

	logger := log.FromContext(ctx).With(
		log.String("rootObjectHash", rootObjectHash),
	)
	logger.Info("handling request")

	// os := []object.Object{}

	if sync {
		// find peers who provide the root object
		ps, err := api.discovery.FindByContent(ctx, rootObjectHash)
		if err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}

		// convert peer infos to addresses
		if len(ps) == 0 {
			c.AbortWithError(404, err) // nolint: errcheck
			return
		}

		addrs := []string{}
		for _, p := range ps {
			addrs = append(addrs, p.Address())
		}

		// if we have the object, and if its signed, include the signer
		if rootObject, err := api.objectStore.Get(rootObjectHash); err == nil {
			sig, err := crypto.GetObjectSignature(rootObject)
			if err == nil {
				addrs = append(addrs, sig.PublicKey.Fingerprint().Address())
			}
		}

		// try to sync the graph with the addresses we gathered
		hs := []string{rootObjectHash}
		if _, err = api.orchestrator.Sync(ctx, hs, addrs); err != nil {
			if errors.CausedBy(err, graph.ErrNotFound) {
				c.AbortWithError(404, err) // nolint: errcheck
				return
			}
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}

		// os = graphObjects.Objects
	}

	graphObjects, err := api.orchestrator.Get(ctx, rootObjectHash)
	if err != nil {
		if errors.CausedBy(err, graph.ErrNotFound) {
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
		dot, err := graph.Dot(os)
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

func (api *API) HandlePostGraph(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")

	req := map[string]interface{}{}
	if err := c.BindBody(&req); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	ls, err := api.objectStore.Tails(rootObjectHash)
	if err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	parents := []string{}
	for _, l := range ls {
		parents = append(parents, l.Hash().String())
	}

	req["@root:s"] = rootObjectHash
	req["@parents:as"] = parents

	o := object.FromMap(req)

	if err := crypto.Sign(o, api.local.GetPeerKey()); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	if err := api.orchestrator.Put(o); err != nil {
		c.AbortWithError(500, errors.Wrap(err, errors.New("could not store object"))) // nolint: errcheck
		return
	}

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}
