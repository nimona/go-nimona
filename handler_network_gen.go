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

func (t *NetworkInfoRequest) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{161}); err != nil {
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

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/network/info.request"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/network/info.request")); err != nil {
		return err
	}
	return nil
}

func (t *NetworkInfoRequest) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = NetworkInfoRequest{}
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
		return fmt.Errorf("NetworkInfoRequest: map struct too large (%d)", extra)
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

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *NetworkJoinRequest) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{163}); err != nil {
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

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/network/join.request"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/network/join.request")); err != nil {
		return err
	}

	// t.Metadata (nimona.Metadata) (struct)
	if len("Metadata") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Metadata\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Metadata"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Metadata")); err != nil {
		return err
	}

	if err := t.Metadata.MarshalCBOR(cw); err != nil {
		return err
	}

	// t.RequestedHandle (string) (string)
	if len("RequestedHandle") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"RequestedHandle\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("RequestedHandle"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("RequestedHandle")); err != nil {
		return err
	}

	if len(t.RequestedHandle) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.RequestedHandle was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(t.RequestedHandle))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.RequestedHandle)); err != nil {
		return err
	}
	return nil
}

func (t *NetworkJoinRequest) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = NetworkJoinRequest{}
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
		return fmt.Errorf("NetworkJoinRequest: map struct too large (%d)", extra)
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
		case "Metadata":

			{

				if err := t.Metadata.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Metadata: %w", err)
				}

			}
			// t.RequestedHandle (string) (string)
		case "RequestedHandle":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.RequestedHandle = string(sval)
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *NetworkJoinResponse) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{165}); err != nil {
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

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/network/join.response"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/network/join.response")); err != nil {
		return err
	}

	// t.Handle (string) (string)
	if len("Handle") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Handle\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Handle"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Handle")); err != nil {
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

	// t.Accepted (bool) (bool)
	if len("Accepted") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Accepted\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Accepted"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Accepted")); err != nil {
		return err
	}

	if err := cbg.WriteBool(w, t.Accepted); err != nil {
		return err
	}

	// t.Error (bool) (bool)
	if len("Error") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Error\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Error"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Error")); err != nil {
		return err
	}

	if err := cbg.WriteBool(w, t.Error); err != nil {
		return err
	}

	// t.ErrorDescription (string) (string)
	if len("ErrorDescription") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"ErrorDescription\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("ErrorDescription"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("ErrorDescription")); err != nil {
		return err
	}

	if len(t.ErrorDescription) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.ErrorDescription was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(t.ErrorDescription))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.ErrorDescription)); err != nil {
		return err
	}
	return nil
}

func (t *NetworkJoinResponse) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = NetworkJoinResponse{}
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
		return fmt.Errorf("NetworkJoinResponse: map struct too large (%d)", extra)
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

		// t.Handle (string) (string)
		case "Handle":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Handle = string(sval)
			}
			// t.Accepted (bool) (bool)
		case "Accepted":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}
			if maj != cbg.MajOther {
				return fmt.Errorf("booleans must be major type 7")
			}
			switch extra {
			case 20:
				t.Accepted = false
			case 21:
				t.Accepted = true
			default:
				return fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
			}
			// t.Error (bool) (bool)
		case "Error":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}
			if maj != cbg.MajOther {
				return fmt.Errorf("booleans must be major type 7")
			}
			switch extra {
			case 20:
				t.Error = false
			case 21:
				t.Error = true
			default:
				return fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
			}
			// t.ErrorDescription (string) (string)
		case "ErrorDescription":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.ErrorDescription = string(sval)
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
