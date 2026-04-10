package handler

import (
	"encoding/json"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/service"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

type subscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubcriptionHandler(service *service.SubscriptionService) *subscriptionHandler {
	return &subscriptionHandler{service: service}
}

func (h *subscriptionHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
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
		sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Subscription successful. Confirmation email sent.")
}

func (h *subscriptionHandler) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.service.Confirm(token)
	if err != nil {
		sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Subscription confirmed successfully")
}

func (h *subscriptionHandler) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	err := h.service.Unsubscribe(token)
	if err != nil {
		sendFromAppError(w, err)
		return
	}

	sendSuccess(w, "Unsubscribed successfully")
}

func (h *subscriptionHandler) GetSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	filter := &model.SubscriptionFilter{Email: r.URL.Query().Get("email")}

	subscriptions, err := h.service.GetSubscriptions(filter)
	if err != nil {
		sendFromAppError(w, err)
		return
	}

	sendJSON(w, http.StatusOK, subscriptions)
}
