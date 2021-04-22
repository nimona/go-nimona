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
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"

	"nimona.io/schema/relationship"
)

//go:embed assets/*
var assets embed.FS

const (
	pKeyIdentity = "IDENTITY_PRIVATE_KEY"
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
)

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

	if v, err := d.Preferences().Get(pKeyIdentity); err == nil {
		privateIdentityKey := &crypto.PrivateKey{}
		fmt.Println("GOT ID", string(v))
		if err := privateIdentityKey.UnmarshalString(v); err == nil {
			d.LocalPeer().PutPrimaryIdentityKey(*privateIdentityKey)
		}
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
			conv.SetPurgeChecker(pscanner.Map())
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
		k := d.LocalPeer().GetPrimaryIdentityKey()
		if k.IsEmpty() {
			return
		}
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
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
		if err := d.Preferences().Put(pKeyIdentity, k.String()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO persist key in config
		d.LocalPeer().PutPrimaryIdentityKey(k)
		http.Redirect(w, r, "/identity", http.StatusFound)
	})

	r.Get("/contacts", func(w http.ResponseWriter, r *http.Request) {
		k := d.LocalPeer().GetPrimaryIdentityKey()
		contacts := map[string]string{} // publickey/alias
		values := struct {
			IdentityLinked bool
			Contacts       []Contact
		}{
			IdentityLinked: false,
			Contacts:       []Contact{},
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
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
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
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
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
				Owner:  k.PublicKey(),
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
		contactsStreamRoot := relationship.RelationshipStreamRoot{
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
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
				Owner:  k.PublicKey(),
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
