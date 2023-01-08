# Pretty CBOR

Provides pretty-printed versions of binary CBOR similar CBOR.me.

```go
package prettycbor

import (
    "encoding/hex"
    "fmt"
    "strings"
    "testing"

    "github.com/fxamacker/cbor/v2"
    "github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
    m := map[string]interface{}{
        "foo": "bar",
        "baz": 42,
        "qux": []interface{}{
            "foo",
            "bar",
            "baz",
        },
        "quux": map[string]interface{}{
            "foo": "bar",
            "baz": 42,
        },
        "corge": []interface{}{
            map[string]interface{}{
                "foo": "bar",
                "baz": 42,
            },
        },
    }

    b, err := cbor.Marshal(m)
    require.NoError(t, err)

    fmt.Println(Dump(b))
}
```

```txt
A5                                      # map(5)
   63                                   # text(3)
      666F6F                            # "foo"
   63                                   # text(3)
      626172                            # "bar"
   63                                   # text(3)
      62617A                            # "baz"
   182A                                 # unsigned(42)
   63                                   # text(3)
      717578                            # "qux"
   83                                   # array(3)
      63                                # text(3)
         666F6F                         # "foo"
      63                                # text(3)
         626172                         # "bar"
      63                                # text(3)
         62617A                         # "baz"
   64                                   # text(4)
      71757578                          # "quux"
   A2                                   # map(2)
      63                                # text(3)
         62617A                         # "baz"
      182A                              # unsigned(42)
      63                                # text(3)
         666F6F                         # "foo"
      63                                # text(3)
         626172                         # "bar"
   65                                   # text(5)
      636F726765                        # "corge"
   81                                   # array(1)
      A2                                # map(2)
         63                             # text(3)
            62617A                      # "baz"
         182A                           # unsigned(42)
         63                             # text(3)
            666F6F                      # "foo"
         63                             # text(3)
            626172                      # "bar"
```
