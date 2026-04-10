package router

import (
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/api/handler"
)

type app struct {
	subscriptionHandler *handler.SubscriptionHandler
}

func NewApp(subscriptionHandler *handler.SubscriptionHandler) *app {
	return &app{
		subscriptionHandler: subscriptionHandler,
	}
}

func (a *app) Router() *http.ServeMux {
	mux := http.NewServeMux()

	a.subscriptionRouter(mux)

	return mux
}
