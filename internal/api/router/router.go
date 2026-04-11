package router

import (
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/handler"
)

type app struct {
	subscriptionHandler *handler.SubscriptionHandler
}

func NewApp(h *handler.SubscriptionHandler) *app {
	return &app{
		subscriptionHandler: h,
	}
}

func (a *app) Router() *http.ServeMux {
	mux := http.NewServeMux()

	apiMux := http.NewServeMux()
	a.subscriptionRouter(apiMux)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	return mux
}
