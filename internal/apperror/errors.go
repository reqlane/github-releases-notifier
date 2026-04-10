package apperror

import (
	"fmt"
	"strings"
)

type ErrField struct {
	Field      string
	Value      any
	Constraint string
}

type ErrValidation struct {
	Errs []ErrField
}

func (ev *ErrValidation) Error() string {
	var msgs []string
	for _, ef := range ev.Errs {
		msgs = append(msgs, fmt.Sprintf("%s '%s' failed validation on %s", ef.Field, ef.Value, ef.Constraint))
	}
	return strings.Join(msgs, ";\n")
}
