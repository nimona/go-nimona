// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package main

import (
	object "nimona.io/pkg/object"
	tilde "nimona.io/pkg/tilde"
)

const FileType = "nimona.io/File"

type File struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/File"`
	Name     string          `nimona:"name:s"`
	Blob     tilde.Digest    `nimona:"blob:r"`
}
