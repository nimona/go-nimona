package object

import "encoding/json"

// Dump returns the object as pretty-printed json
func Dump(o Object, skipKeys ...string) string {
	m := o.ToMap()
	for _, skipKey := range skipKeys {
		delete(m, skipKey)
	}
	m["_hash"] = o.Hash().String()
	// TODO(NOW): add sig chain as _xxx?
	j, _ := json.MarshalIndent(m, "", "  ")
	return string(j)
}
