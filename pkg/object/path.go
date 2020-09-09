package object

import (
	"fmt"
	"strconv"
	"strings"

	"nimona.io/pkg/errors"
)

const (
	pathSeperator = "/"
)

func pathJoin(parts ...string) string {
	return strings.Join(parts, pathSeperator)
}

func pathSplit(path string) []string {
	return strings.Split(path, pathSeperator)
}

func resolvePath(path string, v Value) (Value, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}
	parts := pathSplit(path)
	var gv Value
	switch c := v.(type) {
	case Map:
		if c.m == nil {
			return nil, errors.New("invalid key, " + parts[0])
		}
		cc := c.m.value(parts[0])
		if cc == nil {
			return nil, errors.New("invalid key, " + parts[0])
		}
		gv = cc
	case List:
		i, err := strconv.Atoi(parts[0])
		if err != nil || i < 0 {
			return nil, errors.New("invalid index, " + parts[0])
		}
		ci := 0
		c.iterate(func(cv Value) bool {
			if ci == i {
				gv = cv
				return false
			}
			ci++
			return true
		})
		if gv == nil {
			return nil, errors.New("invalid index, " + parts[0])
		}
	default:
		return nil, errors.New("invalid path")
	}
	if len(parts) == 1 {
		return gv, nil
	}
	return resolvePath(pathJoin(parts[1:]...), gv)
}

func setPath(target Value, path string, value Value) (Value, error) {
	if path == "" {
		return target, nil
	}
	parts := pathSplit(path)
	switch t := target.(type) {
	case Map:
		if len(parts) == 1 {
			return t.set(parts[0], value), nil
		}
		newTarget := t.Value(parts[0])
		if newTarget == nil {
			hs := getHints(parts[0])
			switch hs[0] {
			case HintMap:
				newTarget = Map{}
			default:
				// TODO do we need to implement more cases, maybe lists?
				return nil, fmt.Errorf(
					"empty target for type %s is not supported", hs[0],
				)
			}
		}
		nv, err := setPath(newTarget, pathJoin(parts[1:]...), value)
		if err != nil {
			return nil, err
		}
		return t.set(parts[0], nv), nil

	case List:
		i, err := strconv.Atoi(parts[0])
		if err != nil || i < 0 {
			return nil, errors.New("invalid index, " + parts[0])
		}
		if len(parts) == 1 {
			return t.set(i, value), nil
		}
		nv, err := setPath(t.Get(i), pathJoin(parts[1:]...), value)
		if err != nil {
			return nil, err
		}
		return t.set(i, nv), nil
	}
	return nil, errors.New("invalid path")
}
