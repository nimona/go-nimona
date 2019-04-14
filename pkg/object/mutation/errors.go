package mutation

import (
	"errors"
)

var (
	ErrParsingCursor     = errors.New("could not parse cursor")
	ErrNotImplemented    = errors.New("operation not implemented")
	ErrApplyingOperation = errors.New("could not apply operation")
)
