package apperror

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound                  = errors.New("not found")
	ErrAlreadyExists             = errors.New("resource already exists")
	ErrGithubForbidden           = errors.New("403 forbidden response from github api")
	ErrInvalidGithubAPIToken     = errors.New("invalid github api token in server configuration")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
)

// resource invalid
type ErrInvalidResource struct {
	Resource string
}

func (e *ErrInvalidResource) Error() string {
	return fmt.Sprintf("%s is invalid", e.Resource)
}

// resource not found
type ErrResourceNotFound struct {
	Resource string
}

func (e *ErrResourceNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

// github API rate limit exceeded
type ErrGithubAPIRateLimited struct {
	ResetTime time.Time
}

func (e *ErrGithubAPIRateLimited) Error() string {
	return fmt.Sprintf("github api rate limit exceeded, retry after %s", e.ResetTime.Format(time.RFC3339))
}

// validation
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
