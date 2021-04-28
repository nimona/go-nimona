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

	"nimona.io/internal/rand"
	"nimona.io/pkg/certificateutils"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"

	"nimona.io/schema/relationship"
)

//go:embed assets/*
var assets embed.FS

const (
	pkKeyIdentity     = "IDENTITY_PRIVATE_KEY"
	pkPeerCertificate = "PEER_CERTIFICATE_JSON"
)

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
	tplCertificates = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/inner.certificate.html",
				"assets/frame.certificates.html",
			),
	)
	tplCertificatesCsr = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.certificates-csr.html",
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
		daemon daemon.Daemon
		// identity
		sync.RWMutex
		peerCertificateResponse *object.CertificateResponse
		identityPrivateKey      *crypto.PrivateKey
	}
)

// TODO(geoah): Not sure if we should be setting the identity public key from
// both the certificate and the private key.

func New(
	dae daemon.Daemon,
) (*Hub, error) {
	h := &Hub{
		daemon: dae,
	}

	if v, err := h.daemon.Preferences().Get(pkPeerCertificate); err == nil {
		crtResObj := &object.Object{}
		if err := json.Unmarshal([]byte(v), crtResObj); err != nil {
			return nil, err
		}
		crtRes := &object.CertificateResponse{}
		if err := crtRes.FromObject(crtResObj); err != nil {
			return nil, err
		}
		h.peerCertificateResponse = crtRes
	}

	if v, err := h.daemon.Preferences().Get(pkKeyIdentity); err == nil {
		privateIdentityKey := &crypto.PrivateKey{}
		if err := privateIdentityKey.UnmarshalString(v); err != nil {
			return nil, err
		}
		h.identityPrivateKey = privateIdentityKey
	}

	return h, nil
}

func (h *Hub) PutPeerCertificate(r *object.CertificateResponse) {
	h.Lock()
	defer h.Unlock()
	h.daemon.ObjectStore().Pin(r.ToObject().CID())
	h.daemon.ObjectStore().Put(r.ToObject())
	h.peerCertificateResponse = r
	b, _ := json.Marshal(r.ToObject())
	h.daemon.Preferences().Put(pkPeerCertificate, string(b))
	h.daemon.LocalPeer().PutPeerCertificate(r)
}

func (h *Hub) PutIdentityPrivateKey(k crypto.PrivateKey) {
	h.Lock()
	defer h.Unlock()
	h.identityPrivateKey = &k
	h.daemon.Preferences().Put(pkKeyIdentity, k.String())
}

func (h *Hub) GetIdentityPrivateKey() *crypto.PrivateKey {
	h.RLock()
	defer h.RUnlock()
	return h.identityPrivateKey
}

func (h *Hub) GetIdentityPublicKey() *crypto.PublicKey {
	h.RLock()
	defer h.RUnlock()
	if h.peerCertificateResponse != nil {
		// TODO(geoah) Owner or Signer? Can they be different? DOCUMENT
		return &h.peerCertificateResponse.Metadata.Owner
	}
	return nil
}

