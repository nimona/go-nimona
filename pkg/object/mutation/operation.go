package mutation

import (
	"strings"

	"github.com/joeycumines/go-dotnotation/dotnotation"

	"nimona.io/internal/errors"
	"nimona.io/pkg/object"
)

const (
	// OpAppend ...
	OpAppend = "append"
	// OpAssign ...
	OpAssign = "assign"
	// OpDelete ...
	OpDelete = "delete"
)

type (
	// Operation ...
	Operation struct {
		Operation string      `json:"op:s"`
		Cursor    []string    `json:"cursor:as"`
		Value     interface{} `json:"value"`
	}
)

// Apply ...
func (c Operation) Apply(o *object.Object) error {
	path := strings.Join(c.Cursor, ".")
	m := o.ToMap()

	v, err := dotnotation.Get(m, path)
	if err != nil {
		return errors.Wrap(ErrParsingCursor, err)
	}

	switch c.Operation {
	case OpAppend:
		vc, ok := v.([]interface{})
		if !ok {
			return ErrApplyingOperation
		}
		vc = append(vc, c.Value)
		err := dotnotation.Set(m, path, vc)
		if err != nil {
			return errors.Wrap(ErrApplyingOperation, err)
		}

	case OpAssign:
		err := dotnotation.Set(m, path, c.Value)
		if err != nil {
			return errors.Wrap(ErrApplyingOperation, err)
		}

	case OpDelete:
		return ErrNotImplemented

	default:
		return ErrNotImplemented
	}

	err = o.FromMap(m)
	if err != nil {
		return errors.Wrap(ErrApplyingOperation, err)
	}

	return nil
}
