package router

import (
	"net/http"
)

func (r *Router) subscriptionRouter(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscribe", r.subscriptionHandler.SubscribeHandler)
	mux.HandleFunc("GET /confirm/{token}", r.subscriptionHandler.ConfirmHandler)
	mux.HandleFunc("GET /unsubscribe/{token}", r.subscriptionHandler.UnsubscribeHandler)
	mux.HandleFunc("GET /subscriptions", r.subscriptionHandler.GetSubscriptionsHandler)
}
