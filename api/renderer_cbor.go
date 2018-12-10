package api

import (
	"net/http"

	"github.com/ugorji/go/codec"
	"nimona.io/go/encoding"
)

// Cbor contains the given interface object.
type Cbor struct {
	Data interface{}
}

var cborContentType = []string{"application/cbor; charset=utf-8"}

// WriteContentType (Cbor) writes Cbor ContentType.
func (r Cbor) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, cborContentType)
}

// Render (Cbor) encodes the given interface object and writes data with custom ContentType.
func (r Cbor) Render(w http.ResponseWriter) error {
	return WriteCbor(w, r.Data)
}

// WriteCbor writes Cbor ContentType and encodes the given interface object.
func WriteCbor(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, cborContentType)
	return codec.NewEncoder(w, encoding.CborHandler()).Encode(obj)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
