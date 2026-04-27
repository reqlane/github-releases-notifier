package router

import (
	"github.com/gin-gonic/gin"
)

func (r *Router) subscriptionRouter(rg *gin.RouterGroup) {
	rg.POST("/subscribe", r.subscriptionHandler.SubscribeHandler)
	rg.GET("/confirm/:token", r.subscriptionHandler.ConfirmHandler)
	rg.GET("/unsubscribe/:token", r.subscriptionHandler.UnsubscribeHandler)
	rg.GET("/subscriptions", r.subscriptionHandler.GetSubscriptionsHandler)
}
