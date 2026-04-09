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
		// TODO handling
		http.Error(w, "400", http.StatusBadRequest)
		return
	}

	subscription, err := h.service.Subscribe(&subscribeRequest)
	if err != nil {
		// TODO handling
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string             `json:"status"`
		Message string             `json:"message"`
		Data    model.Subscription `json:"data"`
	}{
		Status:  "success",
		Message: "Subscription successful. Confirmation email sent.",
		Data:    *subscription,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *subscriptionHandler) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	subscription, err := h.service.Confirm(token)
	if err != nil {
		// TODO handling
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string             `json:"status"`
		Message string             `json:"message"`
		Data    model.Subscription `json:"data"`
	}{
		Status:  "success",
		Message: "Subscription confirmed successfully",
		Data:    *subscription,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *subscriptionHandler) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	unsubscribedData, err := h.service.Unsubscribe(token)
	if err != nil {
		// TODO handling
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string                        `json:"status"`
		Message string                        `json:"message"`
		Data    model.UnsubscribeResponseData `json:"data"`
	}{
		Status:  "success",
		Message: "Unsubscribed successfully",
		Data:    *unsubscribedData,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *subscriptionHandler) GetSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	subscriptions, err := h.service.GetSubscriptions(email)
	if err != nil {
		// TODO handling
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string               `json:"status"`
		Message string               `json:"message"`
		Data    []model.Subscription `json:"data"`
	}{
		Status:  "success",
		Message: "Unsubscribed successfully",
		Data:    subscriptions,
	}
	json.NewEncoder(w).Encode(response)
}
