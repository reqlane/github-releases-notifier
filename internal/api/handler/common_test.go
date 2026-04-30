package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	mockusecase "github.com/reqlane/github-releases-notifier/internal/mock/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// --- SubscriptionHandler ---
type subscriptionUsecaseMockBehaviour func(m *mockusecase.SubscriptionUseCase)

func subscriptionHandlerMocks() *mockusecase.SubscriptionUseCase {
	return new(mockusecase.SubscriptionUseCase)
}

func newSubcriptionHandler(usecase *mockusecase.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{
		usecase: usecase,
		logger:  zerolog.New(io.Discard),
	}
}

// --- Common ---
func setupRouter(h *SubscriptionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	rt := gin.New()
	api := rt.Group("/api")
	api.POST("/subscribe", h.SubscribeHandler)
	api.GET("/confirm/:token", h.ConfirmHandler)
	api.GET("/unsubscribe/:token", h.UnsubscribeHandler)
	api.GET("/subscriptions", h.GetSubscriptionsHandler)
	return rt
}

func performRequest(t *testing.T, engine *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if assert.NoError(t, err) {
			reqBody = bytes.NewBuffer(b)
		}
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}
