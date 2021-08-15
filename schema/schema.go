package schema

import (
	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"
)

type (
	Context struct {
		Metadata    object.Metadata `nimona:"@metadata:m,type=Context,context=/schema"`
		Name        string          `nimona:"name:s"`
		Description string          `nimona:"description:s"`
		Version     string          `nimona:"version:s"`
		Types       []*Type         `nimona:"types:am"`
	}
	Type struct {
		Metadata   object.Metadata `nimona:"@metadata:m,type=Type,context=/schema"`
		Name       string          `nimona:"name:s"`
		Properties []*Property     `nimona:"properties:am"`
		Strict     bool            `nimona:"strict:b"`
	}
	Property struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=Property,context=/schema"`
		Name     string          `nimona:"name:s"`
		Hint     tilde.Hint      `nimona:"hint:s"`
		Type     string          `nimona:"type:s"`
		Context  tilde.Digest    `nimona:"context:r"`
		Required bool            `nimona:"required:b"`
		Repeated bool            `nimona:"repeated:b"`
	}
)

// TODO add validation method
