package api

import (
	"net/http"

	"nimona.io/pkg/http/router"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func (api *API) HandleGetLocal(c *router.Context) {
	p := &peer.Peer{
		Addresses: api.network.Addresses(),
		Certificates: api.localpeer.GetCertificates(
			api.localpeer.GetPrimaryPeerKey().PublicKey(),
		),
		Metadata: object.Metadata{
			Owner: api.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
	}
	ms := api.mapObject(p.ToObject())
	c.JSON(http.StatusOK, ms)
}

// func (api *API) HandleGetLookup(c *router.Context) {
// 	opts := []peer.LookupOption{}
// 	if v := c.Query("contentType"); v != "" {
// 		opts = append(opts, peer.LookupByContentType(v))
// 	}
// 	if v := c.Query("certificateSigner"); v != "" {
// 		cs := crypto.PublicKey(v)
// 		opts = append(opts, peer.LookupByCertificateSigner(cs))
// 	}

// 	if len(opts) == 0 {
// 		c.AbortWithError(http.StatusBadRequest, errors.New("missing arguments"))
// 		return
// 	}

// 	ctx := context.New(context.WithTimeout(time.Second * 5))
// 	ps, err := api.discovery.Lookup(ctx, opts...)
// 	if err != nil {
// 		c.AbortWithError(http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, ps)
// }
