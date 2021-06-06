package object

import "github.com/mitchellh/copystructure"

func Copy(s *Object) *Object {
	r, err := copystructure.Copy(s)
	if err != nil {
		panic(err)
	}
	return r.(*Object)
}
