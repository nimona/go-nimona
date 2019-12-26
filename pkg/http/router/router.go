package router

import (
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	Routes       []*Route
	DefaultRoute Handler

	middleware []Handler
}

func New() *Router {
	r := &Router{
		DefaultRoute: func(ctx *Context) {
			ctx.Text(http.StatusNotFound, "Not found")
		},
	}

	return r
}

func (r *Router) Use(handler Handler) {
	r.middleware = append(r.middleware, handler)
}

func (r *Router) Handle(method, pattern string, handler Handler) {
	pattern = strings.Trim(pattern, "/")
	r.Routes = append(r.Routes, &Route{
		Method:  method,
		Pattern: regexp.MustCompile(pattern),
		Handler: handler,
	})
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := &Context{
		Request: req,
		Writer:  rw,
		Params:  map[string]string{},
	}

	handler := r.DefaultRoute

	for _, route := range r.Routes {
		if ctx.Request.Method != route.Method {
			continue
		}
		pattern := route.Pattern
		path := strings.Trim(ctx.Request.URL.Path, "/")
		ok, params := match(pattern, path)
		if !ok {
			continue
		}

		handler = route.Handler
		ctx.Params = params
		break
	}

	for _, middlewareHandler := range r.middleware {
		if ctx.aborted {
			return
		}
		middlewareHandler(ctx)
	}

	handler(ctx)
}

func match(pattern *regexp.Regexp, path string) (bool, map[string]string) {
	match := pattern.FindStringSubmatch(path)
	if len(match) == 0 {
		return false, map[string]string{}
	}
	params := map[string]string{}
	for i, name := range pattern.SubexpNames() {
		if i != 0 && name != "" && match[i] != "" {
			params[name] = match[i]
		}
	}
	return true, params
}
