package usecase

import (
	"fmt"
	"testing"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Subscribe ---
func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
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

	t.Run("invalid email format", func(t *testing.T) {
		t.Parallel()

		for _, tt := range invalidEmails {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)
				err := usecase.Subscribe(&SubscribeInput{
					Email: tt.email,
					Repo:  "owner/repo",
				})
				assert.ErrorAs(t, err, new(*apperror.ErrValidation))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("invalid repo format", func(t *testing.T) {
		t.Parallel()

		for _, tt := range invalidRepos {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)
				err := usecase.Subscribe(&SubscribeInput{
					Email: "user@example.com",
					Repo:  tt.repo,
				})
				assert.ErrorAs(t, err, new(*apperror.ErrValidation))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("subscription already exists", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
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

	t.Run("repo not found on github", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
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

	t.Run("repo has no releases yet", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
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
			On("GetLatestRelease", input.Repo).Return(noReleases, nil).Once()
		repo.
			On("GetOrCreateRepo", input.Repo, noReleases).Return(createdRepo, nil).
			On("CreateSubscription", input.Email, createdRepo.ID, ANY).Return(nil).Once()
		notif.On("SendConfirmation", input.Email, input.Repo, ANY).Return(nil).Once()

		err := usecase.Subscribe(input)
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

// --- Confirm ---
func TestSubscriptionUseCase_Confirm(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for i, token := range validTokens {
			t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)

				repo.On("ConfirmSubscription", token).Return(nil).Once()

				err := usecase.Confirm(token)
				assert.NoError(t, err)
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		t.Parallel()

		for _, tt := range invalidTokens {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)

				err := usecase.Confirm(tt.token)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("token not found", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		repo.On("ConfirmSubscription", validTokens[0]).Return(apperror.ErrNotFound).Once()

		err := usecase.Confirm(validTokens[0])
		assert.ErrorAs(t, err, new(*apperror.ErrResourceNotFound))
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

// --- Unsubscribe ---
func TestSubscriptionUseCase_Unsubscribe(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for i, token := range validTokens {
			t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)

				repo.On("DeleteSubscription", token).Return(nil).Once()

				err := usecase.Unsubscribe(token)
				assert.NoError(t, err)
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		t.Parallel()

		for _, tt := range invalidTokens {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)

				err := usecase.Unsubscribe(tt.token)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})

	t.Run("token not found", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		repo.On("DeleteSubscription", validTokens[0]).Return(apperror.ErrNotFound).Once()

		err := usecase.Unsubscribe(validTokens[0])
		assert.ErrorAs(t, err, new(*apperror.ErrResourceNotFound))
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

// --- GetSubscriptions ---
func TestSubscriptionUseCase_GetSubscriptions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := subscriptionUseCaseMocks()
		usecase := newSubscriptionUseCase(repo, ghclient, notif)

		input := "user@example.com"
		dbsubs := []model.Subscription{
			{Email: "user@example.com", Repo: "owner1/repo1", Confirmed: true, LastSeenTag: "v.1.0.0"},
			{Email: "user@example.com", Repo: "owner2/repo2", Confirmed: false, LastSeenTag: ""},
			{Email: "user@example.com", Repo: "owner3/repo3", Confirmed: false, LastSeenTag: "v.3.0.0"},
		}

		repo.On("GetSubscriptionsByEmail", input).Return(dbsubs, nil).Once()

		subs, err := usecase.GetSubscriptions(input)

		assert.NoError(t, err)
		assert.Equal(t, dbsubs, subs)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("invalid email format", func(t *testing.T) {
		t.Parallel()

		for _, tt := range invalidEmails {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, ghclient, notif := subscriptionUseCaseMocks()
				usecase := newSubscriptionUseCase(repo, ghclient, notif)

				_, err := usecase.GetSubscriptions(tt.email)
				assert.ErrorAs(t, err, new(*apperror.ErrInvalidResource))
				mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
			})
		}
	})
}
