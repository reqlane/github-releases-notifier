package apperror

import (
	"errors"
	"fmt"
	"strings"
)

// invalid github API token
var ErrInvalidGithubAPIToken = errors.New("invalid github api token in server configuration")

// github repo not found
type ErrGithubRepoNotFound struct {
	Repo string
}

func (e *ErrGithubRepoNotFound) Error() string {
	return fmt.Sprintf("github repo %s not found", e.Repo)
}

// github API rate limit exceeded
type ErrGithubAPIRateLimited struct {
	RetryAfter string
}

func (e *ErrGithubAPIRateLimited) Error() string {
	if e.RetryAfter == "" {
		return "github api rate limit exceeded, retry later"
	}
	return fmt.Sprintf("github api rate limit exceeded, retry after %s seconds", e.RetryAfter)
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
