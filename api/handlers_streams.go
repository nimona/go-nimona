package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"nimona.io/go/log"
	"nimona.io/go/primitives"
)

func (api *API) HandleGetStreams(c *gin.Context) {
	ns := c.Param("ns")
	pattern := c.Param("pattern")

	if pattern != "" {
		pattern = ns + pattern
	} else {
		pattern = ns
	}

	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	ctx := context.Background()
	logger := log.Logger(ctx).Named("api")
	signer := api.addressBook.GetLocalPeerKey()
	incoming := make(chan *primitives.Block, 100)
	outgoing := make(chan *blockReq, 100)

	go func() {
		for {
			select {
			case v := <-incoming:
				m := api.mapBlock(v)
				conn.WriteJSON(m)

			case r := <-outgoing:
				keyBlock, err := primitives.BlockFromBase58(r.Recipient)
				if err != nil {
					return
				}
				key := &primitives.Key{}
				key.FromBlock(keyBlock)
				v := &primitives.Block{
					Type: r.Type,
					Annotations: &primitives.Annotations{
						Policies: []primitives.Policy{
							primitives.Policy{
								Subjects: []string{
									r.Recipient,
								},
								Actions: []string{
									"read",
								},
								Effect: "allow",
							},
						},
					},
					Payload: r.Payload,
				}
				m := api.mapBlock(v)
				m["status"] = "ok"
				if err := api.exchange.Send(ctx, v, key, primitives.SignWith(signer)); err != nil {
					logger.Error("could not send outgoing block", zap.Error(err))
					m["status"] = "error"
				}
				// TODO handle error
				conn.WriteJSON(m)
			}
		}
	}()

	hr, err := api.exchange.Handle(pattern, func(v *primitives.Block) error {
		incoming <- v
		return nil
	})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	defer hr()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Error("could not read from ws", zap.Error(err))
			continue
		}
		r := &blockReq{}
		if err := json.Unmarshal(msg, r); err != nil {
			logger.Error("could not unmarshal outgoing block", zap.Error(err))
			continue
		}
		if r.Type == "" || r.Recipient == "" {
			// TODO send error message to ws
			logger.Error("outgoing block missing type or recipient")
			continue
		}
		outgoing <- r
	}
}