func (h *Hub) ForgetIdentity() {
	h.Lock()
	defer h.Unlock()
	h.daemon.Preferences().Remove(pkKeyIdentity)
	h.daemon.Preferences().Remove(pkPeerCertificate)
	h.daemon.LocalPeer().ForgetPeerCertificate()
	h.identityPrivateKey = nil
	h.peerCertificateResponse = nil
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
			// conv.SetPurgeChecker(pscanner.Map())
			return conv
		},
	)
	cssHandler.SetWriteCloserFunc(brotli.HTTPCompressor)
	r.Mount("/css", cssHandler)

	turboStream := hotwire.NewEventStream()
	r.Get("/events", turboStream.ServeHTTP)

	events, eventsClose := d.LocalPeer().ListenForUpdates()
	defer eventsClose()

	go func() {
		for {
			_, ok := <-events
			if !ok {
				return
			}
			if err := turboStream.SendEvent(
				hotwire.StreamActionReplace,
				"peer-content-types",
				tplInnerPeerContentTypes,
				struct {
					ContentTypes []string
				}{
					ContentTypes: d.LocalPeer().GetContentTypes(),
				},
			); err != nil {
				log.Println(err)
			}
		}
	}()

	go func() {
		k := h.GetIdentityPublicKey()
		if k == nil {
			return
		}
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: *k,
			},
		}
		contactEvents := d.ObjectManager().Subscribe(
			objectmanager.FilterByStreamCID(
				contactsStreamRoot.ToObject().CID(),
			),
		)
		for {
			o, err := contactEvents.Read()
			if err != nil {
				return
			}
			switch o.Type {
			case new(relationship.Added).Type():
				r := &relationship.Added{}
				if err := r.FromObject(o); err != nil {
					continue
				}
				if r.Alias == "" || r.RemoteParty.IsEmpty() {
					continue
				}
				turboStream.SendEvent(
					hotwire.StreamActionAppend,
					"contacts",
					tplInnerContact,
					Contact{
						Alias:     r.Alias,
						PublicKey: r.RemoteParty.String(),
					},
				)
			case new(relationship.Removed).Type():
				r := &relationship.Removed{}
				if err := r.FromObject(o); err != nil {
					continue
				}
				if r.RemoteParty.IsEmpty() {
					continue
				}
				turboStream.SendEvent(
					hotwire.StreamActionRemove,
					"contact-"+r.RemoteParty.String(),
					tplInnerContact,
					Contact{},
				)
			}
		}
	}()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		connInfo := d.LocalPeer().ConnectionInfo()
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
				ContentTypes: d.LocalPeer().GetContentTypes(),
				ConfigPath:   d.Config().Path,
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/identity/csr.png", func(w http.ResponseWriter, r *http.Request) {
		cid := r.URL.Query().Get("cid")
		q, _ := qrcode.New(
			"nimona://identity/csr?cid="+cid,
			qrcode.Medium,
		)
		q.DisableBorder = true
		png, _ := q.PNG(256)
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	})

	r.Get("/identity", func(w http.ResponseWriter, r *http.Request) {
		showMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("show"))
		linkMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("link"))
		var csr *object.CertificateRequest
		if linkMnemonic {
			csr = &object.CertificateRequest{
				Metadata: object.Metadata{
					Owner: d.LocalPeer().GetPrimaryPeerKey().PublicKey(),
				},
				Nonce:                  rand.String(12),
				VendorName:             "Nimona",
				VendorURL:              "https://nimona.io",
				ApplicationName:        "Hub",
				ApplicationDescription: "Nimona Hub",
				ApplicationURL:         "https://nimona.io/hub",
				Permissions: []object.CertificatePermission{{
					Metadata: object.Metadata{
						Owner: d.LocalPeer().GetPrimaryPeerKey().PublicKey(),
					},
					Types:   []string{"*"},
					Actions: []string{"*"},
				}},
			}
			k := d.LocalPeer().GetPrimaryPeerKey()
			csrSig, err := object.NewSignature(k, csr.ToObject())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			csr.Metadata.Signature = csrSig
			if _, err = d.ObjectManager().Put(
				context.New(
					context.WithParent(r.Context()),
					context.WithTimeout(3*time.Second),
				),
				csr.ToObject(),
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		values := struct {
			PublicKey    string
			PrivateBIP39 string
			Show         bool
			Link         bool
			CSR          *object.CertificateRequest
		}{
			Show: showMnemonic,
			Link: linkMnemonic,
			CSR:  csr,
		}
		go func() {
			csrResCh := certificateutils.WaitForCertificateResponse(
				context.New(
					context.WithTimeout(15*time.Minute),
				),
				d.Network(),
				csr,
			)
			csrRes := <-csrResCh
			if csrRes == nil {
				return
			}
			h.PutPeerCertificate(csrRes)
			values.PublicKey = h.GetIdentityPublicKey().String()
			turboStream.SendEvent(
				hotwire.StreamActionReplace,
				"peer-identity",
				tplIdentityInner,
				values,
			) // nolint: errcheck
			// TODO figure out how to surface errors?
		}()
		if k := h.GetIdentityPublicKey(); k != nil {
			values.PublicKey = k.String()
		}
		if k := h.GetIdentityPrivateKey(); k != nil {
			values.PrivateBIP39 = k.BIP39()
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
		k, err := crypto.NewEd25519PrivateKey(crypto.IdentityKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO(geoah): also create a certificate
		req := object.CertificateRequest{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
			},
			Nonce:                  rand.String(12),
			VendorName:             "Nimona",
			VendorURL:              "https://nimona.io",
			ApplicationName:        "Hub",
			ApplicationDescription: "Nimona Hub",
			ApplicationURL:         "https://nimona.io/hub",
			Permissions: []object.CertificatePermission{{
				Metadata: object.Metadata{
					Owner: d.LocalPeer().GetPrimaryPeerKey().PublicKey(),
				},
				Types:   []string{"*"},
				Actions: []string{"*"},
			}},
		}
		req.Metadata.Signature, err = object.NewSignature(k, req.ToObject())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		crtRes, err := object.NewCertificate(
			k,
			req,
			true,
			"Created by Nimona Hub",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.PutPeerCertificate(crtRes)
		h.PutIdentityPrivateKey(k)
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Post("/identity/link", func(w http.ResponseWriter, r *http.Request) {
		k, err := crypto.NewEd25519PrivateKeyFromBIP39(
			r.PostFormValue("mnemonic"),
			crypto.IdentityKey,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csrCID := object.CID(r.PostFormValue("csr"))
		if err = d.ObjectStore().Pin(csrCID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csrObj, err := d.ObjectStore().Get(csrCID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csr := &object.CertificateRequest{}
		err = csr.FromObject(csrObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csrRes, err := object.NewCertificate(k, *csr, true, "Signed by Hub")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = d.ObjectStore().Put(csrRes.ToObject())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = d.ObjectStore().Pin(csrRes.ToObject().CID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.PutPeerCertificate(csrRes)
		h.PutIdentityPrivateKey(k)
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/identity/forget", func(w http.ResponseWriter, r *http.Request) {
		h.ForgetIdentity()
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/contacts", func(w http.ResponseWriter, r *http.Request) {
		k := h.GetIdentityPublicKey()
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
		contactsStreamRootCID := contactsStreamRoot.ToObject().CID()
		objectReader, err := d.ObjectStore().GetByStream(contactsStreamRootCID)
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
				case new(relationship.Added).Type():
					r := &relationship.Added{}
					if err := r.FromObject(o); err != nil {
						continue
					}
					if r.Alias == "" || r.RemoteParty.IsEmpty() {
						continue
					}
					contacts[r.RemoteParty.String()] = r.Alias
				case new(relationship.Removed).Type():
					r := &relationship.Removed{}
					if err := r.FromObject(o); err != nil {
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
		k := h.GetIdentityPublicKey()
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
		contactsStreamRootCID := contactsStreamRoot.ToObject().CID()
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
				Owner:  *k,
				Stream: contactsStreamRootCID,
			},
			Alias:       alias,
			RemoteParty: remotePartyKey,
			Datetime:    time.Now().UTC().Format(time.RFC3339),
		}
		if _, err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			contactsStreamRoot.ToObject(),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			rel.ToObject(),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/contacts", http.StatusFound)
	})

	r.Get("/contacts/remove", func(w http.ResponseWriter, r *http.Request) {
		k := h.GetIdentityPublicKey()
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
		contactsStreamRootCID := contactsStreamRoot.ToObject().CID()
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
				Owner:  *k,
				Stream: contactsStreamRootCID,
			},
			RemoteParty: remotePartyKey,
			Datetime:    time.Now().UTC().Format(time.RFC3339),
		}
		if _, err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			contactsStreamRoot.ToObject(),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			rel.ToObject(),
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
			Types:   d.LocalPeer().GetContentTypes(),
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

	r.Get("/objects/{cid}", func(w http.ResponseWriter, r *http.Request) {
		cid := chi.URLParam(r, "cid")
		obj, err := d.ObjectStore().Get(object.CID(cid))
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err == objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		body, err := json.MarshalIndent(obj.ToMap(), "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		values := struct {
			CID           string
			Type          string
			JSON          template.HTML
			StreamRoot    string
			StreamObjects []*object.Object
		}{
			CID:  cid,
			Type: obj.Type,
			JSON: template.HTML(prettyJSON(string(body))),
		}
		if strings.HasPrefix(obj.Type, "stream:") {
			values.StreamRoot = obj.CID().String()
		} else if !obj.Metadata.Stream.IsEmpty() {
			values.StreamRoot = obj.Metadata.Stream.String()
		}
		if values.StreamRoot != "" {
			or, err := d.ObjectStore().GetByStream(
				object.CID(values.StreamRoot),
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

	r.Get("/certificates", func(w http.ResponseWriter, r *http.Request) {
		values := struct {
			URL                 string
			HasIdentity         bool
			CertificateReponses []*object.CertificateResponse
		}{
			URL:                 r.URL.String(),
			HasIdentity:         h.GetIdentityPublicKey() != nil,
			CertificateReponses: []*object.CertificateResponse{},
		}
		if h.GetIdentityPublicKey() == nil {
			err = tplCertificates.Execute(
				w,
				values,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		reader, err := d.ObjectStore().(*sqlobjectstore.Store).Filter(
			sqlobjectstore.FilterByObjectType(
				new(object.CertificateResponse).Type(),
			),
			sqlobjectstore.FilterByOwner(
				*h.GetIdentityPublicKey(),
			),
		)
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err != objectstore.ErrNotFound {
			for {
				o, err := reader.Read()
				if err != nil {
					break
				}
				crtRes := &object.CertificateResponse{}
				if err := crtRes.FromObject(o); err != nil {
					continue
				}
				values.CertificateReponses = append(
					values.CertificateReponses,
					crtRes,
				)
			}
		}
		err = tplCertificates.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/certificates/csr", func(w http.ResponseWriter, r *http.Request) {
		csrCID := object.CID(r.PostFormValue("csrCID"))
		csrProviders, err := d.Resolver().Lookup(
			context.New(
				context.WithParent(r.Context()),
				context.WithTimeout(3*time.Second),
			),
			resolver.LookupByCID(csrCID),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(csrProviders) == 0 {
			http.Error(w, "no results", http.StatusNotFound)
			return
		}
		csrObj, err := d.ObjectManager().Request(
			context.New(
				context.WithParent(r.Context()),
				context.WithTimeout(3*time.Second),
			),
			csrCID,
			csrProviders[0],
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = d.ObjectManager().Put(
			context.New(
				context.WithParent(r.Context()),
				context.WithTimeout(time.Second),
			),
			csrObj,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csr := &object.CertificateRequest{}
		err = csr.FromObject(csrObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		values := struct {
			URL                string
			CertificateRequest *object.CertificateRequest
		}{
			URL:                r.URL.String(),
			CertificateRequest: csr,
		}
		err = tplCertificatesCsr.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/certificates/csr-sign", func(w http.ResponseWriter, r *http.Request) {
		csrCID := object.CID(r.PostFormValue("csrCID"))
		csrObj, err := d.ObjectStore().Get(csrCID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = d.ObjectStore().Pin(csrCID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		csr := &object.CertificateRequest{}
		err = csr.FromObject(csrObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		privateIdentityKey := h.GetIdentityPrivateKey()
		if privateIdentityKey == nil {
			http.Error(w, "no private key", http.StatusInternalServerError)
			return
		}
		csrRes, err := object.NewCertificate(
			*privateIdentityKey,
			*csr,
			true,
			"",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = d.Network().Send(
			context.New(
				context.WithParent(r.Context()),
				context.WithTimeout(3*time.Second),
			),
			csrRes.ToObject(),
			csr.Metadata.Signature.Signer,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/certificates", http.StatusFound)
	})

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("unable to start http server, %s", err.Error())
	}
}

var prettyJSONReg = regexp.MustCompile(`(?mi)"(bah[a-z0-9]{59})"`)

func prettyJSON(b string) string {
	matches := prettyJSONReg.FindAllString(b, -1)
	for _, match := range matches {
		cid := strings.Trim(match, `"`)
		pretty := `"<a href="/objects/` + cid + `" target="_top">` + cid + `</a>"`
		b = strings.Replace(b, match, pretty, -1)
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
