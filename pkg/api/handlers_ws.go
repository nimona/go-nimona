package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tsdtsdtsd/identicon"
	"github.com/tyler-smith/go-bip39"

	"nimona.io/internal/app/identity"
	"nimona.io/internal/http/router"
	"nimona.io/internal/store/sql"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
)

type CreateProfileRequest struct {
	NameFirst      string `json:"nameFirst"`
	NameLast       string `json:"nameLast"`
	DisplayPicture []byte `json:"displayPicture"`
}

type RegisterRequest struct {
	Seed string `json:"privateKeySeed"`
}

type CertificateResponse struct {
	Type                  string `json:"@type:s"`
	Alias                 string `json:"alias"`
	ProfileName           string `json:"profileName"`
	ProfileDisplayPicture string `json:"profileDisplayPicture"`
	IdentityPublicKey     string `json:"identityPublicKey"`
	Developer             string `json:"developer"`
}

// type ContactReponse struct {
// 	Type                  string `json:"@type:s"`
// 	Alias                 string `json:"alias"`
// 	ProfileName           string `json:"profileName"`
// 	ProfileDisplayPicture string `json:"profileDisplayPicture"`
// 	IdentityPublicKey     string `json:"identityPublicKey"`
// }

type ContactRequest struct {
	Alias             string `json:"alias"`
	IdentityPublicKey string `json:"identityPublicKey"`
}

type LocalPeerResponse struct {
	Type                 string                `json:"@type:s"`
	Identity             string                `json:"identity:s"`
	Addresses            []string              `json:"addresses:as"`
	ContentTypes         []string              `json:"contentTypes:as"`
	Bloom                []int64               `json:"bloom:ai"`
	Certificates         []CertificateResponse `json:"certificates:ao"`
	Version              int64                 `json:"version:i"`
	PeerPublicKey        string                `json:"_peerPublicKey:s"`
	PeerPrivateKey       string                `json:"_peerPrivateKey:s"`
	IdentityPublicKey    string                `json:"_identityPublicKey:s"`
	IdentityPrivateKey   string                `json:"_identityPrivateKey:s"`
	IdentitySecretPhrase []string              `json:"_identitySecretPhrase:as"`
	Identicon            []byte                `json:"identicon:d"`
}

type wsConn struct {
	conn *websocket.Conn
	lock *sync.Mutex
}

func write(conn *wsConn, v interface{}) error {
	b, _ := json.Marshal(v)
	fmt.Println("sending", string(b))
	conn.lock.Lock()
	defer conn.lock.Unlock()
	return conn.conn.WriteJSON(v)
}

func getIdenticon(key string) []byte {
	ic, err := identicon.New(key, &identicon.Options{
		BackgroundColor: identicon.RGB(240, 240, 240),
		ImageSize:       500,
	})
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = png.Encode(buf, ic)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
	// return base64.StdEncoding.EncodeToString(b)
}

func (api *API) getContacts(
	conn *wsConn,
) error {
	res, err := api.objectStore.Filter(
		sql.FilterByObjectType("nimona.io/identity.Contact"),
	)
	fmt.Println("found", len(res), "contacts")
	if err != nil {
		return err
	}
	for _, o := range res {
		write(conn, o) // nolint: errcheck
	}

	// // HACK temp profiles
	// ps := []identity.Profile{
	// 	{
	// 		NameFirst:      "Paris",
	// 		NameLast:       "Goudas",
	// 		Identity:       crypto.PublicKey("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngvX"),
	// 		DisplayPicture: getIdenticon("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngvX"),
	// 	},
	// 	{
	// 		NameFirst:      "Jim",
	// 		NameLast:       "Myhrberg",
	// 		Identity:       crypto.PublicKey("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngv2"),
	// 		DisplayPicture: getIdenticon("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngv2"),
	// 	},
	// 	{
	// 		NameFirst:      "Alex",
	// 		NameLast:       "Papadopoulos",
	// 		Identity:       crypto.PublicKey("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngv1"),
	// 		DisplayPicture: getIdenticon("ed25519.2Y4MskGhsZTRk5TPP7ZVw14Wgaevei1TRrkrY6dxngv1"),
	// 	},
	// }
	// for _, p := range ps {
	// 	write(conn, p.ToObject().ToMap())
	// }

	return nil
}

