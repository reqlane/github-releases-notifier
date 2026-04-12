package scanner

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/mock"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/rs/zerolog"
)

// Helpers
func newScanner(r *mock.MockRepository, g *mock.MockGithubClient, n *mock.MockNotifier) *FixedRateScanner {
	return &FixedRateScanner{
		repo:           r,
		githubClient:   g,
		notif:          n,
		logger:         zerolog.New(os.Stderr),
		requestsPerMin: defaultRequestsPerMin,
		sleepOnEmpty:   defaultSleepOnEmpty,
		pauseCh:        make(chan time.Time, 3),
		sleepFn:        func(time.Duration) {}, // no sleep in tests
	}
}

// scan tests
func TestScanNoReposNoGithubCalls(t *testing.T) {
	githubCalled := false

	repo := mock.IdealRepository()
	repo.GetSubscribedReposFunc = func() ([]model.Repo, error) {
		return []model.Repo{}, nil
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		githubCalled = true
		return "", nil
	}

	s := newScanner(repo, githubClient, mock.IdealNotifier())
	s.scan()

	if githubCalled {
		t.Error("expected no github calls for empty repo list")
	}
}

func TestScanDBErrorNoGithubCalls(t *testing.T) {
	githubCalled := false

	repo := mock.IdealRepository()
	repo.GetSubscribedReposFunc = func() ([]model.Repo, error) {
		return nil, errors.New("db error")
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		githubCalled = true
		return "", nil
	}

	s := newScanner(repo, githubClient, mock.IdealNotifier())
	s.scan()

	if githubCalled {
		t.Error("expected no github calls on db error")
	}
}

// checkRepo tests
func TestCheckRepoNewReleaseNotifiesAllUpdadtesTagEvenIfEmpty(t *testing.T) {
	updateCalled := false
	notifyCalled := 0

	repo := mock.IdealRepository()
	repo.UpdateLastSeenTagFunc = func(repoID int, tag string) error {
		updateCalled = true
		return nil
	}
	repo.GetNotificationTargetsByRepoFunc = func(repoID int) ([]model.NotificationTarget, error) {
		return []model.NotificationTarget{
			{Email: "user1@example.com", UnsubscribeToken: "token1"},
			{Email: "user2@example.com", UnsubscribeToken: "token2"},
			{Email: "user3@example.com", UnsubscribeToken: "token3"},
		}, nil
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		// new release
		return "v2.0.0", nil
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		notifyCalled++
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})
	s.checkRepo(model.Repo{LastSeenTag: ""})

	if !updateCalled {
		t.Error("expected UpdateLastSeenTag to be called")
	}
	if notifyCalled != 6 {
		t.Error("expected SendNotification to be called")
	}
}

func TestCheckRepoSameTagNoUpdateNoNotify(t *testing.T) {
	updateCalled := false
	notifyCalled := false

	repo := mock.IdealRepository()
	repo.UpdateLastSeenTagFunc = func(repoID int, tag string) error {
		updateCalled = true
		return nil
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		// same release
		return "v1.0.0", nil
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		notifyCalled = true
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})

	if updateCalled {
		t.Error("expected no UpdateLastSeenTag call for same tag")
	}
	if notifyCalled {
		t.Error("expected no SendNotification call for same tag")
	}
}

func TestCheckRepoNoReleasesNoUpdateNoNotify(t *testing.T) {
	updateCalled := false
	notifyCalled := false

	repo := mock.IdealRepository()
	repo.UpdateLastSeenTagFunc = func(repoID int, tag string) error {
		updateCalled = true
		return nil
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(string) (string, error) {
		return "", apperror.ErrGithubRepoNoReleases
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		notifyCalled = true
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})

	if updateCalled {
		t.Error("expected no UpdateLastSeenTag call when repo has no releases")
	}
	if notifyCalled {
		t.Error("expected no SendNotification call when repo has no releases")
	}
}

func TestCheckRepoRateLimitedSendsPauseSignal(t *testing.T) {
	resetTime := time.Now().Add(5 * time.Minute)

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		return "", &apperror.ErrGithubAPIRateLimited{ResetTime: resetTime}
	}

	s := newScanner(mock.IdealRepository(), githubClient, mock.IdealNotifier())
	s.checkRepo(model.Repo{})

	select {
	case received := <-s.pauseCh:
		if !received.Equal(resetTime) {
			t.Errorf("expected reset time %v, got %v", resetTime, received)
		}
	default:
		t.Error("expected pause signal in pauseCh, channel was empty")
	}
}

func TestCheckRepoUpdateTagFailsNoNotification(t *testing.T) {
	notifyCalled := false

	repo := mock.IdealRepository()
	repo.UpdateLastSeenTagFunc = func(repoID int, tag string) error {
		return errors.New("db error")
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(string) (string, error) {
		return "v2.0.0", nil
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		notifyCalled = true
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})

	if notifyCalled {
		t.Error("expected no notification when UpdateLastSeenTag fails")
	}
}

func TestCheckRepoGetTargetsFailsNoNotification(t *testing.T) {
	notifyCalled := false

	repo := mock.IdealRepository()
	repo.GetNotificationTargetsByRepoFunc = func(repoID int) ([]model.NotificationTarget, error) {
		return nil, errors.New("db error")
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		return "v2.0.0", nil
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		notifyCalled = true
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})

	if notifyCalled {
		t.Error("expected no notification when GetNotificationTargetsByRepo fails")
	}
}

func TestCheckRepoOneNotifyFailsOthersStillNotified(t *testing.T) {
	notifyCount := 0

	repo := mock.IdealRepository()
	repo.GetNotificationTargetsByRepoFunc = func(repoID int) ([]model.NotificationTarget, error) {
		return []model.NotificationTarget{
			{Email: "user1@example.com", UnsubscribeToken: "token1"},
			{Email: "user2@example.com", UnsubscribeToken: "token2"},
			{Email: "user3@example.com", UnsubscribeToken: "token3"},
		}, nil
	}

	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		return "v2.0.0", nil
	}

	notif := mock.IdealNotifier()
	notif.SendNotificationFunc = func(recipient, repo, tag, unsubscribeToken string) error {
		if recipient == "user1@example.com" {
			return errors.New("smtp error")
		}
		notifyCount++
		return nil
	}

	s := newScanner(repo, githubClient, notif)
	s.checkRepo(model.Repo{LastSeenTag: "v1.0.0"})

	if notifyCount != 2 {
		t.Errorf("expected 2 successful notifications, got %d", notifyCount)
	}
}
