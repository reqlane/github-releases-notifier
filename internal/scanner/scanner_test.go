package scanner

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	mockrepository "github.com/reqlane/github-releases-notifier/internal/mock/repository"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helpers
const (
	ANY = mock.Anything
)

func setupMocks() (*mockrepository.SubscriptionRepo, *mockgithubapi.GithubClient, *mocknotifier.Notifier) {
	return new(mockrepository.SubscriptionRepo),
		new(mockgithubapi.GithubClient),
		new(mocknotifier.Notifier)
}

func newScanner(r *mockrepository.SubscriptionRepo, g *mockgithubapi.GithubClient, n *mocknotifier.Notifier) *FixedRateScanner {
	return &FixedRateScanner{
		repo:           r,
		githubClient:   g,
		notif:          n,
		logger:         zerolog.New(io.Discard), // no logs in tests
		requestsPerMin: defaultRequestsPerMin,
		sleepOnEmpty:   defaultSleepOnEmpty,
		pauseCh:        make(chan time.Time, 3),
		sleepFn:        func(time.Duration) {}, // no sleep in tests
	}
}

func TestScanner_Scan(t *testing.T) {
	t.Run("shouldn't call github when no repos are subscribed", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		repo.On("GetSubscribedRepos").Return([]model.Repo{}, nil).Once()

		s.scan()

		ghclient.AssertNotCalled(t, "GetLatestRelease", ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("shouldn't call github on database error", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		repo.On("GetSubscribedRepos").Return(nil, errors.New("db error")).Once()

		s.scan()

		ghclient.AssertNotCalled(t, "GetLatestRelease", ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

func TestScanner_CheckRepo(t *testing.T) {
	t.Run("should notify all and update tag if a new release is found", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		inputs := []model.Repo{
			{ID: 1, LastSeenTag: "v1.0.0"},
			{ID: 2, LastSeenTag: ""},
		}
		targets := []model.NotificationTarget{
			{Email: "user1@example.com", UnsubscribeToken: "token1"},
			{Email: "user2@example.com", UnsubscribeToken: "token2"},
			{Email: "user3@example.com", UnsubscribeToken: "token3"},
		}
		newRelease := "v2.0.0"

		for _, input := range inputs {
			ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
			repo.
				On("UpdateLastSeenTag", input.ID, &newRelease).Return(nil).Once().
				On("GetNotificationTargetsByRepo", input.ID).Return(targets, nil).Once()
			for _, target := range targets {
				notif.On("SendNotification", target.Email, input.Repo, newRelease, target.UnsubscribeToken).Return(nil).Once()
			}
			s.checkRepo(input)
		}

		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should neither update tag neither notify if release tag is unchanged", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}

		ghclient.On("GetLatestRelease", input.Repo).Return(&input.LastSeenTag, nil).Once()

		s.checkRepo(input)

		repo.AssertNotCalled(t, "UpdateLastSeenTag", ANY, ANY)
		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should neither update tag neither notify if no releases found", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}

		ghclient.On("GetLatestRelease", input.Repo).Return((*string)(nil), nil).Once()

		s.checkRepo(input)

		repo.AssertNotCalled(t, "UpdateLastSeenTag", ANY, ANY)
		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should not notify if tag update fails", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}
		newRelease := "v2.0.0"

		ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
		repo.On("UpdateLastSeenTag", input.ID, &newRelease).Return(errors.New("db error")).Once()

		s.checkRepo(input)

		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should skip notification if targets fetching fails", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}
		newRelease := "v2.0.0"

		ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
		repo.
			On("UpdateLastSeenTag", input.ID, &newRelease).Return(nil).Once().
			On("GetNotificationTargetsByRepo", input.ID).Return(nil, errors.New("db error")).Once()

		s.checkRepo(input)

		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should continue notifying others if one fails", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}
		targets := []model.NotificationTarget{
			{Email: "user1@example.com", UnsubscribeToken: "token1"},
			{Email: "user2@example.com", UnsubscribeToken: "token2"},
			{Email: "user3@example.com", UnsubscribeToken: "token3"},
		}
		newRelease := "v2.0.0"

		ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
		repo.
			On("UpdateLastSeenTag", input.ID, &newRelease).Return(nil).Once().
			On("GetNotificationTargetsByRepo", input.ID).Return(targets, nil).Once()
		notif.
			On("SendNotification", targets[0].Email, input.Repo, newRelease, ANY).Return(errors.New("smtp error")).Once().
			On("SendNotification", ANY, input.Repo, newRelease, ANY).Return(nil).Twice()

		s.checkRepo(input)

		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

func TestScanner_RateLimiting(t *testing.T) {
	t.Run("should send pause signal if github api returns rate limit error", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo"}
		resetTime := time.Now().Add(5 * time.Minute)

		ghclient.On("GetLatestRelease", input.Repo).Return((*string)(nil), &apperror.ErrGithubAPIRateLimited{ResetTime: resetTime}).Once()

		s.checkRepo(input)

		select {
		case received := <-s.pauseCh:
			assert.Equal(t, resetTime, received)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for pause signal in pauseCh")
		}

		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}
