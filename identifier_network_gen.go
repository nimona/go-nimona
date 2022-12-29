// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package nimona

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = bytes.Compare
var _ = xerrors.Errorf
var _ = cid.Undef
var _ = math.E
var _ = sort.Sort

func (t *NetworkID) MarshalCBORBytes() ([]byte, error) {
	w := bytes.NewBuffer(nil)
	err := t.MarshalCBOR(w)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *NetworkID) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{161}); err != nil {
		return err
	}

	// t.Hostname (string) (string)
	if len("Hostname") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Hostname\" was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len("Hostname"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Hostname")); err != nil {
		return err
	}

	if len(t.Hostname) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.Hostname was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(t.Hostname))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.Hostname)); err != nil {
		return err
	}
	return nil
}

func (t *NetworkID) UnmarshalCBORBytes(b []byte) (err error) {
	return t.UnmarshalCBOR(bytes.NewReader(b))
}

func (t *NetworkID) UnmarshalCBOR(r io.Reader) (err error) {
	*t = NetworkID{}

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
		return fmt.Errorf("NetworkID: map struct too large (%d)", extra)
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
		// t.Hostname (string) (string)
		case "Hostname":

			{
				sval, err := cbg.ReadString(cr)
				if err != nil {
					return err
				}

				t.Hostname = string(sval)
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
