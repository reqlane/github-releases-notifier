package usecase

import (
	"fmt"
	"strings"
	"testing"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	mockrepository "github.com/reqlane/github-releases-notifier/internal/mock/repository"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helpers
const (
	ANY = mock.Anything
)

var (
	invalidEmails = []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"missing @", "userexample.com"},
		{"missing domain", "user@"},
		{"missing local part", "@example.com"},
		{"spaces", "user @example.com"},
	}
	invalidRepos = []struct {
		name string
		repo string
	}{
		{"missing slash", "ownerrepo"},
		{"owner too long", strings.Repeat("a", 40) + "/repo"},
		{"repo name too long", "owner/" + strings.Repeat("a", 101)},
		{"starts with hyphen", "-owner/repo"},
		{"double slash", "owner//repo"},
	}
	validTokens = []string{
		"32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
		"aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
		"aabbcca4752b3a1f5ddd5cddaabbccdb1f5fa568335fa56833eeeeeeeebbccdd",
	}
	invalidTokens = []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"too short", "aabbccdd"},
		{"too long", "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd00"},
		{"non-hex characters", "zzbbaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabb"},
		{"correct length but spaces", "aabbccddaabbccdd aabbccddaabbccddaabbccddaabbccddaabbccddaabbcc"},
	}
)

func setupMocks() (*mockrepository.SubscriptionRepo, *mockgithubapi.GithubClient, *mocknotifier.Notifier) {
	return new(mockrepository.SubscriptionRepo),
		new(mockgithubapi.GithubClient),
		new(mocknotifier.Notifier)
}

func newSubscriptionUseCase(r *mockrepository.SubscriptionRepo, g *mockgithubapi.GithubClient, n *mocknotifier.Notifier) SubscriptionUseCase {
	return NewSubscriptionUseCase(r, g, n)
}

