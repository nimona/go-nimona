package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/Masterminds/sprig"
	"github.com/geoah/go-hotwire"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"
)

//go:embed assets/*
var assets embed.FS

var (
	tplPeer = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
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
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.identity.html",
			),
	)
	tplObjects = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.objects.html",
			),
	)
	tplInnerPeerContentTypes = template.Must(
		template.New("inner.peer-content-types.html").
			Funcs(sprig.FuncMap()).
			ParseFS(
				assets,
				"assets/inner.peer-content-types.html",
			),
	)
)

func main() {
	r := chi.NewRouter()
	d, err := daemon.New(context.New())
	if err != nil {
		log.Fatal(err)
	}

	cssAssets, _ := fs.Sub(assets, "assets/css")
	r.Use(middleware.Logger)
	r.Mount("/css", twhandler.New(http.FS(cssAssets), "/css", twembed.New()))

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
			}{
				PublicKey:    connInfo.PublicKey.String(),
				Addresses:    connInfo.Addresses,
				ContentTypes: d.LocalPeer().GetContentTypes(),
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
