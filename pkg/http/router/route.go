package router

import (
	"regexp"
)

type Route struct {
	Method  string
	Pattern *regexp.Regexp
	Handler Handler
}
