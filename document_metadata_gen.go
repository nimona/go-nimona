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

func (t *Metadata) MarshalCBORBytes() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := t.MarshalCBOR(w)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *Metadata) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)
	fieldCount := 3

	if zero.IsZeroVal(t.Owner) {
		fieldCount--
	}

	if zero.IsZeroVal(t.Timestamp) {
		fieldCount--
	}

	if zero.IsZeroVal(t.Signature) {
		fieldCount--
	}

	if _, err := cw.Write(cbg.CborEncodeMajorType(cbg.MajMap, uint64(fieldCount))); err != nil {
		return err
	}

	// t.Owner (string) (string)
	if !zero.IsZeroVal(t.Owner) {

		if len("owner") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"owner\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("owner"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("owner")); err != nil {
			return err
		}

		if len(t.Owner) > cbg.MaxLength {
			return xerrors.Errorf("Value in field t.Owner was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(t.Owner))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string(t.Owner)); err != nil {
			return err
		}
	}

	// t.Timestamp (typegen.CborTime) (struct)
	if !zero.IsZeroVal(t.Timestamp) {

		if len("timestamp") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"timestamp\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("timestamp"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("timestamp")); err != nil {
			return err
		}

		if err := t.Timestamp.MarshalCBOR(cw); err != nil {
			return err
		}
	}

	// t.Signature (nimona.Signature) (struct)
	if !zero.IsZeroVal(t.Signature) {

		if len("_signature") > cbg.MaxLength {
			return xerrors.Errorf("Value in field \"_signature\" was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("_signature"))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string("_signature")); err != nil {
			return err
		}

		if err := t.Signature.MarshalCBOR(cw); err != nil {
			return err
		}
	}
	return nil
}

func (t *Metadata) UnmarshalCBORBytes(b []byte) (err error) {
	*t = Metadata{}
	return t.UnmarshalCBOR(bytes.NewReader(b))
}

func (t *Metadata) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = Metadata{}
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
		return fmt.Errorf("Metadata: map struct too large (%d)", extra)
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
		// t.Owner (string) (string)
		case "owner":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Owner = string(sval)
			}
			// t.Timestamp (typegen.CborTime) (struct)
		case "timestamp":

			{

				if err := t.Timestamp.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Timestamp: %w", err)
				}

			}
			// t.Signature (nimona.Signature) (struct)
		case "_signature":

			{

				if err := t.Signature.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Signature: %w", err)
				}

			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *Signature) MarshalCBORBytes() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := t.MarshalCBOR(w)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *Signature) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{162}); err != nil {
		return err
	}

	// t.Signer (nimona.PeerID) (struct)
	if len("signer") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"signer\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("signer"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("signer")); err != nil {
		return err
	}

	if err := t.Signer.MarshalCBOR(cw); err != nil {
		return err
	}

	// t.X ([]uint8) (slice)
	if len("x") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"x\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("x"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("x")); err != nil {
		return err
	}

	if len(t.X) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.X was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(t.X))); err != nil {
		return err
	}

	if _, err := cw.Write(t.X[:]); err != nil {
		return err
	}
	return nil
}

func (t *Signature) UnmarshalCBORBytes(b []byte) (err error) {
	*t = Signature{}
	return t.UnmarshalCBOR(bytes.NewReader(b))
}

func (t *Signature) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = Signature{}
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
		return fmt.Errorf("Signature: map struct too large (%d)", extra)
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
		// t.Signer (nimona.PeerID) (struct)
		case "signer":

			{

				if err := t.Signer.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Signer: %w", err)
				}

			}
			// t.X ([]uint8) (slice)
		case "x":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.X: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.X = make([]uint8, extra)
			}

			if _, err := io.ReadFull(cr, t.X[:]); err != nil {
				return err
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