func (api *API) subContacts(
	conn *wsConn,
) error {
	res := api.objectStore.Subscribe(
		sql.FilterByObjectType("nimona.io/identity.Contact"),
	)
	defer res.Cancel()
	for {
		o, err := res.Next()
		if err != nil || o == nil {
			return err
		}
		write(conn, o) // nolint: errcheck
	}
	return nil
}

func (api *API) addContact(
	alias string,
	identityPublicKey crypto.PublicKey,
	conn *wsConn,
) error {
	contact := identity.Contact{
		Alias:     alias,
		PublicKey: identityPublicKey,
	}
	fmt.Println(contact)
	signature, err := crypto.NewSignature(
		api.local.GetPeerPrivateKey(),
		contact.ToObject(),
	)
	if err != nil {
		return err
	}
	contact.Signature = signature
	fmt.Println("_PUT", api.objectStore.Put(contact.ToObject()))

	return nil
}

func (api *API) newIdentity(
	conn *wsConn,
) error {
	if !api.local.GetIdentityPublicKey().Equals(api.local.GetPeerPublicKey()) {
		return api.sendLocalPeer(conn)
	}

	fmt.Println("___creating new id")

	identityKey, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		return err
	}

	if err := api.local.AddIdentityKey(identityKey); err != nil {
		return err
	}

	api.config.Peer.IdentityKey = identityKey
	if err := api.config.Update(); err != nil {
		panic(err)
	}

	return nil
}

func (api *API) createProfile(
	conn *wsConn,
	req CreateProfileRequest,
) error {
	profile := identity.Profile{
		NameFirst:      req.NameFirst,
		NameLast:       req.NameLast,
		DisplayPicture: req.DisplayPicture,
		Identity:       api.local.GetIdentityPublicKey(),
	}
	api.objectStore.Put(profile.ToObject()) // nolint: errcheck
	return nil
}

func (api *API) sendLocalPeer(
	conn *wsConn,
) error {
	phrase, err := bip39.NewMnemonic(
		api.local.GetIdentityPrivateKey().Bytes(),
	)
	if err != nil {
		panic(err)
	}
	localPeer := api.local.GetSignedPeer()
	localPeerRes := &LocalPeerResponse{
		Type:                 "mochi.io/local.Peer",
		Identity:             api.local.GetIdentityPublicKey().String(),
		Addresses:            localPeer.Addresses,
		ContentTypes:         localPeer.ContentTypes,
		Bloom:                localPeer.Bloom,
		Certificates:         []CertificateResponse{},
		Version:              localPeer.Version,
		PeerPublicKey:        api.local.GetPeerPublicKey().String(),
		PeerPrivateKey:       api.local.GetPeerPrivateKey().String(),
		IdentityPublicKey:    api.local.GetIdentityPublicKey().String(),
		IdentityPrivateKey:   api.local.GetIdentityPrivateKey().String(),
		IdentitySecretPhrase: strings.Split(phrase, " "),
		Identicon:            getIdenticon(api.local.GetIdentityPublicKey().String()),
	}
	if err := write(conn, localPeerRes); err != nil {
		return err
	}
	return nil
}

