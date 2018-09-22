package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/dht"
	"nimona.io/go/log"
	nnet "nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/storage"
)

// API for HTTP
type API struct {
	router *gin.Engine
}

type blockReq struct {
	Type        string                 `json:"type,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Recipient   string                 `json:"recipient"`
}

// New HTTP API
func New(addressBook *peers.AddressBook, dht *dht.DHT, exchange nnet.Exchange, bls storage.Storage) *API {
	router := gin.Default()
	router.Use(cors.Default())

	local := router.Group("/api/v1/local")
	local.GET("/", func(c *gin.Context) {
		v := addressBook.GetLocalPeerInfo()
		m, err := mapTyped(v)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, m)
	})

	peers := router.Group("/api/v1/peers")
	peers.GET("/", func(c *gin.Context) {
		peers, err := addressBook.GetAllPeerInfo()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		ms := []map[string]interface{}{}
		for _, v := range peers {
			m, err := mapTyped(v)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			ms = append(ms, m)
		}
		c.JSON(http.StatusOK, ms)
	})

	providers := router.Group("/api/v1/providers")
	providers.GET("/", func(c *gin.Context) {
		providers, err := dht.GetAllProviders()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, providers)
	})

	blocksEnd := router.Group("/api/v1/blocks")
	blocksEnd.GET("/", func(c *gin.Context) {
		blockIDs, err := bls.List()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		ms := []map[string]interface{}{}
		for _, blockID := range blockIDs {
			block, err := bls.Get(blockID)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			v, err := blocks.UnpackDecode(block)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			m, err := mapTyped(v)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			ms = append(ms, m)
		}
		c.JSON(http.StatusOK, ms)
	})
	blocksEnd.GET("/:blockID", func(c *gin.Context) {
		blockID := c.Param("blockID")
		block, err := bls.Get(blockID)
		if err != nil {
			if err == storage.ErrNotFound {
				c.AbortWithError(404, err)
				return
			}
			c.AbortWithError(500, err)
			return
		}
		v, err := blocks.UnpackDecode(block)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		m, err := mapTyped(v)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, m)
	})
	blocksEnd.POST("/", func(c *gin.Context) {
		body := &blockReq{}
		if err := c.BindJSON(body); err != nil {
			c.AbortWithError(400, err)
			return
		}

		if body.Recipient == "" {
			c.AbortWithError(400, errors.New("missing recipient"))
			return
		}

		typedKey, err := blocks.UnpackDecodeBase58(body.Recipient)
		if err != nil {
			c.AbortWithError(400, errors.New("invalid recipient"))
			return
		}

		recipientKey := typedKey.(*crypto.Key)

		unknown := &blocks.Unknown{
			Type:        body.Type,
			Payload:     body.Payload,
			Annotations: body.Annotations,
		}
		ctx := context.Background()
		signer := addressBook.GetLocalPeerKey()
		if err := exchange.Send(ctx, unknown, recipientKey, blocks.SignWith(signer)); err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	streamsEnd := router.Group("/api/v1/streams")
	streamsEnd.GET("/:pattern", func(c *gin.Context) {
		pattern := c.Param("pattern")

		var wsupgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		ctx := context.Background()
		logger := log.Logger(ctx).Named("api")
		signer := addressBook.GetLocalPeerKey()
		incoming := make(chan blocks.Typed, 100)
		outgoing := make(chan *blockReq, 100)

		go func() {
			for {
				select {
				case v := <-incoming:
					m, err := mapTyped(v)
					if err != nil {
						// TODO handle error
						continue
					}
					// TODO handle error
					conn.WriteJSON(m)

				case r := <-outgoing:
					k, err := blocks.UnpackDecodeBase58(r.Recipient)
					if err != nil {
						// TODO send error message to ws
						logger.Error("could not parse outgoing block", zap.Error(err))
						continue
					}
					v := &blocks.Unknown{
						Type:        r.Type,
						Annotations: r.Annotations,
						Payload:     r.Payload,
					}
					if err := exchange.Send(ctx, v, k.(*crypto.Key), blocks.SignWith(signer)); err != nil {
						logger.Error("could not send outgoing block", zap.Error(err))
						// TODO send error message to ws
						continue
					}
				}
			}
		}()

		hr, err := exchange.Handle(pattern, func(v blocks.Typed) error {
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
				logger.Error("could not read outgoing block", zap.Error(err))
				continue
			}
			fmt.Println("got", string(msg))
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
	})

	router.LoadHTMLFiles("./api/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	return &API{
		router: router,
	}
}

// Serve HTTP API
func (api *API) Serve(address string) error {
	return api.router.Run(address)
}

func mapTyped(v blocks.Typed) (map[string]interface{}, error) {
	m, err := blocks.MapTyped(v)
	if err != nil {
		return nil, err
	}
	m["owner"] = v.GetSignature().Key.Thumbprint()
	m["id"] = blocks.ID(v)
	delete(m, "signature")
	return m, nil
}
