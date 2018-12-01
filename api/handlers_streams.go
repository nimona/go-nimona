package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/log"
)

func (api *API) HandleGetStreams(c *gin.Context) {
	ns := c.Param("ns")
	pattern := c.Param("pattern")

	if pattern != "" {
		pattern = ns + pattern
	} else {
		pattern = ns
	}

	write := func(conn *websocket.Conn, data interface{}) error {
		contentType := strings.ToLower(c.ContentType())
		if strings.Contains(contentType, "cbor") {
			bs, err := encoding.MarshalSimple(data)
			if err != nil {
				return err
			}
			if err := conn.WriteMessage(2, bs); err != nil {
				return err
			}
		}

		return conn.WriteJSON(data)
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
	incoming := make(chan *encoding.Object, 100)
	outgoing := make(chan *encoding.Object, 100)

	go func() {
		for {
			select {
			case v := <-incoming:
				m := api.mapBlock(v)
				write(conn, m)

			case req := <-outgoing:
				singedReq, err := crypto.Sign(req, signer)
				if err != nil {
					logger.Error("could not sign outgoing block", zap.Error(err))
					// resp["status"] = "error signing block"
					// TODO handle error
					write(conn, req)
					continue
				}
				// TODO(geoah) better way to require recipients?
				// TODO(geoah) helper function for getting subjects
				subjects := []string{}
				// if ps, ok := req["@ann.policy.subjects"]; ok {
				// 	if subs, ok := ps.([]string); ok {
				// 		subjects = subs
				// 	}
				// }
				if len(subjects) == 0 {
					// TODO handle error
					// req["status"] = "no subjects"
					write(conn, req)
					continue
				}
				for _, recipient := range subjects {
					addr := "peer:" + recipient
					if err := api.exchange.Send(ctx, singedReq, addr); err != nil {
						logger.Error("could not send outgoing block", zap.Error(err))
						// req["status"] = "error sending block"
					}
					// TODO handle error
					write(conn, req)
				}
			}
		}
	}()
	fmt.Println(pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	hr, err := api.exchange.Handle(pattern, func(o *encoding.Object) error {
		incoming <- o
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
			if err == io.EOF {
				logger.Debug("ws conn is dead", zap.Error(err))
				return
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Debug("ws conn closed", zap.Error(err))
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				logger.Warn("ws conn closed with unexpected error", zap.Error(err))
				return
			}

			logger.Warn("could not read from ws", zap.Error(err))
			continue
		}
		m := map[string]interface{}{}
		if err := json.Unmarshal(msg, &m); err != nil {
			logger.Error("could not unmarshal outgoing block", zap.Error(err))
			continue
		}
		o := encoding.NewObjectFromMap(m)
		outgoing <- o
	}
}
