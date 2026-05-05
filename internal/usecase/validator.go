package usecase

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/reqlane/github-releases-notifier/internal/apperror"
)

var (
	validate        = validator.New(validator.WithRequiredStructEnabled())
	githubRepoRegex = regexp.MustCompile(`(?i)^[a-z\d](([a-z\d-]?[a-z\d])*)?/[a-z\d_.-][a-z\d_.-]*$`)
)

func init() {
	err := validate.RegisterValidation("github_repo", func(fl validator.FieldLevel) bool {
		valueString := fl.Field().String()
		parts := strings.Split(valueString, "/")
		if len(parts) != 2 || len(parts[0]) > 39 || len(parts[1]) > 100 {
			return false
		}
		return githubRepoRegex.MatchString(valueString)
	})
	if err != nil {
		panic(err)
	}

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("label")
	})
}

func structValidationError(err error) error {
	if ve, ok := err.(validator.ValidationErrors); ok {
		errVal := &apperror.ErrValidation{}
		for _, fe := range ve {
			errVal.Errs = append(errVal.Errs, apperror.ErrField{
				Field:      fe.Field(),
				Value:      fe.Value(),
				Constraint: fe.Tag(),
			})
		}
		return errVal
	}
	return err
}
