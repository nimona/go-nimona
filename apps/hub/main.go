package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/andybalholm/brotli"
	"github.com/geoah/go-hotwire"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gotailwindcss/tailwind"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"
	"github.com/gotailwindcss/tailwind/twpurge"
	"github.com/shurcooL/httpfs/vfsutil"
	"github.com/skip2/go-qrcode"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/did"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
	"nimona.io/schema/relationship"
)

//go:embed assets/*
var assets embed.FS

var (
	tplFuncMap = map[string]interface{}{
		"lastN": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[len(s)-n:]
		},
		"setQueryParam":    setQueryParam,
		"addQueryParam":    addQueryParam,
		"removeQueryParam": removeQueryParam,
	}
	tplPeer = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.peer.html",
				"assets/inner.peer-content-types.html",
			),
	)
	tplIdentity = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.identity.html",
				"assets/frame.identity-inner.html",
			),
	)
	tplIdentityInner = template.Must(
		template.New("frame.identity-inner.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/frame.identity-inner.html",
			),
	)
	tplInnerPeerContentTypes = template.Must(
		template.New("inner.peer-content-types.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/inner.peer-content-types.html",
			),
	)
	tplContacts = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/inner.contact.html",
				"assets/frame.contacts.html",
			),
	)
	tplInnerContact = template.Must(
		template.New("inner.contact.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/inner.contact.html",
			),
	)
	tplObjects = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/inner.object.html",
				"assets/frame.objects.html",
			),
	)
	tplObject = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.object.html",
			),
	)
)

type (
	Contact struct {
		Alias     string
		PublicKey string
	}
	Hub struct {
		daemon              daemon.Daemon
		keyStreamManager    keystream.Manager
		keyStreamController keystream.Controller
		sync.RWMutex
	}
)

func New(
	dae daemon.Daemon,
) (*Hub, error) {
	ksm, err := keystream.NewKeyManager(
		dae.Network(),
		dae.ObjectStore().(*sqlobjectstore.Store),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct keystream manager: %w", err)
	}

	h := &Hub{
		daemon:           dae,
		keyStreamManager: ksm,
	}

	kscs := ksm.ListControllers()
	switch len(kscs) {
	case 0:
	case 1:
		h.keyStreamController = kscs[0]
	default:
		return nil, fmt.Errorf(
			"expected 1 keystream controller, got %d",
			len(kscs),
		)
	}

	return h, nil
}

// TODO(geoah): fix identity
// func (h *Hub) SetPeerCertificate(r *object.CertificateResponse) {
// 	h.Lock()
// 	defer h.Unlock()
// 	h.daemon.ObjectStore().Pin(object.MustMarshal(r).Hash())
// 	h.daemon.ObjectStore().Put(object.MustMarshal(r))
// 	h.peerCertificateResponse = r
// 	b, _ := json.Marshal(object.MustMarshal(r))
// 	h.daemon.Preferences().Put(pkPeerCertificate, string(b))
// 	h.daemon.LocalPeer().SetPeerCertificate(r)
// }

func (h *Hub) PutKeyStreamController(c keystream.Controller) {
	h.Lock()
	defer h.Unlock()
	h.keyStreamController = c
}

func (h *Hub) GetKeyStreamController() keystream.Controller {
	h.RLock()
	defer h.RUnlock()
	return h.keyStreamController
}

func (h *Hub) GetIdentityDID() *did.DID {
	h.RLock()
	defer h.RUnlock()
	if h.keyStreamController == nil {
		return nil
	}
	d := h.keyStreamController.GetKeyStream().GetDID()
	return &d
}

