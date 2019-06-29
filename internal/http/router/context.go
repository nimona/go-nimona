package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	Params map[string]string

	aborted bool
}

func (c *Context) Text(code int, body string) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)

	io.WriteString(c.Writer, fmt.Sprintf("%s\n", body)) // nolint: errcheck
}

func (c *Context) JSON(code int, body interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)

	json.NewEncoder(c.Writer).Encode(body) // nolint: errcheck
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) AbortWithError(code int, err error) {
	c.Text(code, err.Error())
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Abort() {
	c.aborted = true
}

func (c *Context) BindBody(v interface{}) error {
	// TODO check content type
	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close() // nolint: errcheck
	return decoder.Decode(&v)
}
