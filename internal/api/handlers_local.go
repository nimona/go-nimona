package api

import (
	"fmt"
	"net/http"

	"nimona.io/internal/store/graph"
	"nimona.io/internal/http/router"
)

func (api *API) HandleGetLocal(c *router.Context) {
	ms := api.mapObject(api.local.GetSignedPeer().ToObject())
	c.JSON(http.StatusOK, ms)
}

func (api *API) HandleGetDump(c *router.Context) {
	dos, _ := api.objectStore.(*graph.Graph).Dump()
	dot, _ := graph.Dot(dos)
	fmt.Println(dot)
	c.Text(http.StatusOK, dot)
}
