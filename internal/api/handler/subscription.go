package handler

import (
	"encoding/json"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/handler/dto"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/rs/zerolog"
)

type SubscriptionHandler struct {
	usecase usecase.SubscriptionUseCase
	logger  zerolog.Logger
}

func NewSubcriptionHandler(s usecase.SubscriptionUseCase, l zerolog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{usecase: s, logger: l}
}

func (h *SubscriptionHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.SubscribeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.usecase.Subscribe(&usecase.SubscribeInput{Email: req.Email, Repo: req.Repo})
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	h.sendSuccess(w, "Subscription successful. Confirmation email sent.")
}

func (h *SubscriptionHandler) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.usecase.Confirm(token)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	h.sendSuccess(w, "Subscription confirmed successfully")
}

func (h *SubscriptionHandler) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.usecase.Unsubscribe(token)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	h.sendSuccess(w, "Unsubscribed successfully")
}

func (h *SubscriptionHandler) GetSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	subscriptions, err := h.usecase.GetSubscriptions(email)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, subscriptions)
}
