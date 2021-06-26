// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package filesharing

import (
	chore "nimona.io/pkg/chore"
	object "nimona.io/pkg/object"
)

type File struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/File"`
	Name     string          `nimona:"name:s"`
	Chunks   []chore.Hash    `nimona:"chunks:as"`
}

func (e *File) Type() string {
	return "nimona.io/File"
}

type TransferDone struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/TransferDone"`
	Nonce    string          `nimona:"nonce:s"`
}

func (e *TransferDone) Type() string {
	return "nimona.io/TransferDone"
}

type TransferRequest struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/TransferRequest"`
	File     File            `nimona:"file:m"`
	Nonce    string          `nimona:"nonce:s"`
}

func (e *TransferRequest) Type() string {
	return "nimona.io/TransferRequest"
}

type TransferResponse struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/TransferResponse"`
	Nonce    string          `nimona:"nonce:s"`
	Accepted bool            `nimona:"accepted:b"`
}

func (e *TransferResponse) Type() string {
	return "nimona.io/TransferResponse"
}
