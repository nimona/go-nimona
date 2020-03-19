package router

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		match   bool
		params  map[string]string
	}{
		{
			name:    "simple path, match",
			pattern: "/foo",
			path:    "/foo",
			match:   true,
			params:  map[string]string{},
		},
		{
			name:    "simple path, no match",
			pattern: "/bar",
			path:    "/foo",
			match:   false,
			params:  map[string]string{},
		},
		{
			name:    "path with non optional param, match",
			pattern: "/foo/(?P<foo>.+)",
			path:    "/foo/bar",
			match:   true,
			params: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:    "path with non optional param, no match, param",
			pattern: "/foo/(?P<foo>.+)",
			path:    "/not-foo/bar",
			match:   false,
			params:  map[string]string{},
		},
		{
			name:    "path with non optional param, no match, no param",
			pattern: "/foo/(?P<foo>.+)",
			path:    "/not-foo",
			match:   false,
			params:  map[string]string{},
		},
		{
			name:    "path with non optional params, match",
			pattern: "/foo/(?P<foo1>.+)/(?P<foo2>.+)",
			path:    "/foo/BAR11/BAR22",
			match:   true,
			params: map[string]string{
				"foo1": "BAR11",
				"foo2": "BAR22",
			},
		},
		{
			name:    "path with params, missing, match",
			pattern: "/foo/(?P<foo1>.+)/(?P<foo2>.+)",
			path:    "/foo//BAR22",
			match:   false,
			params:  map[string]string{},
		},
	}
	for _, tt := range tests {
		re := regexp.MustCompile(tt.pattern)
		match, params := match(re, tt.path)
		if match != tt.match {
			t.Errorf("match() match = %v, want %v", match, tt.match)
		}
		if !reflect.DeepEqual(params, tt.params) {
			t.Errorf("match() params = %v, want %v", params, tt.params)
		}
	}
}
