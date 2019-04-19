package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object/mutation"
)

func (api *API) HandleGetAggregates(c *gin.Context) {
	// TODO this will be replaced by manager.Subscribe()
	graphRoots, err := api.objectStore.Heads()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	aos := []interface{}{}
	for _, graphRoot := range graphRoots {
		ao, err := api.agg.Get(ctx, graphRoot.HashBase58())
		if err != nil {
			c.AbortWithError(500, errors.Wrap(
				errors.Error("could not get aggregates"),
				err,
			))
			return
		}
		aos = append(aos, api.mapObject(ao.Aggregate))
	}
	c.Render(http.StatusOK, Renderer(c, aos))
}

func (api *API) HandlePostAggregates(c *gin.Context) {
	api.HandlePostGraphs(c)
}

func (api *API) HandleGetAggregate(c *gin.Context) {
	rootObjectHash := c.Param("rootObjectHash")

	if rootObjectHash == "" {
		c.AbortWithError(400, errors.Error("missing root object hash"))
		return
	}

	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()

	ps, err := api.discovery.Discover(&peer.PeerInfoRequest{
		ContentIDs: []string{rootObjectHash},
	})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if len(ps) == 0 {
		c.AbortWithError(404, err)
		return
	}
	addrs := []string{}
	for _, p := range ps {
		addrs = append(addrs, p.Address())
	}

	// TODO check err, properly timeout ctx
	api.dag.Sync(ctx, []string{rootObjectHash}, addrs)

	ao, err := api.agg.Get(ctx, rootObjectHash)
	if err != nil {
		c.AbortWithError(500, errors.Wrap(
			errors.Error("could not get aggregates"),
			err,
		))
		return
	}

	c.Render(http.StatusOK, Renderer(c, api.mapObject(ao.Aggregate)))
}

func (api *API) HandlePostAggregate(c *gin.Context) {
	rootObjectHash := c.Param("rootObjectHash")
	op := &mutation.Operation{}
	if err := c.BindJSON(op); err != nil {
		c.AbortWithError(400, err)
		return
	}

	if err := api.agg.Append(rootObjectHash, op); err != nil {
		c.AbortWithError(500, errors.Wrap(
			errors.Error("could append operation"),
			err,
		))
		return
	}

	c.Render(http.StatusCreated, Renderer(c, nil))
}
