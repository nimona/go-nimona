package api

import (
	"net/http"

	"nimona.io/internal/http/router"
)

func (api *API) HandleGetIdentities(c *router.Context) {
	// p := api.local.GetPeerPrivateKey()
	// if p.PublicKey.Signature == nil {
	// 	c.JSON(http.StatusNotFound, []interface{}{})
	// 	return
	// }

	// c.JSON(http.StatusOK, []interface{}{
	// 	api.mapObject(p.PublicKey.Signature.PublicKey.ToObject()),
	// })
}

func (api *API) HandleGetIdentity(c *router.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

func (api *API) HandlePostIdentities(c *router.Context) {
	// req := map[string]interface{}{}
	// if err := c.BindBody(req); err != nil {
	// 	c.AbortWithError(400, err) // nolint: errcheck
	// 	return
	// }

	// o := object.FromMap(req)

	// p := &crypto.PrivateKey{}
	// if err := p.FromObject(o); err != nil {
	// 	c.AbortWithError(400, errors.New("invalid private key object")) // nolint: errcheck
	// 	return
	// }

	// if p.Key() == nil {
	// 	c.AbortWithError(400, errors.New("invalid private key")) // nolint: errcheck
	// 	return
	// }

	// if err := api.local.AddIdentityKey(p); err != nil {
	// 	c.AbortWithError(400, errors.New("could not add key")) // nolint: errcheck
	// 	return
	// }

	// c.JSON(http.StatusOK, nil)
}
