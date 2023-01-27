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

func (t *PeerCapabilitiesRequest) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{161}); err != nil {
		return err
	}

	// t.Type (string) (string)
	if len("$type") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"$type\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$type"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("$type")); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/peer/capabilities.request"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/peer/capabilities.request")); err != nil {
		return err
	}
	return nil
}

func (t *PeerCapabilitiesRequest) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = PeerCapabilitiesRequest{}
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
		return fmt.Errorf("PeerCapabilitiesRequest: map struct too large (%d)", extra)
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
		// t.Type (string) (string)
		case "$type":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Type = string(sval)
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}

func (t *PeerCapabilitiesResponse) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{162}); err != nil {
		return err
	}

	// t.Type (string) (string)
	if len("$type") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"$type\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("$type"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("$type")); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/peer/capabilities.response"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/peer/capabilities.response")); err != nil {
		return err
	}

	// t.Capabilities ([]string) (slice)
	if len("Capabilities") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Capabilities\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Capabilities"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Capabilities")); err != nil {
		return err
	}

	if len(t.Capabilities) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Capabilities was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajArray, uint64(len(t.Capabilities))); err != nil {
		return err
	}
	for _, v := range t.Capabilities {
		if len(v) > cbg.MaxLength {
			return xerrors.Errorf("Value in field v was too long")
		}

		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(v))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, string(v)); err != nil {
			return err
		}
	}
	return nil
}

func (t *PeerCapabilitiesResponse) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = PeerCapabilitiesResponse{}
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
		return fmt.Errorf("PeerCapabilitiesResponse: map struct too large (%d)", extra)
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
		// t.Type (string) (string)
		case "$type":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Type = string(sval)
			}
			// t.Capabilities ([]string) (slice)
		case "Capabilities":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return err
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.Capabilities: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.Capabilities = make([]string, extra)
			}

			for i := 0; i < int(extra); i++ {

				{
					sval, err := cbg.ReadString(cr)
					if err != nil {
						return err
					}

					t.Capabilities[i] = string(sval)
				}
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
