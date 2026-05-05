package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/api/handler/dto"
	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/mock"
)

// --- POST /api/subscribe ---
func (s *IntegrationTestSuite) TestAPI_Subscribe_Success() {
	body := dto.SubscribeRequest{
		Email: "user@example.com",
		Repo:  "owner/repo",
	}
	latestRelease := "v1.0.0"

	s.ghclient.On("RepoExists", body.Repo).Return(nil).Once()
	s.ghclient.On("GetLatestRelease", body.Repo).Return(latestRelease, nil)
	s.notif.On("SendConfirmation", body.Email, body.Repo, mock.AnythingOfType("model.SubscriptionTokens")).Return(nil).Once()

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Equal(http.StatusOK, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_AlreadyExists() {
	body := dto.SubscribeRequest{
		Email: "user@example.com",
		Repo:  "owner/repo",
	}

	s.seedSubscription(body.Email, body.Repo, false)

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Equal(http.StatusConflict, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_InvalidEmail() {
	body := dto.SubscribeRequest{
		Email: "not an email",
		Repo:  "owner/repo",
	}

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Require().Equal(http.StatusBadRequest, w.Code)
	var resp handler.APIResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Contains(resp.Details, "email")

	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_InvalidRepo() {
	body := dto.SubscribeRequest{
		Email: "user@example.com",
		Repo:  "not a repo",
	}

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Require().Equal(http.StatusBadRequest, w.Code)
	var resp handler.APIResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Contains(resp.Details, "repo")

	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_InvalidEmailAndRepo() {
	body := dto.SubscribeRequest{
		Email: "not an email",
		Repo:  "not a repo",
	}

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Require().Equal(http.StatusBadRequest, w.Code)
	var resp handler.APIResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Contains(resp.Details, "email")
	s.Contains(resp.Details, "repo")

	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_RepoNotFound() {
	body := dto.SubscribeRequest{
		Email: "user@example.com",
		Repo:  "owner/repo",
	}

	s.ghclient.On("RepoExists", body.Repo).Return(&apperror.ErrResourceNotFound{}).Once()

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Equal(http.StatusNotFound, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_GithubRateLimited() {
	body := dto.SubscribeRequest{
		Email: "user@example.com",
		Repo:  "owner/repo",
	}
	resetTime := time.Now().Add(time.Minute)

	s.ghclient.On("RepoExists", body.Repo).Return(&apperror.ErrGithubAPIRateLimited{ResetTime: resetTime}).Once()

	w := s.performRequest(http.MethodPost, "/api/subscribe", body)

	s.Equal(http.StatusServiceUnavailable, w.Code)
	s.NotEmpty(w.Header().Get("Retry-After"))
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_MissingFields() {
	w := s.performRequest(http.MethodPost, "/api/subscribe", map[string]string{})

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Subscribe_InvalidJSONBody() {
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

// --- GET /api/confirm/:token ---
func (s *IntegrationTestSuite) TestAPI_Confirm_Success() {
	confirmToken, unsubscribeToken := s.seedSubscription("user@example.com", "owner/repo", false)

	w := s.performRequest(http.MethodGet, "/api/confirm/"+confirmToken, nil)

	s.Equal(http.StatusOK, w.Code)

	var fields struct {
		Confirmed    bool
		ConfirmToken *string
	}
	s.Require().NoError(s.db.Raw(
		`SELECT confirmed, confirm_token FROM subscriptions WHERE unsubscribe_token=?`,
		unsubscribeToken,
	).Scan(&fields).Error)

	s.True(fields.Confirmed)
	s.Nil(fields.ConfirmToken)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Confirm_InvalidToken() {
	w := s.performRequest(http.MethodGet, "/api/confirm/bad-token-format", nil)

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Confirm_TokenNotFound() {
	validToken := s.tokenGen.Generate()
	w := s.performRequest(http.MethodGet, "/api/confirm/"+validToken, nil)

	s.Equal(http.StatusNotFound, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Confirm_AlreadyConfirmed() {
	confirmToken, _ := s.seedSubscription("user@example.com", "owner/repo", false)

	w := s.performRequest(http.MethodGet, "/api/confirm/"+confirmToken, nil)
	s.Equal(http.StatusOK, w.Code)

	w = s.performRequest(http.MethodGet, "/api/confirm/"+confirmToken, nil)
	s.Equal(http.StatusNotFound, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

// --- GET /api/unsubscribe/:token ---
func (s *IntegrationTestSuite) TestAPI_Unsubscribe_Success() {
	_, unsubscribeToken := s.seedSubscription("user@example.com", "owner/repo", false)

	w := s.performRequest(http.MethodGet, "/api/unsubscribe/"+unsubscribeToken, nil)

	s.Equal(http.StatusOK, w.Code)

	var count int64
	s.Require().NoError(s.db.Raw(
		`SELECT COUNT(*) FROM subscriptions WHERE unsubscribe_token=?`,
		unsubscribeToken,
	).Scan(&count).Error)

	s.Equal(int64(0), count)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Unsubscribe_InvalidToken() {
	w := s.performRequest(http.MethodGet, "/api/unsubscribe/bad-token-format", nil)

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Unsubscribe_TokenNotFound() {
	validToken := s.tokenGen.Generate()
	w := s.performRequest(http.MethodGet, "/api/unsubscribe/"+validToken, nil)

	s.Equal(http.StatusNotFound, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_Unsubscribe_AlreadyConfirmed() {
	_, unsubscribeToken := s.seedSubscription("user@example.com", "owner/repo", false)

	w := s.performRequest(http.MethodGet, "/api/unsubscribe/"+unsubscribeToken, nil)
	s.Equal(http.StatusOK, w.Code)

	w = s.performRequest(http.MethodGet, "/api/unsubscribe/"+unsubscribeToken, nil)
	s.Equal(http.StatusNotFound, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

// --- GET /api/subscriptions ---
func (s *IntegrationTestSuite) TestAPI_GetSubscriptions_Success() {
	s.seedSubscription("user@example.com", "owner/repo1", true)
	s.seedSubscription("user@example.com", "owner/repo2", false)
	s.seedSubscription("user@example.com", "owner/repo3", false)
	s.seedSubscription("user@example.com", "owner/repo4", true)
	s.seedSubscription("user@example.com", "owner/repo5", true)

	w := s.performRequest(http.MethodGet, "/api/subscriptions?email=user@example.com", nil)

	s.Equal(http.StatusOK, w.Code)

	var subs []model.Subscription
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &subs))

	s.Len(subs, 5)
	for _, sub := range subs {
		s.Equal("user@example.com", sub.Email)
	}
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_GetSubscriptions_Empty() {
	w := s.performRequest(http.MethodGet, "/api/subscriptions?email=user@example.com", nil)

	s.Equal(http.StatusOK, w.Code)

	var subs []model.Subscription
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &subs))

	s.Empty(subs)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestAPI_GetSubscriptions_InvalidEmail() {
	w := s.performRequest(http.MethodGet, "/api/subscriptions?email=not-an-email", nil)

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}

func (s *IntegrationTestSuite) TestGetSubscriptions_MissingEmail() {
	w := s.performRequest(http.MethodGet, "/api/subscriptions", nil)

	s.Equal(http.StatusBadRequest, w.Code)
	mock.AssertExpectationsForObjects(s.T(), s.ghclient, s.notif)
}
