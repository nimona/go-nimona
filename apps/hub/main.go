package main

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"

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
				"assets/frame.contacts.html",
			),
	)
	tplObjects = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			Funcs(tplFuncMap).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.objects.html",
			),
	)
)

func main() {
	r := chi.NewRouter()
	d, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath("~/.nimona-hub"),
		),
	)
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
			vfsutil.Walk(http.FS(assets), "assets", pscanner.WalkFunc(func(fn string) bool {
				switch filepath.Ext(fn) {
				case ".html":
					return true
				}
				return false
			}))
			conv := tailwind.New(w, dist)
			conv.SetPurgeChecker(pscanner.Map())
			return conv
		},
	)
	cssHandler.SetWriteCloserFunc(brotli.HTTPCompressor)
	r.Mount("/css", cssHandler)

	es := hotwire.NewEventStream()
	r.Get("/events", es.ServeHTTP)

	events, eventsClose := d.LocalPeer().ListenForUpdates()
	defer eventsClose()

	go func() {
		for {
			_, ok := <-events
			if !ok {
				return
			}
			if err := es.SendEvent(
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

	r.Get("/identity", func(w http.ResponseWriter, r *http.Request) {
		showMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("show"))
		linkMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("link"))
		values := struct {
			PublicKey    string
			PrivateBIP39 string
			Show         bool
			Link         bool
		}{
			Show: showMnemonic,
			Link: linkMnemonic,
		}
		if k := d.LocalPeer().GetPrimaryIdentityKey(); !k.IsEmpty() {
			values.PublicKey = k.PublicKey().String()
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
		// TODO persist key in config
		d.LocalPeer().PutPrimaryIdentityKey(k)
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
		// TODO persist key in config
		d.LocalPeer().PutPrimaryIdentityKey(k)
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/contacts", func(w http.ResponseWriter, r *http.Request) {
		k := d.LocalPeer().GetPrimaryIdentityKey()
		values := struct {
			IdentityLinked bool
			Contacts       map[string]string
		}{
			IdentityLinked: false,
			Contacts:       map[string]string{},
		}
		if k.IsEmpty() {
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
		contactsStreemRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
			},
		}
		contactsStreemRootCID := contactsStreemRoot.ToObject().CID()
		objectReader, err := d.ObjectStore().GetByStream(contactsStreemRootCID)
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
					values.Contacts[r.Alias] = r.RemoteParty.String()
				case new(relationship.Removed).Type():
					// TODO implement removal
				}
			}
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
		k := d.LocalPeer().GetPrimaryIdentityKey()
		values := struct {
			IdentityLinked bool
			Contacts       map[string]string
		}{
			IdentityLinked: false,
			Contacts:       map[string]string{},
		}
		if k.IsEmpty() {
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
		contactsStreemRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
			},
		}
		contactsStreemRootCID := contactsStreemRoot.ToObject().CID()
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
				Owner:  k.PublicKey(),
				Stream: contactsStreemRootCID,
			},
			Alias:       alias,
			RemoteParty: remotePartyKey,
			Datetime:    time.Now().UTC().Format(time.RFC3339),
		}
		if _, err := d.ObjectManager().Put(
			context.FromContext(r.Context()),
			contactsStreemRoot.ToObject(),
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
		reader, err := d.ObjectStore().(*sqlobjectstore.Store).Filter()
		if err != nil && err != objectstore.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		objects := []*object.Object{}
		if err != objectstore.ErrNotFound {
			objects, err = object.ReadAll(reader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		err = tplObjects.Execute(
			w,
			objects,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":3000", r)
}
