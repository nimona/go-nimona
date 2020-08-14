package api

import (
	"net/http"

	"nimona.io/pkg/http/router"
)

func (api *API) HandleGetIdentities(c *router.Context) {
	p := api.keychain.GetPrimaryIdentityKey().PublicKey()
	c.JSON(http.StatusOK, p)
}

func (api *API) HandleGetIdentity(c *router.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

func (api *API) HandlePostIdentities(c *router.Context) {
	// TODO implement
	c.JSON(http.StatusNotImplemented, nil)
}
