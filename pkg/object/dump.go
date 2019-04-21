package object

import "encoding/json"

// Dump returns the object as pretty-printed json
func Dump(o *Object, skipKeys ...string) string {
	m := o.ToMap()
	for _, skipKey := range skipKeys {
		delete(m, skipKey)
	}
	m["_hash"] = o.HashBase58()
	if sk := o.GetSignerKey(); sk != nil {
		m["_signer"] = sk.HashBase58()
	}
	j, _ := json.MarshalIndent(m, "", "  ")
	return string(j)
}
