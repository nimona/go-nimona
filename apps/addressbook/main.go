package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"

	"nimona.io/pkg/context"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"
)

//go:embed assets/*
var assets embed.FS

var (
	tplIndex = template.Must(
		template.ParseFS(
			assets,
			"assets/base.html",
			"assets/frame.peer.html",
		),
	)
	tplObjects = template.Must(
		template.ParseFS(
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
		identityKey := d.LocalPeer().GetPrimaryIdentityKey()
		identityPublicKey := ""
		if !identityKey.IsEmpty() {
			identityPublicKey = identityKey.PublicKey().String()
		}
		connInfo := d.LocalPeer().ConnectionInfo()
		err := tplIndex.Execute(
			w,
			struct {
				PublicKey         string
				IdentityPublicKey string
				Addresses         []string
				ContentTypes      []string
			}{
				PublicKey:         connInfo.PublicKey.String(),
				IdentityPublicKey: identityPublicKey,
				Addresses:         connInfo.Addresses,
				ContentTypes:      d.LocalPeer().GetContentTypes(),
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
