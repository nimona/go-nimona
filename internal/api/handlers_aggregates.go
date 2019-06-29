package api

import (
	"net/http"
	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/http/router"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object/mutation"
)

func (api *API) HandleGetAggregates(c *router.Context) {
	// TODO this will be replaced by manager.Subscribe()
	graphRoots, err := api.objectStore.Heads()
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}
	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	aos := []interface{}{}
	for _, graphRoot := range graphRoots {
		ao, err := api.agg.Get(ctx, graphRoot.HashBase58())
		if err != nil {
			c.AbortWithError(500, errors.Wrap( // nolint: errcheck
				errors.Error("could not get aggregates"),
				err,
			))
			return
		}
		aos = append(aos, api.mapObject(ao.Aggregate))
	}
	c.JSON(http.StatusOK, aos)
}

func (api *API) HandlePostAggregates(c *router.Context) {
	api.HandlePostGraphs(c)
}

func (api *API) HandleGetAggregate(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")

	if rootObjectHash == "" {
		c.AbortWithError(400, errors.Error("missing root object hash")) // nolint: errcheck
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

	// if we have the object, and if it's signed, include the signer
	if rootObject, err := api.objectStore.Get(rootObjectHash); err == nil {
		sig, err := crypto.GetObjectSignature(rootObject)
		if err == nil {
			addrs = append(addrs, "peer:"+sig.PublicKey.Fingerprint().String())
		}
	}

	// try to sync the graph with the addresses we gathered
	_, err = api.dag.Sync(ctx, []string{rootObjectHash}, addrs)
	if err != nil {
		c.AbortWithError(500, errors.Wrap( // nolint: errcheck
			errors.Error("could not sync"),
			err,
		))
		return
	}

	ao, err := api.agg.Get(ctx, rootObjectHash)
	if err != nil {
		c.AbortWithError(500, errors.Wrap( // nolint: errcheck
			errors.Error("could not get aggregates"),
			err,
		))
		return
	}

	c.JSON(http.StatusOK, api.mapObject(ao.Aggregate))
}

func (api *API) HandlePostAggregate(c *router.Context) {
	rootObjectHash := c.Param("rootObjectHash")
	op := &mutation.Operation{}
	if err := c.BindBody(op); err != nil {
		c.AbortWithError(400, err) // nolint: errcheck
		return
	}

	if err := api.agg.Append(rootObjectHash, op); err != nil {
		c.AbortWithError(500, errors.Wrap( // nolint: errcheck
			errors.Error("could append operation"),
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, nil)
}
