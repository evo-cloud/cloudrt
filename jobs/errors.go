package jobs

import "errors"

// Common errors
var (
	ErrTaskNonRevertable = errors.New("task is not revertable")
)
