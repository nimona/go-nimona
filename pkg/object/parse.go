package object

import "regexp"

var (
	regexObjectType = regexp.MustCompile(
		`(?m)` +
			`(?:(?P<type>[a-zA-Z0-9]+):)?` +
			`(?:(?P<namespace>[a-zA-Z0-9\.]+)/)?` +
			`(?P<object>[a-zA-Z0-9\.]+)`,
	)
)

type (
	ParsedType struct {
		PrimaryType string
		Namespace   string
		Object      string
	}
)

func ParseType(objectType string) ParsedType {
	match := regexObjectType.FindStringSubmatch(objectType)
	if len(match) < 3 {
		return ParsedType{}
	}
	return ParsedType{
		PrimaryType: match[1],
		Namespace:   match[2],
		Object:      match[3],
	}
}
