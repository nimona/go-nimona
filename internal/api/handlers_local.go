package api

import (
	"net/http"

	"nimona.io/pkg/http/router"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func (api *API) HandleGetLocal(c *router.Context) {
	p := &peer.Peer{
		Addresses:    api.network.Addresses(),
		Certificates: api.localpeer.GetCertificates(),
		Metadata: object.Metadata{
			Owner: api.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
	}
	ms := api.mapObject(p.ToObject())
	c.JSON(http.StatusOK, ms)
}