func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	t.Run("subscribe success path", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := &SubscribeInput{
			Email: "user@example.com",
			Repo:  "owner/repo",
		}
		lastSeenTag := "1.0.0"
		repoID := uint(100)

		// subscription existence
		repo.On("SubscriptionExists", input.Email, input.Repo).Return(false, nil).Once()

		// github api calls
		ghclient.
			On("RepoExists", input.Repo).Return(nil).Once().
			On("GetLatestRelease", input.Repo).Return(&lastSeenTag, nil).Once()

		// database creation
		repo.
			On("GetOrCreateRepo", input.Repo, &lastSeenTag).Return(model.Repo{ID: repoID}, nil).Once().
			On("CreateSubscription", input.Email, repoID, ANY, ANY).Return(nil).Once()

		// confirmation email
		notif.On("SendConfirmation", input.Email, input.Repo, ANY, ANY).Return(nil).Once()

		err := usecase.Subscribe(input)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should return validation error if email format is invalid", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)
		for _, tt := range invalidEmails {
			t.Run(tt.name, func(t *testing.T) {
				err := usecase.Subscribe(&SubscribeInput{
					Email: tt.email,
					Repo:  "owner/repo",
				})
				assert.ErrorAs(t, err, new(*apperror.ErrValidation))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return validation error if repo format is invalid", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)
		for _, tt := range invalidRepos {
			t.Run(tt.name, func(t *testing.T) {
				err := usecase.Subscribe(&SubscribeInput{
					Email: "user@example.com",
					Repo:  tt.repo,
				})
				assert.ErrorAs(t, err, new(*apperror.ErrValidation))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return specific error if subscription already exists", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := &SubscribeInput{
			Email: "user@example.com",
			Repo:  "owner/repo",
		}

		repo.On("SubscriptionExists", input.Email, input.Repo).Return(true, nil).Once()

		err := usecase.Subscribe(input)
		assert.ErrorIs(t, err, apperror.ErrSubscriptionAlreadyExists)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should return specific error if repo not found on github", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := &SubscribeInput{
			Email: "user@example.com",
			Repo:  "owner/repo",
		}

		repo.On("SubscriptionExists", input.Email, input.Repo).Return(false, nil).Once()
		ghclient.On("RepoExists", input.Repo).Return(&apperror.ErrResourceNotFound{}).Once()

		err := usecase.Subscribe(input)
		assert.ErrorAs(t, err, new(*apperror.ErrResourceNotFound))
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("no error if repo has no releases yet", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := &SubscribeInput{
			Email: "user@example.com",
			Repo:  "owner/repo",
		}
		var noReleases *string
		createdRepo := model.Repo{ID: 100, Repo: input.Repo, LastSeenTag: ""}

		repo.On("SubscriptionExists", input.Email, input.Repo).Return(false, nil).Once()
		ghclient.
			On("RepoExists", input.Repo).Return(nil).Once().
			On("GetLatestRelease", input.Repo).Return(noReleases, apperror.ErrGithubRepoNoReleases).Once()
		repo.
			On("GetOrCreateRepo", input.Repo, noReleases).Return(createdRepo, nil).
			On("CreateSubscription", input.Email, createdRepo.ID, ANY).Return(nil).Once()
		notif.On("SendConfirmation", input.Email, input.Repo, ANY).Return(nil).Once()

		err := usecase.Subscribe(input)
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

func TestSubscriptionUseCase_Confirm(t *testing.T) {
	t.Run("confirm success with valid token", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		for i, token := range validTokens {
			t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
				repo.On("ConfirmSubscription", token).Return(nil).Once()
				err := usecase.Confirm(token)
				assert.NoError(t, err)
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return specific error if token format is invalid", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		for _, tt := range invalidTokens {
			t.Run(tt.name, func(t *testing.T) {
				err := usecase.Confirm(tt.token)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return specific error if token not found", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		repo.On("ConfirmSubscription", validTokens[0]).Return(apperror.ErrNotFound).Once()

		err := usecase.Confirm(validTokens[0])
		assert.ErrorAs(t, err, new(*apperror.ErrResourceNotFound))
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

func TestSubscriptionUseCase_Unsubscribe(t *testing.T) {
	t.Run("unsubscribe success with valid token", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		for i, token := range validTokens {
			t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
				repo.On("DeleteSubscription", token).Return(nil).Once()

				err := usecase.Unsubscribe(token)
				assert.NoError(t, err)
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return specific error if token format is invalid", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		for _, tt := range invalidTokens {
			t.Run(tt.name, func(t *testing.T) {
				err := usecase.Unsubscribe(tt.token)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("should return specific error if token not found", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		repo.On("DeleteSubscription", validTokens[0]).Return(apperror.ErrNotFound).Once()

		err := usecase.Unsubscribe(validTokens[0])
		assert.ErrorAs(t, err, new(*apperror.ErrResourceNotFound))
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

func TestSubscriptionUseCase_GetSubscriptions(t *testing.T) {
	t.Run("get subscriptions success path", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := "user@example.com"
		dbsubs := []model.Subscription{
			{Email: "user@example.com", Repo: "owner1/repo1", Confirmed: true, LastSeenTag: "v.1.0.0"},
			{Email: "user@example.com", Repo: "owner2/repo2", Confirmed: false, LastSeenTag: "v.1.0.0"},
			{Email: "user@example.com", Repo: "owner3/repo3", Confirmed: false, LastSeenTag: "v.1.0.0"},
		}

		repo.On("GetSubscriptionsByEmail", input).Return(dbsubs, nil).Once()

		subs, err := usecase.GetSubscriptions(input)

		assert.NoError(t, err)
		assert.Equal(t, dbsubs, subs)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("should return specific error if email format is invalid", func(t *testing.T) {
		repo, ghclient, notif := setupMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		for _, tt := range invalidEmails {
			t.Run(tt.name, func(t *testing.T) {
				_, err := usecase.GetSubscriptions(tt.email)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})
}
