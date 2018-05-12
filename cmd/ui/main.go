package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/asticode/go-astilectron"
	"github.com/nimona/go-nimona/api"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/wire"
)

const (
	width  = 300
	height = 450
)

var bootstrapPeerInfos = []mesh.PeerInfo{
	mesh.PeerInfo{
		ID: "andromeda.nimona.io",
		Addresses: []string{
			"tcp:andromeda.nimona.io:26800",
		},
	},
}

func main() {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		log.Fatal("Missing PEER_ID")
	}

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	reg := mesh.NewRegisty(peerID)
	msh := mesh.New(reg)

	for _, peerInfo := range bootstrapPeerInfos {
		reg.PutPeerInfo(&peerInfo)
	}

	msh.Listen(fmt.Sprintf(":%d", port))

	wre, _ := wire.NewWire(msh, reg)
	dht, _ := dht.NewDHT(wre, reg)

	msh.RegisterHandler("wire", wre)

	app, err := astilectron.New(astilectron.Options{
		AppName:            "nimona",
		AppIconDefaultPath: "resources/icon.png",
		AppIconDarwinPath:  "resources/icon.icns",
		BaseDirectoryPath:  os.TempDir(),
	})

	wre.HandleExtensionEvents("msg", func(event *wire.Message) error {
		// fmt.Printf("___ Got message from %s: %s\n", event.From, string(event.Payload))
		// Create the notification
		msg := ""
		json.Unmarshal(event.Payload, &msg)
		var n = app.NewNotification(&astilectron.NotificationOptions{
			Body:     msg,
			HasReply: astilectron.PtrBool(true),
			// Icon: "/path/to/icon",
			ReplyPlaceholder: "reply here",
			Title:            "Message from " + event.From,
		})
		n.Show()
		return nil
	})

	api := api.New(reg, dht)
	api.Router.Static("/webui", "./resources/app")

	listener, _ := net.Listen("tcp", ":0")
	go http.Serve(listener, api.Router)
	httpPort := listener.Addr().(*net.TCPAddr).Port

	fmt.Println("HTTP listening on", httpPort)

	if err != nil {
		log.Printf("[FATAL] Failed application setup: %v.", err)
		return
	}
	defer app.Close()
	app.HandleSignals()
	if err = app.Start(); err != nil {
		log.Printf("[FATAL] Failed application start: %v.", err)
	}

	url := fmt.Sprintf("http://localhost:%d/webui", httpPort)
	w, _ := app.NewWindow(url, &astilectron.WindowOptions{
		AcceptFirstMouse: astilectron.PtrBool(false),
		AlwaysOnTop:      astilectron.PtrBool(true),
		Center:           astilectron.PtrBool(false),
		Focusable:        astilectron.PtrBool(true),
		Frame:            astilectron.PtrBool(false),
		Height:           astilectron.PtrInt(height),
		Maximizable:      astilectron.PtrBool(false),
		Minimizable:      astilectron.PtrBool(false),
		Movable:          astilectron.PtrBool(false),
		Resizable:        astilectron.PtrBool(true), // false
		Show:             astilectron.PtrBool(false),
		Transparent:      astilectron.PtrBool(true),
		Width:            astilectron.PtrInt(width),
	})

	w.Create()

	t := app.NewTray(&astilectron.TrayOptions{
		Image:   astilectron.PtrStr("resources/tray.png"),
		Tooltip: astilectron.PtrStr("Tray's tooltip"),
	})

	t.Create()

	wVisible := false

	t.On(astilectron.EventNameTrayEventClicked, func(e astilectron.Event) bool {
		trayX := *e.Bounds.PositionOptions.X
		trayY := *e.Bounds.PositionOptions.Y
		var focusedDisplay *astilectron.Display
		for _, display := range app.Displays() {
			displayBounds := display.Bounds()
			if trayX >= displayBounds.X &&
				trayX <= displayBounds.X+displayBounds.Width &&
				trayY >= displayBounds.Y &&
				trayY <= displayBounds.Y+displayBounds.Height {
				focusedDisplay = display
				break
			}
		}
		if wVisible {
			w.Hide()
			wVisible = false
		} else {
			w.MoveInDisplay(focusedDisplay, trayX-width/2+10, 0)
			w.Show()
			wVisible = true
		}
		return false
	})

	w.On(astilectron.EventNameWindowEventBlur, func(e astilectron.Event) bool {
		w.Hide()
		wVisible = false
		return false
	})

	w.On(astilectron.EventNameWindowEventFocus, func(e astilectron.Event) bool {
		wVisible = true
		return false
	})

	app.Wait()
}
