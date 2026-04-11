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
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrGithubRepoNotFound        = errors.New("github repo not found")
	ErrGithubRepoNoReleases      = errors.New("github repo has no releases yet")
	ErrGithubForbidden           = errors.New("403 forbidden response from github api")
	ErrInvalidGithubAPIToken     = errors.New("invalid github api token in server configuration")
)

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
