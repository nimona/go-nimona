package protocol

import (
	shortid "github.com/teris-io/shortid"
)

func generateReqID() string {
	rid, _ := shortid.Generate()
	return rid
}
