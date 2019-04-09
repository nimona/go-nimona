package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
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
			bs, err := object.MarshalSimple(data)
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
	incoming := make(chan *object.Object, 100)
	outgoing := make(chan *object.Object, 100)

	var mandateObj *object.Object
	if mandate := api.local.GetPeerInfo().Mandate; mandate != nil {
		mandateObj = mandate.ToObject()
	}

	go func() {
		for {
			select {
			case v := <-incoming:
				m := api.mapObject(v)
				if err := write(conn, m); err != nil {
					// TODO handle error
					continue
				}

			case req := <-outgoing:
				if mandateObj != nil {
					req.SetMandate(mandateObj)
				}
				if err := crypto.Sign(req, api.key); err != nil {
					logger.Error("could not sign outgoing object", zap.Error(err))
					req.SetRaw("_status", "error signing object")
					if err := write(conn, api.mapObject(req)); err != nil {
						// TODO handle error
						continue
					}
				}
				// TODO(geoah) better way to require recipients?
				// TODO(geoah) helper function for getting subjects
				subjects := []string{}
				if ps := req.GetRaw("_recipients"); ps != nil {
					if subsi, ok := ps.([]interface{}); ok {
						for _, subi := range subsi {
							if sub, ok := subi.(string); ok {
								subjects = append(subjects, sub)
							}
						}
					}
				}
				if len(subjects) == 0 {
					// TODO handle error
					req.SetRaw("_status", "no subjects")
					if err := write(conn, api.mapObject(req)); err != nil {
						// TODO handle error
					}
					continue
				}
				for _, recipient := range subjects {
					addr := recipient
					// TODO(geoah) Rephrase ui and api and remove
					if !strings.Contains(addr, ":") {
						addr = "peer:" + recipient
					}
					if err := api.exchange.Send(ctx, req, addr); err != nil {
						logger.Error("could not send outgoing object",
							zap.Error(err),
							zap.String("addr", addr))
						req.SetRaw("_status", "error sending object")
					}
					// TODO handle error
					if err := write(conn, api.mapObject(req)); err != nil {
						// TODO handle error
						continue
					}
				}
			}
		}
	}()
	hr, err := api.exchange.Handle(pattern, func(o *object.Object) error {
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
			logger.Error("could not unmarshal outgoing object", zap.Error(err))
			continue
		}
		o := object.FromMap(m)
		outgoing <- o
	}
}
