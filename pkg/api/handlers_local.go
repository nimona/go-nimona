package api

import (
	"fmt"
	"net/http"
	"time"

	"nimona.io/internal/http/router"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
)

func (api *API) HandleGetLocal(c *router.Context) {
	p := api.local.GetSignedPeer()
	ms := api.mapObject(p.ToObject())
	ms["_fingerprint"] = p.PublicKey().String()
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandleGetLookup(c *router.Context) {
	opts := []discovery.LookupOption{}
	if v := c.Query("contentType"); v != "" {
		opts = append(opts, discovery.LookupByContentType(v))
	}
	if v := c.Query("certificateSigner"); v != "" {
		cs := crypto.PublicKey(v)
		opts = append(opts, discovery.LookupByCertificateSigner(cs))
	}

	if len(opts) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("missing arguments"))
		return
	}

	ctx := context.New(context.WithTimeout(time.Second * 5))
	ps, err := api.discovery.Lookup(ctx, opts...)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ps)
}

func (api *API) HandleGetDump(c *router.Context) {
	dos, _ := api.objectStore.Filter()
	dot, _ := graph.Dot(dos)
	fmt.Println(dot)
	c.Text(http.StatusOK, dot)
}
