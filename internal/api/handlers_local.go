package api

import (
	"fmt"
	"net/http"

	"nimona.io/internal/http/router"
	"nimona.io/internal/store/graph"
)

func (api *API) HandleGetLocal(c *router.Context) {
	p := api.local.GetSignedPeer()
	ms := api.mapObject(p.ToObject())
	ms["_fingerprint"] = p.Fingerprint().String()
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandleGetDump(c *router.Context) {
	dos, _ := api.objectStore.(*graph.Graph).Dump()
	dot, _ := graph.Dot(dos)
	fmt.Println(dot)
	c.Text(http.StatusOK, dot)
}
