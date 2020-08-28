package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"nimona.io/pkg/network"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"

	"nimona.io/internal/rand"
	"nimona.io/pkg/peer"
)

var indexTemplate = `
<div>nonce: {{ .Nonce }}</div>
<div>asr: {{ .ASR }}</div>
<div><img src="qr.png?s={{ .ASR }}" /></div>
<!--
<input id="input" type="text" />
<button onclick="send()">Send</button>
-->
<pre id="output"></pre>
<script>
    var input = document.getElementById("input");
    var output = document.getElementById("output");
    var socket = new WebSocket("ws://nimona.io:8080/echo?n={{ .Nonce }}");

    socket.onopen = function () {
        output.innerHTML += "status: waiting\n";
    };

    socket.onmessage = function (e) {
        output.innerHTML += "server: " + e.data + "\n";
    };

    function send() {
        socket.send(input.value);
        input.value = "";
    }
</script>
`

type Values struct {
	Nonce string
	ASR   string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// nolint: errcheck
func main() {
	net := serve()

	http.HandleFunc("/qr.png", func(w http.ResponseWriter, r *http.Request) {
		s := strings.Replace(r.URL.Query().Get("s"), "/qr.png?s=", "", 1)
		fmt.Println(">>>", s)
		png, _ := qrcode.Encode(s, qrcode.Medium, 256)
		w.Header().Set("content-type", "image/png")
		w.Write(png)
	})

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("*** sub start")
		conn, _ := upgrader.Upgrade(w, r, nil)
		sub := net.Subscribe(
			network.FilterByObjectType("nimona.io/peer.Certificate"),
		)
		n := strings.Replace(r.URL.Query().Get("n"), "/echo?n=", "", 1)
		fmt.Println("nonce:", n)

		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("status: subscribed to events"),
		)
		defer fmt.Println("*** sub stop")
		defer sub.Cancel()
		go func() {
			for {
				e, err := sub.Next()
				if err != nil {
					return
				}
				fmt.Println("*** GOT", e.Payload.GetType())
				r := &peer.Certificate{}
				if err := r.FromObject(e.Payload); err != nil {
					conn.WriteMessage(
						websocket.TextMessage,
						[]byte(
							"update: received invalid event from "+
								e.Sender.String(),
						),
					)
					continue
				}

				if r.Nonce != n {
					conn.WriteMessage(
						websocket.TextMessage,
						[]byte(
							"update: received invalid nonce from "+
								e.Sender.String(),
						),
					)
					continue
				}

				conn.WriteMessage(
					websocket.TextMessage,
					[]byte("update: signed by "+e.Sender.String()),
				)

				conn.WriteMessage(
					websocket.TextMessage,
					[]byte("status: done "),
				)
			}
		}()
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	tmpl := template.Must(template.New("index").Parse(indexTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		n := rand.String(8)
		asr := &peer.CertificateRequest{
			ApplicationName:        "Foobar App",
			ApplicationDescription: "An app that does nothing",
			ApplicationURL:         "https://github.com/nimona",
			Subject: net.LocalPeer().
				GetPrimaryPeerKey().
				PublicKey().
				String(),
			Resources: []string{
				"nimona.io/**",
				"mochi.io/**",
			},
			Actions: []string{
				"read",
				"write",
				"delete",
				"archive",
			},
			Nonce: n,
		}

		m := asr.ToObject().ToMap()
		delete(m, "_signature:m")

		b, _ := json.Marshal(m)
		s := string(b)
		v := Values{
			ASR:   "/qr.png?s=" + s,
			Nonce: n,
		}
		w.Header().Set("content-type", "text/html")
		tmpl.Execute(w, v) // nolint: errcheck
	})

	http.ListenAndServe(":8080", nil) // nolint: errcheck
}
