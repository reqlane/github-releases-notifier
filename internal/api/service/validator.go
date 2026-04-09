package service

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate        = validator.New(validator.WithRequiredStructEnabled())
	githubRepoRegex = regexp.MustCompile(`(?i)^[a-z\d](([a-z\d-]?[a-z\d])*)?/[a-z\d_-][a-z\d_.-]*$`)
)

func init() {
	validate.RegisterValidation("github_repo", func(fl validator.FieldLevel) bool {
		valueString := fl.Field().String()
		parts := strings.Split(valueString, "/")
		if len(parts) != 2 || len(parts[0]) > 39 || len(parts[1]) > 100 {
			return false
		}
		return githubRepoRegex.MatchString(valueString)
	})

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			return field.Name
		}
		return strings.Split(jsonTag, ",")[0]
	})
}

func handleValidationError(err error) error {
	if ve, ok := err.(validator.ValidationErrors); ok {
		fieldErr := ve[0]
		switch ve[0].Tag() {
		case "required":
			return fmt.Errorf("%s is required", fieldErr.Field())
		case "email":
			return fmt.Errorf("%s is invalid email", fieldErr.Value())
		case "github_repo":
			return fmt.Errorf("%s is invalid repo, must be in owner/repo format", fieldErr.Value())
		default:
			return fmt.Errorf("invalid %s", fieldErr.Field())
		}
	}
	return err
}
