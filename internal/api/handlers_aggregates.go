package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
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

	ctx, cf := context.WithTimeout(
		context.New(
			context.WithMethod("HandleGetAggregate"),
			context.WithArgument("rootObjectHash", rootObjectHash),
		),
		time.Second*3,
	)
	defer cf()

	logger := log.Logger(ctx).With(
		zap.String("rootObjectHash", rootObjectHash),
	)
	logger.Info("handling request")

	// find peers who provide the root object
	ps, err := api.discovery.FindByContent(ctx, rootObjectHash)
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

	// if we have the object, and if it's signed, include the signer
	if rootObject, err := api.objectStore.Get(rootObjectHash); err == nil {
		sig, err := crypto.GetObjectSignature(rootObject)
		if err == nil {
			addrs = append(addrs, "peer:"+sig.PublicKey.Fingerprint())
		}
	}

	// try to sync the graph with the addresses we gathered
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
