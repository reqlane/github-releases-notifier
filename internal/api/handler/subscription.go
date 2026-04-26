package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reqlane/github-releases-notifier/internal/api/handler/dto"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/rs/zerolog"
)

type SubscriptionHandler struct {
	usecase usecase.SubscriptionUseCase
	logger  zerolog.Logger
}

func NewSubcriptionHandler(s usecase.SubscriptionUseCase, l zerolog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{usecase: s, logger: l}
}

func (h *SubscriptionHandler) SubscribeHandler(c *gin.Context) {
	var input dto.SubscribeRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	err = h.usecase.Subscribe(&usecase.SubscribeInput{Email: input.Email, Repo: input.Repo})
	if err != nil {
		h.sendFromAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Subscription successful. Confirmation email sent.",
	})
}

func (h *SubscriptionHandler) ConfirmHandler(c *gin.Context) {
	token := c.Param("token")

	err := h.usecase.Confirm(token)
	if err != nil {
		h.sendFromAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Subscription confirmed successfully",
	})
}

func (h *SubscriptionHandler) UnsubscribeHandler(c *gin.Context) {
	token := c.Param("token")

	err := h.usecase.Unsubscribe(token)
	if err != nil {
		h.sendFromAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Unsubscribed successfully",
	})
}

func (h *SubscriptionHandler) GetSubscriptionsHandler(c *gin.Context) {
	email := c.Query("email")

	subscriptions, err := h.usecase.GetSubscriptions(email)
	if err != nil {
		h.sendFromAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}
