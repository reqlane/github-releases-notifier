package router

import (
	"net/http"
)

func (a *app) subscriptionRouter(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscribe", a.subscriptionHandler.SubscribeHandler)
	mux.HandleFunc("GET /confirm/{token}", a.subscriptionHandler.ConfirmHandler)
	mux.HandleFunc("GET /unsubscribe/{token}", a.subscriptionHandler.UnsubscribeHandler)
	mux.HandleFunc("GET /subscriptions", a.subscriptionHandler.GetSubscriptionsHandler)
}
