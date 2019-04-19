package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

var (
	indexTpl = template.Must(template.New("main").Parse(`
	<!DOCTYPE html>
	<head>
	<title>nimona.io</title>
	<meta name="go-import" content="{{ .Package }} {{ .VCS }} {{ .VCSRoot }}">
	</head>
	<body>{{ .Body }}</body>
	</html>
	`))
)

func handler(w http.ResponseWriter, req *http.Request) {
	d := struct {
		Package string
		VCS     string
		VCSRoot string
		Body    string
	}{
		Package: "nimona.io",
		VCS:     "git",
		VCSRoot: "https://github.com/nimona/go-nimona",
		Body:    "github.com/nimona/go-nimona",
	}
	var buf bytes.Buffer
	err := indexTpl.Execute(&buf, d)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(buf.Bytes()) // nolint
}
func main() {
	fmt.Println("Listening on :8000")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8000", nil)
}
