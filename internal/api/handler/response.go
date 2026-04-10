package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
)

type APIResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

func sendSuccess(w http.ResponseWriter, message string) {
	response := APIResponse{
		Status:  "success",
		Message: message,
	}
	sendJSON(w, http.StatusOK, response)
}

func sendError(w http.ResponseWriter, message string, code int) {
	response := APIResponse{
		Status:  "error",
		Message: message,
	}
	sendJSON(w, code, response)
}

func sendFromAppError(w http.ResponseWriter, err error) {
	response := APIResponse{Status: "error"}
	var code int

	if ev, ok := errors.AsType[*apperror.ErrValidation](err); ok {
		code = http.StatusBadRequest
		response.Message = "Validation failed"
		response.Details = make(map[string]string)
		for _, ef := range ev.Errs {
			switch ef.Constraint {
			case "required":
				response.Details[ef.Field] = "value is empty"
			case "email":
				response.Details[ef.Field] = fmt.Sprintf("%s is invalid email", ef.Value)
			case "github_repo":
				response.Details[ef.Field] = fmt.Sprintf("%s is invalid github repo, must be in owner/repo format", ef.Value)
			default:
				response.Details[ef.Field] = fmt.Sprintf("%s is invalid value", ef.Value)
			}
		}
	} else {
		code = http.StatusInternalServerError
		response.Message = "An unexpected error occured"
	}

	sendJSON(w, code, response)
}

func sendJSON(w http.ResponseWriter, code int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
