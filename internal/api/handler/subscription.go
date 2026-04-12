package handler

import (
	"encoding/json"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/service"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/rs/zerolog"
)

type SubscriptionHandler struct {
	service   *service.SubscriptionService
	errLogger zerolog.Logger
}

func NewSubcriptionHandler(s *service.SubscriptionService, el zerolog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{service: s, errLogger: el}
}

func (h *SubscriptionHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	var subscribeRequest model.SubscribeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&subscribeRequest)
	if err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.Subscribe(&subscribeRequest)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Subscription successful. Confirmation email sent.")
}

func (h *SubscriptionHandler) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.service.Confirm(token)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Subscription confirmed successfully")
}

func (h *SubscriptionHandler) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.service.Unsubscribe(token)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Unsubscribed successfully")
}

func (h *SubscriptionHandler) GetSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	subscriptions, err := h.service.GetSubscriptions(email)
	if err != nil {
		h.sendFromAppError(w, err)
		return
	}

	sendJSON(w, http.StatusOK, subscriptions)
}
