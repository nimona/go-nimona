package encoding

import "strings"

func getPrimaryType(k string) string {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return ""
	}

	if len(ps[1]) == 0 {
		return ""
	}

	return ps[1][:1]
}

func getCleanKeyName(k string) string {
	ps := strings.Split(k, ":")
	return ps[0]
}

func getFullType(k string) string {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return ""
	}

	if len(ps[1]) == 0 {
		return ""
	}

	return ps[1]
}
