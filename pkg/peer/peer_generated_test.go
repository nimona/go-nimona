package peer

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	c := &ConnectionInfo{}
	b, _ := json.Marshal(c.ToObject().ToMap())
	fmt.Println(string(b))
}