func (api *API) HandleWS(c *router.Context) {
	// write := func(conn *wsConn, data interface{}) error {
	// 	return write(conn,data)
	// }

	fmt.Println("new connection")

	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	httpConn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithError(500, err) // nolint: errcheck
		return
	}

	conn := &wsConn{
		conn: httpConn,
		lock: &sync.Mutex{},
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).Named("api")

	// incoming := make(chan object.Object, 100000)
	// outgoing := make(chan object.Object, 100)

	// go func() {
	// 	peer := api.local.GetSignedPeer().ToObject()
	// 	if err := write(conn, api.mapObject(peer)); err != nil {
	// 		fmt.Println("ERRR", err)
	// 	}
	// }()

	// subscribe to contact updates
	go api.subContacts(conn) // nolint: errcheck

	// retrieve existing contacts
	go api.getContacts(conn) // nolint: errcheck

	for {
		_, msg, err := httpConn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				logger.Debug("ws conn is dead", log.Error(err))
				return
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Debug("ws conn closed", log.Error(err))
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				logger.Warn("ws conn closed with unexpected error", log.Error(err))
				return
			}

			logger.Warn("could not read from ws", log.Error(err))
			continue
		}

		fmt.Println("got", string(msg))

		m := map[string]interface{}{}
		if err := json.Unmarshal(msg, &m); err != nil {
			logger.Error("could not unmarshal outgoing object", log.Error(err))
			continue
		}

		switch m["_action"] {
		case "register":
			api.newIdentity(conn)   // nolint: errcheck
			api.sendLocalPeer(conn) // nolint: errcheck
			continue
		case "createProfile":
			req := CreateProfileRequest{}
			json.Unmarshal(msg, &req)    // nolint: errcheck
			api.createProfile(conn, req) // nolint: errcheck
			api.sendLocalPeer(conn)      // nolint: errcheck
			continue
		case "ping":
			api.sendLocalPeer(conn) // nolint: errcheck
			api.getContacts(conn)   // nolint: errcheck
			continue
		}

		if _, ok := m["alias"]; ok {
			req := ContactRequest{}
			json.Unmarshal(msg, &req) // nolint: errcheck
			if req.Alias == "" || req.IdentityPublicKey == "" {
				continue
			}
			api.addContact(req.Alias, crypto.PublicKey(req.IdentityPublicKey), conn) // nolint: errcheck
			continue
		}

		go api.sendLocalPeer(conn) // nolint: errcheck

		// contacts := []ContactReponse{
		// 	ContactReponse{
		// 		Type:                  "nimona.io/contact.Contact",
		// 		Alias:                 "superdecimal",
		// 		ProfileName:           "Paris Goudas",
		// 		IdentityPublicKey:     "AF7nwTFTC4dBGBaXDtdZFUvxR4fofFycDQpR585cnGEM",
		// 		ProfileDisplayPicture: "https://avatars1.githubusercontent.com/u/5281398?s=400&v=4",
		// 	},
		// 	ContactReponse{
		// 		Type:                  "nimona.io/contact.Contact",
		// 		Alias:                 "jimeh",
		// 		ProfileName:           "Jim Myhrberg",
		// 		IdentityPublicKey:     "6NDRgaGK5hk9Aa91WanepaZkGYNtAnZPrcvFfxLBq9oL",
		// 		ProfileDisplayPicture: "https://avatars0.githubusercontent.com/u/39563?s=400&v=4",
		// 	},
		// }
		// for _, contact := range contacts {
		// 	if err := write(conn, contact); err != nil {
		// 		fmt.Println("ERRR", err)
		// 	}
		// }

		// applications := []CertificateResponse{
		// 	CertificateResponse{
		// 		Type:                  "nimona.io/application.Application",
		// 		Alias:                 "mochi",
		// 		Developer:             "mochi",
		// 		ProfileName:           "mochi",
		// 		IdentityPublicKey:     "ELCCyo6YF54LBdzuwUg8qC8vpfJX5rVzHU72buxQqL1C",
		// 		ProfileDisplayPicture: "https://avatarfiles.alphacoders.com/594/59481.jpg",
		// 	},
		// 	CertificateResponse{
		// 		Type:                  "nimona.io/application.Application",
		// 		Alias:                 "identity",
		// 		Developer:             "identity",
		// 		ProfileName:           "nimona identity",
		// 		IdentityPublicKey:     "7qGnd26xjLa1Y3P76AC7sZq4SE3rppqbuLhxaWFW7knB",
		// 		ProfileDisplayPicture: "https://avatarfiles.alphacoders.com/445/44578.jpg",
		// 	},
		// }
		// for _, application := range applications {
		// 	if err := write(conn, application); err != nil {
		// 		fmt.Println("ERRR", err)
		// 	}
		// }

		// o := object.FromMap(m)
		// incoming <- o
	}
}
