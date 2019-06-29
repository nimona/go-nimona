package api

import "nimona.io/internal/http/router"

func (api *API) HandleGetPeers(c *router.Context) {
	// peers, err := api.addressBook.GetAllPeerInfo()
	// if err != nil {
	// 	c.AbortWithError(500, err)
	// 	return
	// }
	// ms := []map[string]interface{}{}
	// for _, v := range peers {
	// 	ms = append(ms, api.mapObject(v.ToObject()))
	// }
	// c.Render(http.StatusOK, Renderer(c, ms))
}

func (api *API) HandleGetPeer(c *router.Context) {
	// fingerprint := c.Param("fingerprint")
	// m, err := api.addressBook.GetPeerInfo(fingerprint)
	// if err != nil {
	// 	c.AbortWithError(500, err)
	// 	return
	// }
	// ms := api.mapObject(m.ToObject())
	// c.Render(http.StatusOK, Renderer(c, ms))
}
