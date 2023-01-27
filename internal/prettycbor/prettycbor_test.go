package prettycbor

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	hexString := `a465247479706571636f72652f73747265616d2f706174636869246d65746164617461a1656f776e657263666f6f6c646570656e64656e6369657381826c6e696d6f6e613a2f2f646f6343666f6f6a6f7065726174696f6e7382a3626f70636164646470617468662f696e7436346576616c7565182aa3626f70677265706c6163656470617468672f737472696e676576616c756563626172`

	expString := `
A4                                      # map(4)
   65                                   # text(5)
      2474797065                        # "$type"
   71                                   # text(17)
      636F72652F73747265616D2F7061746368 # "core/stream/patch"
   69                                   # text(9)
      246D65746164617461                # "$metadata"
   A1                                   # map(1)
      65                                # text(5)
         6F776E6572                     # "owner"
      63                                # text(3)
         666F6F                         # "foo"
   6C                                   # text(12)
      646570656E64656E63696573          # "dependencies"
   81                                   # array(1)
      82                                # array(2)
         6C                             # text(12)
            6E696D6F6E613A2F2F646F63    # "nimona://doc"
         43                             # bytes(3)
            666F6F                      # "foo"
   6A                                   # text(10)
      6F7065726174696F6E73              # "operations"
   82                                   # array(2)
      A3                                # map(3)
         62                             # text(2)
            6F70                        # "op"
         63                             # text(3)
            616464                      # "add"
         64                             # text(4)
            70617468                    # "path"
         66                             # text(6)
            2F696E743634                # "/int64"
         65                             # text(5)
            76616C7565                  # "value"
         182A                           # unsigned(42)
      A3                                # map(3)
         62                             # text(2)
            6F70                        # "op"
         67                             # text(7)
            7265706C616365              # "replace"
         64                             # text(4)
            70617468                    # "path"
         67                             # text(7)
            2F737472696E67              # "/string"
         65                             # text(5)
            76616C7565                  # "value"
         63                             # text(3)
            626172                      # "bar"
`

	bytes, err := hex.DecodeString(hexString)
	require.NoError(t, err)

	out := SPrint(bytes)
	require.Equal(t, strings.TrimSpace(expString), strings.TrimSpace(out))
}
