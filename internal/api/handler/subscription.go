package handler

import (
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/service"
)

type subscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubcriptionHandler(service *service.SubscriptionService) *subscriptionHandler {
	return &subscriptionHandler{service: service}
}

func (h *subscriptionHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {

}
