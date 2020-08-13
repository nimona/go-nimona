package main

import (
	"io/ioutil"
	"net/http"
)

func handleIndex(w http.ResponseWriter, req *http.Request) {
	dat, _ := ioutil.ReadFile("index.html")
	w.Write(dat) // nolint
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.ListenAndServe(":http", nil)
}
