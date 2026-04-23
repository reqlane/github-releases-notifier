package router

import (
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/handler"
)

type Router struct {
	subscriptionHandler *handler.SubscriptionHandler
}

func NewRouter(h *handler.SubscriptionHandler) *Router {
	return &Router{
		subscriptionHandler: h,
	}
}

func (r *Router) Build() *http.ServeMux {
	mux := http.NewServeMux()

	apiMux := http.NewServeMux()
	r.subscriptionRouter(apiMux)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	return mux
}
