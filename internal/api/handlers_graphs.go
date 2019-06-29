package api

import (
	"net/http"
	"strconv"
	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/http/router"
	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
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

	if err := crypto.Sign(o, api.key); err != nil {
		c.AbortWithError(500, errors.New("could not sign object")) // nolint: errcheck
		return
	}

	if err := api.dag.Put(o); err != nil {
		c.AbortWithError(500, errors.New("could not store object")) // nolint: errcheck
		return
	}

	m := api.mapObject(o)
	c.JSON(http.StatusOK, m)
}

func (api *API) HandleGetGraph(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")
	returnDot, _ := strconv.ParseBool(c.Query("dot"))

	if rootObjectHash == "" {
		c.AbortWithError(400, errors.New("missing root object hash")) // nolint: errcheck
		return
	}

	ctx, cf := context.WithTimeout(
		context.New(),
		time.Second*10,
	)
	defer cf()

	logger := log.FromContext(ctx).With(
		log.String("rootObjectHash", rootObjectHash),
	)
	logger.Info("handling request")

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
			addrs = append(addrs, "peer:"+sig.PublicKey.Fingerprint().String())
		}
	}

	// try to sync the graph with the addresses we gathered
	graphObjects, err := api.dag.Sync(ctx, []string{rootObjectHash}, addrs)
	if err != nil {
		if errors.CausedBy(err, graph.ErrNotFound) {
			c.AbortWithError(404, err) // nolint: errcheck
			return
		}
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	if len(graphObjects.Objects) == 0 {
		c.AbortWithError(404, err) // nolint: errcheck
		return
	}

	if returnDot {
		dot, err := graph.Dot(graphObjects.Objects)
		if err != nil {
			c.AbortWithError(500, err) // nolint: errcheck
			return
		}
		c.Header("Content-Type", "text/vnd.graphviz")
		c.Text(http.StatusOK, dot)
		return
	}

	ms := []interface{}{}
	for _, graphObject := range graphObjects.Objects {
		ms = append(ms, api.mapObject(graphObject))
	}
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandlePostGraph(c *router.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
