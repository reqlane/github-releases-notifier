package router

import (
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/api/repository"
	"github.com/reqlane/github-releases-notifier/internal/api/service"
)

func (a *app) subscriptionRouter(mux *http.ServeMux) {
	subscriptionRepository := repository.NewSubcriptionRepository(a.db)
	subscriptionService := service.NewSubcriptionService(subscriptionRepository)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionService)

	mux.HandleFunc("POST /subscribe", subscriptionHandler.SubscribeHandler)
	mux.HandleFunc("GET /confirm/{token}", subscriptionHandler.ConfirmHandler)
	mux.HandleFunc("GET /unsubscribe/{token}", subscriptionHandler.UnsubscribeHandler)
	mux.HandleFunc("GET /subscriptions", subscriptionHandler.GetSubscriptionsHandler)
}
