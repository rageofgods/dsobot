package data

import (
	"fmt"
	"strings"
)

// ContextError defines new struct for error handling to persist calling context
type ContextError struct {
	Context string
	Err     error
}

// Error declares custom error for ContextError struct
func (c *ContextError) Error() string {
	return fmt.Sprintf("%s->%v", strings.ToLower(c.Context), c.Err)
}

// CtxError returns error in specific context
func CtxError(cInfo string, err error) *ContextError {
	return &ContextError{
		Context: cInfo,
		Err:     err,
	}
}
