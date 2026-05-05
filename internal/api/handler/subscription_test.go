package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	mockusecase "github.com/reqlane/github-releases-notifier/internal/mock/usecase"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SubscribeHandler ---
func TestSubscriptionHandler_SubscribeHandler(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name                        string
		email                       string
		repo                        string
		mocksBehaviour              handlerMocksBehaviour
		expectedStatusCode          int
		expectedResponseDetailsKeys []string
		expectRetryAfter            bool
	}{
		{
			name:  "success",
			email: "user@example.com",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "owner/repo"}).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:  "subscription already exists",
			email: "user@example.com",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "owner/repo"}).
					Return(apperror.ErrSubscriptionAlreadyExists).Once()
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:  "invalid email format",
			email: "not_an_email",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "not_an_email", Repo: "owner/repo"}).
					Return(&apperror.ErrValidation{
						Errs: []apperror.ErrField{
							{Field: "email", Constraint: "email", Value: "not_an_email"},
						},
					}).Once()
			},
			expectedStatusCode:          http.StatusBadRequest,
			expectedResponseDetailsKeys: []string{"email"},
		},
		{
			name:  "invalid repo format",
			email: "user@example.com",
			repo:  "not_a_repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "not_a_repo"}).
					Return(&apperror.ErrValidation{
						Errs: []apperror.ErrField{
							{Field: "repo", Constraint: "github_repo", Value: "not_a_repo"},
						},
					}).Once()
			},
			expectedStatusCode:          http.StatusBadRequest,
			expectedResponseDetailsKeys: []string{"repo"},
		},
		{
			name:  "invalid email and repo formats",
			email: "not_an_email",
			repo:  "not_a_repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "not_an_email", Repo: "not_a_repo"}).
					Return(&apperror.ErrValidation{
						Errs: []apperror.ErrField{
							{Field: "email", Constraint: "email", Value: "not_an_email"},
							{Field: "repo", Constraint: "github_repo", Value: "not_a_repo"},
						},
					}).Once()
			},
			expectedStatusCode:          http.StatusBadRequest,
			expectedResponseDetailsKeys: []string{"email", "repo"},
		},
		{
			name:  "repo not found",
			email: "user@example.com",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "owner/repo"}).
					Return(&apperror.ErrResourceNotFound{}).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:  "rate limited",
			email: "user@example.com",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "owner/repo"}).
					Return(&apperror.ErrGithubAPIRateLimited{ResetTime: time.Now().Add(time.Minute)}).Once()
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectRetryAfter:   true,
		},
		{
			name:  "internal server error",
			email: "user@example.com",
			repo:  "owner/repo",
			mocksBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Subscribe", &usecase.SubscribeInput{Email: "user@example.com", Repo: "owner/repo"}).
					Return(assert.AnError).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := subscriptionHandlerMocks()
			h := newSubcriptionHandler(uc)
			rt := setupRouter(h)

			body := map[string]string{
				"email": tt.email,
				"repo":  tt.repo,
			}

			tt.mocksBehaviour(uc)

			w := performRequest(t, rt, http.MethodPost, "/api/subscribe", body)

			require.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.expectedResponseDetailsKeys != nil {
				var resp APIResponse
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				for _, key := range tt.expectedResponseDetailsKeys {
					assert.Contains(t, resp.Details, key)
				}
			}
			if tt.expectRetryAfter {
				retryAfter := w.Header().Get("Retry-After")
				assert.NotEmpty(t, retryAfter)
				_, err := strconv.Atoi(retryAfter)
				assert.NoError(t, err)
			}
			uc.AssertExpectations(t)
		})
	}

	t.Run("invalid json body", func(t *testing.T) {
		t.Parallel()

		uc := subscriptionHandlerMocks()
		h := newSubcriptionHandler(uc)
		rt := setupRouter(h)

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString("not json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		rt.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		uc.AssertExpectations(t)
	})
}

