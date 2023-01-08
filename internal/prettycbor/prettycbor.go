package prettycbor

import (
	"bytes"
	"fmt"
	"strings"

	cbg "github.com/whyrusleeping/cbor-gen"
)

// Dump returns the given CBOR bytes in a human readable format similar to CBOR.me
func Dump(b []byte) string {
	r := cbg.NewCborReader(bytes.NewReader(b))
	w := &strings.Builder{}
	printMapWithHeader(w, r, 0)
	return w.String()
}

func printHeader(sb *strings.Builder, maj byte, n uint64, indent int) {
	w := &bytes.Buffer{}
	err := cbg.WriteMajorTypeHeader(w, maj, n)
	if err != nil {
		panic(fmt.Errorf("error writing header: %w", err))
	}
	t := ""
	switch maj {
	case cbg.MajUnsignedInt:
		t = "unsigned"
	case cbg.MajNegativeInt:
		t = "int"
	case cbg.MajByteString:
		t = "bytes"
	case cbg.MajTextString:
		t = "text"
	case cbg.MajArray:
		t = "array"
	case cbg.MajMap:
		t = "map"
	case cbg.MajTag:
		t = "tag"
	case cbg.MajOther:
		t = "other"
	}
	left := fmt.Sprintf("%X", w.Bytes())
	right := fmt.Sprintf("%s(%d)", t, n)
	printLine(sb, left, right, indent)
}

func printLine(sb *strings.Builder, left, right string, indent int) {
	left = fmt.Sprintf("%*s", indent+len(left), left)
	out := fmt.Sprintf("%-39s # %s\n", left, right)
	sb.WriteString(out)
}

func printValue(sb *strings.Builder, r *cbg.CborReader, maj byte, extra uint64, indent int) {
	switch maj {
	case cbg.MajUnsignedInt, cbg.MajNegativeInt:
		return
	}
	b := make([]byte, extra)
	_, err := r.Read(b)
	if err != nil {
		panic(fmt.Errorf("error reading value: %w", err))
	}
	// printHeader(sb, maj, extra, indent+3)
	printLine(sb, fmt.Sprintf("%X", b), fmt.Sprintf("%q", b), indent+3)
}

func printMapWithHeader(sb *strings.Builder, r *cbg.CborReader, indent int) {
	maj, n, err := r.ReadHeader()
	if err != nil {
		panic(fmt.Errorf("error reading header: %w", err))
	}

	if maj != cbg.MajMap {
		panic("can only print maps")
	}

	printHeader(sb, maj, n, 0)
	printMap(sb, r, n, indent)
}

func printMap(sb *strings.Builder, r *cbg.CborReader, extra uint64, indent int) {
	for i := uint64(0); i < extra; i++ {
		// read the key
		key, err := cbg.ReadString(r)
		if err != nil {
			panic(fmt.Errorf("error reading key: %w", err))
		}
		printHeader(sb, cbg.MajTextString, uint64(len(key)), indent+3)
		printLine(sb, fmt.Sprintf("%X", key), fmt.Sprintf("\"%s\"", key), indent+6)
		// read the value
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			panic(fmt.Errorf("error reading header: %w", err))
		}
		printHeader(sb, valMaj, extra, indent+3)
		switch valMaj {
		case cbg.MajMap:
			printMap(sb, r, extra, indent+3)
		case cbg.MajUnsignedInt:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajNegativeInt:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajByteString:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajTextString:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajArray:
			printArray(sb, r, extra, indent+3)
		// case cbg.MajTag:
		// 	messageHashTag(r, extra)
		case cbg.MajOther: // bool
			printValue(sb, r, valMaj, extra, indent+3)
		default:
			panic(fmt.Errorf("unhandled major type: %d", valMaj))
		}
	}
}

func printArray(sb *strings.Builder, r *cbg.CborReader, extra uint64, indent int) {
	for i := uint64(0); i < extra; i++ {
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			panic(fmt.Errorf("error reading header: %w", err))
		}
		printHeader(sb, valMaj, extra, indent+3)
		switch valMaj {
		case cbg.MajMap:
			printMap(sb, r, extra, indent+3)
		case cbg.MajUnsignedInt:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajNegativeInt:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajByteString:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajTextString:
			printValue(sb, r, valMaj, extra, indent+3)
		case cbg.MajArray:
			printArray(sb, r, extra, indent+3)
		// case cbg.MajTag:
		// 	hh, err = messageHashTag(r, extra)
		// case cbg.MajOther: // bool
		// 	hh, err = messageHashOther(r, extra)
		default:
			panic("unhandled major type, " + fmt.Sprintf("%d", valMaj))
		}
	}
}
