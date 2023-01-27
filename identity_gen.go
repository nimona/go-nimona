// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package nimona

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	cid "github.com/ipfs/go-cid"
	zero "github.com/vikyd/zero"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = bytes.Compare
var _ = xerrors.Errorf
var _ = cid.Undef
var _ = math.E
var _ = sort.Sort
var _ = zero.IsZeroVal

func (t *Identity) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{162}); err != nil {
		return err
	}

	// t._ (string) (string)
	if len("$type") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"$type\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$type"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("$type")); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/identity/id"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/identity/id")); err != nil {
		return err
	}

	// t.KeyGraphID (nimona.DocumentID) (struct)
	if len("keyGraphID") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"keyGraphID\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("keyGraphID"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("keyGraphID")); err != nil {
		return err
	}

	if err := t.KeyGraphID.MarshalCBOR(cw); err != nil {
		return err
	}
	return nil
}

func (t *Identity) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = Identity{}
	}

	cr := cbg.NewCborReader(r)

	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("Identity: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadString(cr)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t._ (string) (string) - ignored

		// t.KeyGraphID (nimona.DocumentID) (struct)
		case "keyGraphID":

			{

				if err := t.KeyGraphID.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.KeyGraphID: %w", err)
				}

			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *IdentityAlias) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)
	fieldCount := 3

	if zero.IsZeroVal(t.Network) {
		fieldCount--
	}

	if zero.IsZeroVal(t.Handle) {
		fieldCount--
	}

	if _, err := cw.Write(cbg.CborEncodeMajorType(cbg.MajMap, uint64(fieldCount))); err != nil {
		return err
	}

	// t._ (string) (string)
	if len("$type") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"$type\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$type"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("$type")); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/identity.alias"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/identity.alias")); err != nil {
		return err
	}

	// t.Network (nimona.NetworkAlias) (struct)
	if !zero.IsZeroVal(t.Network) {

		if len("network") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"network\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("network"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("network")); err != nil {
			return err
		}

		if err := t.Network.MarshalCBOR(cw); err != nil {
			return err
		}
	}

	// t.Handle (string) (string)
	if !zero.IsZeroVal(t.Handle) {

		if len("handle") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"handle\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("handle"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("handle")); err != nil {
			return err
		}

		if len(t.Handle) > cbg.MaxLength {
			return xerrors.Errorf("Value in field t.Handle was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(t.Handle))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string(t.Handle)); err != nil {
			return err
		}
	}
	return nil
}

func (t *IdentityAlias) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = IdentityAlias{}
	}

	cr := cbg.NewCborReader(r)

	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("IdentityAlias: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadString(cr)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t._ (string) (string) - ignored

		// t.Network (nimona.NetworkAlias) (struct)
		case "network":

			{

				if err := t.Network.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Network: %w", err)
				}

			}
			// t.Handle (string) (string)
		case "handle":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Handle = string(sval)
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *KeyGraph) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)
	fieldCount := 4

	if zero.IsZeroVal(t.Metadata) {
		fieldCount--
	}

	if _, err := cw.Write(cbg.CborEncodeMajorType(cbg.MajMap, uint64(fieldCount))); err != nil {
		return err
	}

	// t._ (string) (string)
	if len("$type") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"$type\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$type"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("$type")); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/identity"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/identity")); err != nil {
		return err
	}

	// t.Metadata (nimona.Metadata) (struct)
	if !zero.IsZeroVal(t.Metadata) {

		if len("$metadata") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"$metadata\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$metadata"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("$metadata")); err != nil {
			return err
		}

		if err := t.Metadata.MarshalCBOR(cw); err != nil {
			return err
		}
	}

	// t.Keys (nimona.PublicKey) (slice)
	if len("keys") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"keys\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("keys"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("keys")); err != nil {
		return err
	}

	if len(t.Keys) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Keys was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(t.Keys))); err != nil {
		return err
	}

	if _, err := cw.Write(t.Keys[:]); err != nil {
		return err
	}

	// t.Next (nimona.PublicKey) (slice)
	if len("next") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"next\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("next"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("next")); err != nil {
		return err
	}

	if len(t.Next) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Next was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(t.Next))); err != nil {
		return err
	}

	if _, err := cw.Write(t.Next[:]); err != nil {
		return err
	}
	return nil
}

func (t *KeyGraph) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = KeyGraph{}
	}

	cr := cbg.NewCborReader(r)

	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("KeyGraph: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadString(cr)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t._ (string) (string) - ignored

		// t.Metadata (nimona.Metadata) (struct)
		case "$metadata":

			{

				if err := t.Metadata.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Metadata: %w", err)
				}

			}
			// t.Keys (nimona.PublicKey) (slice)
		case "keys":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Keys: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Keys = make([]uint8, extra)
			}

			if _, err := io.ReadFull(cr, t.Keys[:]); err != nil {
				return err
			}
			// t.Next (nimona.PublicKey) (slice)
		case "next":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Next: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Next = make([]uint8, extra)
			}

			if _, err := io.ReadFull(cr, t.Next[:]); err != nil {
				return err
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