// --- ConfirmHandler ---
func TestSubscriptionHandler_ConfirmHandler(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name                 string
		confirmToken         string
		usecaseMockBehaviour handlerMocksBehaviour
		expectedStatusCode   int
	}{
		{
			name:         "success",
			confirmToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Confirm", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:         "invalid token format",
			confirmToken: "abc123123123",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Confirm", "abc123123123").
					Return(&apperror.ErrInvalidResource{}).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "token not found",
			confirmToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Confirm", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(&apperror.ErrResourceNotFound{}).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:         "internal server error",
			confirmToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Confirm", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(assert.AnError).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := subscriptionHandlerMocks()
			h := newSubcriptionHandler(uc)
			rt := setupRouter(h)

			tt.usecaseMockBehaviour(uc)

			w := performRequest(t, rt, http.MethodGet, "/api/confirm/"+tt.confirmToken, nil)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			uc.AssertExpectations(t)
		})
	}
}

// --- UnsubscribeHandler ---
func TestSubscriptionHandler_UnsubscribeHandler(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name                 string
		unsubscribeToken     string
		usecaseMockBehaviour handlerMocksBehaviour
		expectedStatusCode   int
	}{
		{
			name:             "success",
			unsubscribeToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Unsubscribe", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "invalid token format",
			unsubscribeToken: "abc123123123",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Unsubscribe", "abc123123123").
					Return(&apperror.ErrInvalidResource{}).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "token not found",
			unsubscribeToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Unsubscribe", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(&apperror.ErrResourceNotFound{}).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:             "internal server error",
			unsubscribeToken: "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("Unsubscribe", "32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1").
					Return(assert.AnError).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := subscriptionHandlerMocks()
			h := newSubcriptionHandler(uc)
			rt := setupRouter(h)

			tt.usecaseMockBehaviour(uc)

			w := performRequest(t, rt, http.MethodGet, "/api/unsubscribe/"+tt.unsubscribeToken, nil)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			uc.AssertExpectations(t)
		})
	}
}

// --- GetSubscriptionsHandler ---
func TestSubscriptionHandler_GetSubscriptionsHandler(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name                  string
		email                 string
		usecaseMockBehaviour  handlerMocksBehaviour
		expectedStatusCode    int
		expectedSubscriptions []model.Subscription
	}{
		{
			name:  "success",
			email: "user@example.com",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("GetSubscriptions", "user@example.com").
					Return([]model.Subscription{
						{Email: "user@example.com", Repo: "owner1/repo1", Confirmed: true, LastSeenTag: "v1.0.0"},
						{Email: "user@example.com", Repo: "owner2/repo2", Confirmed: true, LastSeenTag: ""},
						{Email: "user@example.com", Repo: "owner3/repo3", Confirmed: true, LastSeenTag: "v3.0.0"},
					}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedSubscriptions: []model.Subscription{
				{Email: "user@example.com", Repo: "owner1/repo1", Confirmed: true, LastSeenTag: "v1.0.0"},
				{Email: "user@example.com", Repo: "owner2/repo2", Confirmed: true, LastSeenTag: ""},
				{Email: "user@example.com", Repo: "owner3/repo3", Confirmed: true, LastSeenTag: "v3.0.0"},
			},
		},
		{
			name:  "empty result",
			email: "user@example.com",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("GetSubscriptions", "user@example.com").
					Return([]model.Subscription{}, nil).Once()
			},
			expectedStatusCode:    http.StatusOK,
			expectedSubscriptions: []model.Subscription{},
		},
		{
			name:  "invalid email format",
			email: "not_an_email",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("GetSubscriptions", "not_an_email").
					Return(nil, &apperror.ErrInvalidResource{}).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:  "internal server error",
			email: "user@example.com",
			usecaseMockBehaviour: func(m *mockusecase.SubscriptionUseCase) {
				m.On("GetSubscriptions", "user@example.com").
					Return(nil, assert.AnError).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := subscriptionHandlerMocks()
			h := newSubcriptionHandler(uc)
			rt := setupRouter(h)

			tt.usecaseMockBehaviour(uc)

			w := performRequest(t, rt, http.MethodGet, "/api/subscriptions?email="+tt.email, nil)

			require.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.expectedSubscriptions != nil {
				var subscriptions []model.Subscription
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &subscriptions))
				assert.Equal(t, tt.expectedSubscriptions, subscriptions)
			}
			uc.AssertExpectations(t)
		})
	}
}
