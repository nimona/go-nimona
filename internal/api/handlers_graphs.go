package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

func (api *API) HandleGetGraphs(c *gin.Context) {
	// TODO this will be replaced by manager.Subscribe()
	graphRoots, err := api.objectStore.Heads()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ms := []interface{}{}
	for _, graphRoot := range graphRoots {
		ms = append(ms, api.mapObject(graphRoot))
	}
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostGraphs(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}

	o := object.FromMap(req)

	if err := crypto.Sign(o, api.key); err != nil {
		c.AbortWithError(500, errors.New("could not sign object"))
		return
	}

	if err := api.dag.Put(o); err != nil {
		c.AbortWithError(500, errors.New("could not store object"))
		return
	}

	m := api.mapObject(o)
	c.Render(http.StatusOK, Renderer(c, m))
}

func (api *API) HandleGetGraph(c *gin.Context) {
	rootObjectHash := c.Param("rootObjectHash")

	if rootObjectHash == "" {
		c.AbortWithError(400, errors.New("missing root object hash"))
	}

	ctx, cf := context.WithTimeout(
		context.New(),
		time.Second*10,
	)
	defer cf()

	logger := log.Logger(ctx).With(
		zap.String("rootObjectHash", rootObjectHash),
	)
	logger.Info("handling request")

	// find peers who provide the root object
	ps, err := api.discovery.Discover(ctx, &peer.PeerInfoRequest{
		ContentIDs: []string{rootObjectHash},
	})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// convert peer infos to addresses
	if len(ps) == 0 {
		c.AbortWithError(404, err)
		return
	}
	addrs := []string{}
	for _, p := range ps {
		addrs = append(addrs, p.Address())
	}

	// if we have the object, and if its signed, include the signer
	if rootObject, err := api.objectStore.Get(rootObjectHash); err == nil {
		if sk := rootObject.GetSignerKey(); sk != nil {
			addrs = append(addrs, "peer:"+sk.HashBase58())
		}
	}

	// try to sync the graph with the addresses we gathered
	graphObjects, err := api.dag.Sync(ctx, []string{rootObjectHash}, addrs)
	if err != nil {
		if errors.CausedBy(err, graph.ErrNotFound) {
			c.AbortWithError(404, err)
			return
		}
		c.AbortWithError(500, err)
		return
	}

	if len(graphObjects) == 0 {
		c.AbortWithError(404, err)
		return
	}

	ms := []interface{}{}
	for _, graphObject := range graphObjects {
		ms = append(ms, api.mapObject(graphObject))
	}
	c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandlePostGraph(c *gin.Context) {
	c.Render(http.StatusNotImplemented, Renderer(c, nil))
}
