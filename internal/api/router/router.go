package router

import (
	"github.com/gin-gonic/gin"
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

func (r *Router) Build() *gin.Engine {
	engine := gin.Default()

	api := engine.Group("/api")
	r.subscriptionRouter(api)

	return engine
}
