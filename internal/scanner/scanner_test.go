package scanner

import (
	"testing"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- scan ---
func TestScanner_Scan(t *testing.T) {
	t.Parallel()

	t.Run("no subscribed repos", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		repo.On("GetSubscribedRepos").Return([]model.Repo{}, nil).Once()

		s.scan()

		ghclient.AssertNotCalled(t, "GetLatestRelease", ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		repo.On("GetSubscribedRepos").Return(nil, assert.AnError).Once()

		s.scan()

		ghclient.AssertNotCalled(t, "GetLatestRelease", ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

// --- checkRepo ---
func TestScanner_CheckRepo(t *testing.T) {
	t.Parallel()

	t.Run("a new release is found", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
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

	t.Run("release tag is unchanged", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}

		ghclient.On("GetLatestRelease", input.Repo).Return(&input.LastSeenTag, nil).Once()

		s.checkRepo(input)

		repo.AssertNotCalled(t, "UpdateLastSeenTag", ANY, ANY)
		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("no releases found", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}

		ghclient.On("GetLatestRelease", input.Repo).Return(nil, nil).Once()

		s.checkRepo(input)

		repo.AssertNotCalled(t, "UpdateLastSeenTag", ANY, ANY)
		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("tag update fails", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}
		newRelease := "v2.0.0"

		ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
		repo.On("UpdateLastSeenTag", input.ID, &newRelease).Return(assert.AnError).Once()

		s.checkRepo(input)

		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("targets fetching fails", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo", LastSeenTag: "v1.0.0"}
		newRelease := "v2.0.0"

		ghclient.On("GetLatestRelease", input.Repo).Return(&newRelease, nil).Once()
		repo.
			On("UpdateLastSeenTag", input.ID, &newRelease).Return(nil).Once().
			On("GetNotificationTargetsByRepo", input.ID).Return(nil, assert.AnError).Once()

		s.checkRepo(input)

		notif.AssertNotCalled(t, "SendNotification", ANY, ANY, ANY, ANY)
		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})

	t.Run("notifications fail partially", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
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
			On("SendNotification", targets[0].Email, input.Repo, newRelease, ANY).Return(assert.AnError).Once().
			On("SendNotification", ANY, input.Repo, newRelease, ANY).Return(nil).Twice()

		s.checkRepo(input)

		mock.AssertExpectationsForObjects(t, repo, ghclient, notif)
	})
}

// --- Rate limited ---
func TestScanner_RateLimited(t *testing.T) {
	t.Parallel()

	t.Run("github api returns rate limit error", func(t *testing.T) {
		t.Parallel()

		repo, ghclient, notif := scannerMocks()
		s := newScanner(repo, ghclient, notif)

		input := model.Repo{Repo: "owner/repo"}
		resetTime := time.Now().Add(5 * time.Minute)

		ghclient.On("GetLatestRelease", input.Repo).Return(nil, &apperror.ErrGithubAPIRateLimited{ResetTime: resetTime}).Once()

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
