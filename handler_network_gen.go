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

func (t *NetworkInfoRequest) MarshalCBORBytes() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := t.MarshalCBOR(w)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

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

func (t *NetworkInfoRequest) UnmarshalCBORBytes(b []byte) (err error) {
	*t = NetworkInfoRequest{}
	return t.UnmarshalCBOR(bytes.NewReader(b))
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

func (t *NetworkInfo) MarshalCBORBytes() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := t.MarshalCBOR(w)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *NetworkInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)
	fieldCount := 4

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

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("core/network/info"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("core/network/info")); err != nil {
		return err
	}

	// t.Metadata (nimona.Metadata) (struct)
	if len("metadata") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"metadata\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("metadata"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("metadata")); err != nil {
		return err
	}

	if err := t.Metadata.MarshalCBOR(cw); err != nil {
		return err
	}

	// t.NetworkID (nimona.NetworkID) (struct)
	if len("networkID") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"networkID\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("networkID"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("networkID")); err != nil {
		return err
	}

	if err := t.NetworkID.MarshalCBOR(cw); err != nil {
		return err
	}

	// t.PeerAddresses ([]nimona.PeerAddr) (slice)
	if len("peerAddresses") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"peerAddresses\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("peerAddresses"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("peerAddresses")); err != nil {
		return err
	}

	if len(t.PeerAddresses) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.PeerAddresses was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajArray, uint64(len(t.PeerAddresses))); err != nil {
		return err
	}
	for _, v := range t.PeerAddresses {
		if err := v.MarshalCBOR(cw); err != nil {
			return err
		}
	}

	// t.RawBytes ([]uint8) (slice) - ignored

	return nil
}

func (t *NetworkInfo) UnmarshalCBORBytes(b []byte) (err error) {
	*t = NetworkInfo{}
	t.RawBytes = b
	return t.UnmarshalCBOR(bytes.NewReader(b))
}

func (t *NetworkInfo) UnmarshalCBOR(r io.Reader) (err error) {
	if t == nil {
		*t = NetworkInfo{}
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
		return fmt.Errorf("NetworkInfo: map struct too large (%d)", extra)
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
		case "metadata":

			{

				if err := t.Metadata.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.Metadata: %w", err)
				}

			}
			// t.NetworkID (nimona.NetworkID) (struct)
		case "networkID":

			{

				if err := t.NetworkID.UnmarshalCBOR(cr); err != nil {
					return xerrors.Errorf("unmarshaling t.NetworkID: %w", err)
				}

			}
			// t.PeerAddresses ([]nimona.PeerAddr) (slice)
		case "peerAddresses":

			maj, extra, err = cr.ReadHeader()
			if err != nil {
				return fmt.Errorf("t.PeerAddresses readHeader: %w", err)
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.PeerAddresses: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.PeerAddresses = make([]PeerAddr, extra)
			}

			for i := 0; i < int(extra); i++ {

				var v PeerAddr
				if err := v.UnmarshalCBOR(cr); err != nil {
					return err
				}

				t.PeerAddresses[i] = v
			}

			// t.RawBytes ([]uint8) (slice) - ignored

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
