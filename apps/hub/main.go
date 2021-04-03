package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/Masterminds/sprig"
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
			),
	)
	tplPeerIdentity = template.Must(
		template.New("base.html").
			Funcs(sprig.FuncMap()).
			ParseFS(
				assets,
				"assets/base.html",
				"assets/frame.peer-identity.html",
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

	r.Get("/peer/identity", func(w http.ResponseWriter, r *http.Request) {
		showMnemonic, _ := strconv.ParseBool(r.URL.Query().Get("show"))
		values := struct {
			PublicKey    string
			PrivateBIP39 string
			Show         bool
		}{
			Show: showMnemonic,
		}
		if k := d.LocalPeer().GetPrimaryIdentityKey(); !k.IsEmpty() {
			values.PublicKey = k.PublicKey().String()
			values.PrivateBIP39 = k.BIP39()
		}
		err := tplPeerIdentity.Execute(
			w,
			values,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/peer/identity-new", func(w http.ResponseWriter, r *http.Request) {
		k, err := crypto.NewEd25519PrivateKey(crypto.IdentityKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO persist key in config
		d.LocalPeer().PutPrimaryIdentityKey(k)
		http.Redirect(w, r, "/peer/identity", http.StatusFound)
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