func (h *Hub) ForgetIdentity() {
	h.Lock()
	defer h.Unlock()
	// h.daemon.Preferences().Remove(pkKeyIdentity)
	// h.daemon.Preferences().Remove(pkPeerCertificate)
	// TODO(geoah): fix identity
	// h.daemon.LocalPeer().ForgetPeerCertificate()
	// h.identityPrivateKey = nil
	// h.peerCertificateResponse = nil
	// TODO truncate db
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	r := chi.NewRouter()
	d, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath("~/.nimona-hub"),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultListenOnExternalPort(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	h, err := New(d)
	if err != nil {
		log.Fatal(err)
	}

	cssAssets, _ := fs.Sub(assets, "assets/css")
	r.Use(middleware.Logger)

	cssHandler := twhandler.NewFromFunc(
		http.FS(cssAssets),
		"/css",
		func(w io.Writer) *tailwind.Converter {
			dist := twembed.New()
			pscanner, err := twpurge.NewScannerFromDist(dist)
			if err != nil {
				log.Fatal(err)
			}
			vfsutil.Walk(
				http.FS(assets),
				"assets",
				pscanner.WalkFunc(
					func(fn string) bool {
						switch filepath.Ext(fn) {
						case ".html":
							return true
						}
						return false
					},
				),
			)
			conv := tailwind.New(w, dist)
			return conv
		},
	)
	cssHandler.SetWriteCloserFunc(brotli.HTTPCompressor)
	r.Mount("/css", cssHandler)

	turboStream := hotwire.NewEventStream()
	r.Get("/events", turboStream.ServeHTTP)

	// events, eventsClose := d.LocalPeer().ListenForUpdates()
	// defer eventsClose()

	// go func() {
	// 	for {
	// 		_, ok := <-events
	// 		if !ok {
	// 			return
	// 		}
	// 		if err := turboStream.SendEvent(
	// 			"any",
	// 			hotwire.StreamActionReplace,
	// 			"peer-content-types",
	// 			tplInnerPeerContentTypes,
	// 			struct {
	// 				ContentTypes []string
	// 			}{
	// 				ContentTypes: []string{}, // TODO get from object store
	// 			},
	// 		); err != nil {
	// 			log.Println(err)
	// 		}
	// 	}
	// }()

	go func() {
		k := h.GetIdentityDID()
		if k == nil {
			return
		}
		contactsStreamRoot := &relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: *k,
			},
		}
		contactEvents := d.ObjectManager().Subscribe(
			objectmanager.FilterByStreamHash(
				object.MustMarshal(contactsStreamRoot).Hash(),
			),
		)
		for {
			o, err := contactEvents.Read()
			if err != nil {
				return
			}
			switch o.Type {
			case relationship.AddedType:
				r := &relationship.Added{}
				if err := object.Unmarshal(o, r); err != nil {
					continue
				}
				if r.Alias == "" || r.RemoteParty.IsEmpty() {
					continue
				}
				turboStream.SendEvent(
					"any",
					hotwire.StreamActionAppend,
					"contacts",
					tplInnerContact,
					Contact{
						Alias:     r.Alias,
						PublicKey: r.RemoteParty.String(),
					},
				)
			case relationship.RemovedType:
				r := &relationship.Removed{}
				if err := object.Unmarshal(o, r); err != nil {
					continue
				}
				if r.RemoteParty.IsEmpty() {
					continue
				}
				turboStream.SendEvent(
					"any",
					hotwire.StreamActionRemove,
					"contact-"+r.RemoteParty.String(),
					tplInnerContact,
					Contact{},
				)
			}
		}
	}()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		connInfo := d.Network().GetConnectionInfo()
		err := tplPeer.Execute(
			w,
			struct {
				PublicKey    string
				Addresses    []string
				ContentTypes []string
				ConfigPath   string
			}{
				PublicKey:    connInfo.PublicKey.String(),
				Addresses:    connInfo.Addresses,
				ContentTypes: []string{}, // TODO get types from sql store
				ConfigPath:   d.Config().Path,
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/identity/delegationRequest.png", func(w http.ResponseWriter, r *http.Request) {
		hash := r.URL.Query().Get("hash")
		q, _ := qrcode.New(
			"nimona://identity/delegationRequest?hash="+hash,
			qrcode.Medium,
		)
		q.DisableBorder = true
		png, _ := q.PNG(256)
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	})

	r.Get("/identity", func(w http.ResponseWriter, r *http.Request) {
		showMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("show"))
		requestDelegation, _ := strconv.ParseBool(r.URL.Query().Get("link"))
		delegateRequestHash := r.URL.Query().Get("delegateRequestHash")
		delegateRequestSign, _ := strconv.ParseBool(r.URL.Query().Get("delegateRequestSign"))
		peerKey := d.Network().GetPeerKey()

		values := struct {
			DID                      string
			DelegateDIDs             []did.DID
			Delegated                bool
			PrivateBIP39             string
			Show                     bool
			Link                     bool
			DelegationRequest        *keystream.DelegationRequest
			DelegationRequestHash    string
			DelegationRequestSuccess bool
			DelegationRequestError   string
		}{
			Show:                  showMnemonic,
			Link:                  requestDelegation,
			DelegationRequestHash: delegateRequestHash,
		}

		if requestDelegation {
			dr, cCh, err := h.keyStreamManager.NewDelegationRequest(
				context.Background(), // TODO: add timeout
				keystream.DelegationRequestVendor{
					VendorName:             "Nimona",
					VendorURL:              "https://nimona.io",
					ApplicationName:        "Hub",
					ApplicationDescription: "Nimona Hub",
					ApplicationURL:         "https://nimona.io/hub",
				},
				keystream.Permissions{
					Contexts: []string{"*"},
					Actions:  []string{"*"},
				},
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			csrObj := object.MustMarshal(dr)
			err = object.Sign(peerKey, csrObj)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			go func() {
				cC := <-cCh
				if cC == nil {
					return
				}
				values.Link = false
				h.PutKeyStreamController(cC)
				kss := h.GetKeyStreamController().GetKeyStream()
				values.DID = kss.GetDID().String()
				values.DelegateDIDs = kss.Delegates
				values.Delegated = !kss.Delegator.IsEmpty()
				fmt.Println(">>> pushing")
				turboStream.SendEvent(
					"any",
					hotwire.StreamActionReplace,
					"peer-identity",
					tplIdentityInner,
					values,
				) // nolint: errcheck
				// TODO figure out how to surface errors?
			}()

			values.DelegationRequest = dr
			values.DelegationRequestHash = csrObj.Hash().String()
		}

		if delegateRequestHash != "" {
			drh := tilde.Digest(delegateRequestHash)
			ps, err := h.daemon.Resolver().Lookup(
				context.New(context.WithTimeout(5*time.Second)),
				resolver.LookupByHash(drh),
			)

			if err != nil || len(ps) == 0 {
				values.DelegationRequestError = "Could not find request providers"
				if err != nil {
					values.DelegationRequestError += ", " + err.Error()
				}
			} else {
				dro, err := h.daemon.ObjectManager().Request(
					context.New(context.WithTimeout(time.Second)),
					drh,
					ps[0],
				)
				if err != nil {
					values.DelegationRequestError = "Could not fetch request"
				} else {
					dr := &keystream.DelegationRequest{}
					object.Unmarshal(dro, dr)
					values.DelegationRequest = dr
					if delegateRequestSign {
						err := h.keyStreamManager.HandleDelegationRequest(
							context.New(context.WithTimeout(5*time.Second)),
							dr,
							h.GetKeyStreamController(),
						)
						if err != nil {
							values.DelegationRequestError = "Could not handle request"
						} else {
							values.DelegationRequestSuccess = true
						}
					}
				}
			}
		}

		if h.GetKeyStreamController() != nil {
			kss := h.GetKeyStreamController().GetKeyStream()
			values.DID = kss.GetDID().String()
			values.DelegateDIDs = kss.Delegates
			values.Delegated = !kss.Delegator.IsEmpty()
			// values.PrivateBIP39 = kss.CurrentKey().BIP39()
		}

		err := tplIdentity.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/identity/new", func(w http.ResponseWriter, r *http.Request) {
		ikc, err := h.keyStreamManager.NewController(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.PutKeyStreamController(ikc)

		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/identity/forget", func(w http.ResponseWriter, r *http.Request) {
		h.ForgetIdentity()
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/contacts", func(w http.ResponseWriter, r *http.Request) {
		k := h.GetIdentityDID()
		contacts := map[string]string{} // publickey/alias
		values := struct {
			IdentityLinked bool
			Contacts       []Contact
		}{
			IdentityLinked: false,
			Contacts:       []Contact{},
		}
		if k == nil {
			err := tplContacts.Execute(
				w,
				values,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		values.IdentityLinked = true
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: *k,
			},
		}
		contactsStreamRootHash := object.MustMarshal(contactsStreamRoot).Hash()
		objectReader, err := d.ObjectStore().GetByStream(contactsStreamRootHash)
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if objectReader != nil {
			for {
				o, err := objectReader.Read()
				if err != nil {
					break
				}
				switch o.Type {
				case relationship.AddedType:
					r := &relationship.Added{}
					if err := object.Unmarshal(o, r); err != nil {
						continue
					}
					if r.Alias == "" || r.RemoteParty.IsEmpty() {
						continue
					}
					contacts[r.RemoteParty.String()] = r.Alias
				case relationship.RemovedType:
					r := &relationship.Removed{}
					if err := object.Unmarshal(o, r); err != nil {
						continue
					}
					if r.RemoteParty.IsEmpty() {
						continue
					}
					delete(contacts, r.RemoteParty.String())
				}
			}
		}
		for pk, alias := range contacts {
			values.Contacts = append(
				values.Contacts,
				Contact{
					Alias:     alias,
					PublicKey: pk,
				},
			)
		}
		if err := tplContacts.Execute(
			w,
			values,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/contacts/add", func(w http.ResponseWriter, r *http.Request) {
		k := h.GetIdentityDID()
		values := struct {
			IdentityLinked bool
			Contacts       map[string]string
		}{
			IdentityLinked: false,
			Contacts:       map[string]string{},
		}
		if k == nil {
			err := tplContacts.Execute(
				w,
				values,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		values.IdentityLinked = true
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: *k,
			},
		}
		contactsStreamRootHash := object.MustMarshal(contactsStreamRoot).Hash()
		alias := r.PostFormValue("alias")
		remoteParty := r.PostFormValue("remoteParty")
		if alias == "" || remoteParty == "" {
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		remotePartyKey := crypto.PublicKey{}
		if err := remotePartyKey.UnmarshalString(remoteParty); err != nil {
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		rel := relationship.Added{
			Metadata: object.Metadata{
				Owner: *k,
				Root:  contactsStreamRootHash,
			},
			Alias:       alias,
			RemoteParty: remotePartyKey,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		if err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			object.MustMarshal(contactsStreamRoot),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := d.ObjectManager().Append(
			context.FromContext(r.Context()),
			object.MustMarshal(rel),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/contacts", http.StatusFound)
	})

	r.Get("/contacts/remove", func(w http.ResponseWriter, r *http.Request) {
		k := h.GetIdentityDID()
		values := struct {
			IdentityLinked bool
			Contacts       map[string]string
		}{
			IdentityLinked: false,
			Contacts:       map[string]string{},
		}
		if k == nil {
			err := tplContacts.Execute(
				w,
				values,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		values.IdentityLinked = true
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: *k,
			},
		}
		contactsStreamRootHash := object.MustMarshal(contactsStreamRoot).Hash()
		remoteParty := r.URL.Query().Get("publicKey")
		if remoteParty == "" {
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		remotePartyKey := crypto.PublicKey{}
		if err := remotePartyKey.UnmarshalString(remoteParty); err != nil {
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		rel := relationship.Removed{
			Metadata: object.Metadata{
				Owner: *k,
				Root:  contactsStreamRootHash,
			},
			RemoteParty: remotePartyKey,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		if err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			object.MustMarshal(contactsStreamRoot),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := d.ObjectManager().Append(
			context.FromContext(r.Context()),
			object.MustMarshal(rel),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/contacts", http.StatusFound)
	})

	r.Get("/objects", func(w http.ResponseWriter, r *http.Request) {
		sqlFilters := []sqlobjectstore.FilterOption{}
		filters := []string{}
		vs, _ := url.ParseQuery(r.URL.RawQuery)
		if vf, ok := vs["type"]; ok {
			filters = vf
			for _, vvf := range vf {
				sqlFilters = append(
					sqlFilters,
					sqlobjectstore.FilterByObjectType(vvf),
				)
			}
		}
		reader, err := d.ObjectStore().(*sqlobjectstore.Store).Filter(sqlFilters...)
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		values := struct {
			URL     string
			Filters []string
			Types   []string
			Objects []*object.Object
		}{
			URL:     r.URL.String(),
			Filters: filters,
			Types:   []string{}, // TODO get from object store
			Objects: []*object.Object{},
		}
		if err != objectstore.ErrNotFound {
			values.Objects, err = object.ReadAll(reader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		err = tplObjects.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/objects/{hash}", func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		obj, err := d.ObjectStore().Get(tilde.Digest(hash))
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err == objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		body, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		values := struct {
			Hash          string
			Type          string
			JSON          template.HTML
			StreamRoot    string
			StreamObjects []*object.Object
		}{
			Hash: hash,
			Type: obj.Type,
			JSON: template.HTML(prettyJSON(string(body))),
		}
		if strings.HasPrefix(obj.Type, "stream:") {
			values.StreamRoot = obj.Hash().String()
		} else if !obj.Metadata.Root.IsEmpty() {
			values.StreamRoot = obj.Metadata.Root.String()
		}
		if values.StreamRoot != "" {
			or, err := d.ObjectStore().GetByStream(
				tilde.Digest(values.StreamRoot),
			)
			if err == nil {
				os, err := object.ReadAll(or)
				if err == nil {
					values.StreamObjects = os
				}
			}
		}
		err = tplObject.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("unable to start http server, %s", err.Error())
	}
}

var prettyJSONReg = regexp.MustCompile(`(?mi):r": "([a-zA-Z0-9]{44})"`)

func prettyJSON(b string) string {
	matches := prettyJSONReg.FindAllStringSubmatch(b, -1)
	for _, match := range matches {
		hash := strings.Trim(match[1], `"`)
		pretty := `<a href="/objects/` + hash + `" target="_top">` + hash + `</a>`
		b = strings.Replace(b, hash, pretty, -1)
	}
	return b
}

func setQueryParam(u, k, v string) string {
	vu, _ := url.Parse(u)
	vq := vu.Query()
	vq.Set(k, v)
	vu.RawQuery = vq.Encode()
	return vu.String()
}

func addQueryParam(u, k, v string) string {
	vu, _ := url.Parse(u)
	vq := vu.Query()
	vq.Add(k, v)
	vu.RawQuery = vq.Encode()
	return vu.String()
}

func removeQueryParam(u, k, v string) string {
	vu, _ := url.Parse(u)
	if vq, ok := vu.Query()[k]; ok {
		// left:=[]string{}
		vuq := vu.Query()
		vuq.Del(k)
		for _, vv := range vq {
			if vv == v {
				continue
			}
			// left=append(left, vv)
			vuq.Add(k, vv)
		}
		vu.RawQuery = vuq.Encode()
	}
	return vu.String()
}
