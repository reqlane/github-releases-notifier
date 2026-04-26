package usecase

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/mock"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

// Helpers
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

func newService(r *mock.MockRepository, g *mock.MockGithubClient, n *mock.MockNotifier) SubscriptionUseCase {
	return NewSubscriptionUseCase(r, g, n)
}

// Subscribe tests
func TestSubscribeSuccess(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestSubscribeInvalidEmail(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for _, tt := range invalidEmails {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Subscribe(&SubscribeInput{Email: "not-an-email", Repo: "owner/repo"})
			if _, ok := errors.AsType[*apperror.ErrValidation](err); !ok {
				t.Errorf("expected *ErrValidation, got: %T", err)
			}
		})
	}
}

func TestSubscribeInvalidRepo(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for _, tt := range invalidRepos {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: tt.repo})
			if _, ok := errors.AsType[*apperror.ErrValidation](err); !ok {
				t.Errorf("expected *ErrValidation for %s, got: %T", tt.repo, err)
			}
		})
	}
}

func TestSubscribeSubscriptionAlreadyExists(t *testing.T) {
	repo := mock.IdealRepository()
	repo.SubscriptionExistsFunc = func(email string, repoName string) (bool, error) { return true, nil }

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if !errors.Is(err, apperror.ErrSubscriptionAlreadyExists) {
		t.Errorf("expected ErrSubscriptionAlreadyExists, got: %v", err)
	}
}

func TestSubscribeRepoNotFoundOnGithub(t *testing.T) {
	githubClient := mock.IdealGithubClient()
	githubClient.RepoExistsFunc = func(repo string) error {
		return &apperror.ErrResourceNotFound{Resource: "Github repository"}
	}

	svc := newService(mock.IdealRepository(), githubClient, mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if _, ok := errors.AsType[*apperror.ErrResourceNotFound](err); !ok {
		t.Errorf("expected *ErrResourceNotFound, got: %T", err)
	}
}

func TestSubscribeRepoHasNoReleasesSucceeds(t *testing.T) {
	githubClient := mock.IdealGithubClient()
	githubClient.GetLatestReleaseFunc = func(repo string) (string, error) {
		return "", apperror.ErrGithubRepoNoReleases
	}

	svc := newService(mock.IdealRepository(), githubClient, mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if err != nil {
		t.Errorf("expected no error for repo with no releases, got: %v", err)
	}
}

func TestSubscribeRepoNotInDBCreatesIt(t *testing.T) {
	createRepoCalled := false

	repo := mock.IdealRepository()
	repo.GetRepoByNameFunc = func(repoName string) (model.Repo, error) {
		return model.Repo{}, apperror.ErrNotFound
	}
	repo.CreateRepoFunc = func(r model.Repo) (model.Repo, error) {
		createRepoCalled = true
		return model.Repo{}, nil
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !createRepoCalled {
		t.Error("expected CreateRepo to be called")
	}
}

func TestSubscribeRepoRaceConditionFallsBackToGet(t *testing.T) {
	// CreateRepo fails with ErrAlreadyExists if race condition lost, then GetRepoByName is expected to be called
	repo := mock.IdealRepository()
	getRepoByNameCalls := 0
	repo.GetRepoByNameFunc = func(repoName string) (model.Repo, error) {
		getRepoByNameCalls++
		if getRepoByNameCalls == 1 {
			return model.Repo{}, apperror.ErrNotFound
		}
		return model.Repo{}, nil
	}
	repo.CreateRepoFunc = func(repo model.Repo) (model.Repo, error) {
		return model.Repo{}, apperror.ErrAlreadyExists
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if err != nil {
		t.Errorf("expected no error on race condition fallback, got: %v", err)
	}
	if getRepoByNameCalls != 2 {
		t.Errorf("expected GetRepoByName to be called twice, got %d", getRepoByNameCalls)
	}
}

func TestSubscribeConfirmationEmailFails(t *testing.T) {
	notif := mock.IdealNotifier()
	notif.SendConfirmationFunc = func(recipient, repo, confirmToken, unsubscribeToken string) error {
		return errors.New("smtp error")
	}

	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), notif)
	err := svc.Subscribe(&SubscribeInput{Email: "user@example.com", Repo: "owner/repo"})
	if err == nil {
		t.Error("expected error when email fails, got nil")
	}
}

// Confirm tests
func TestConfirmSuccess(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for i, token := range validTokens {
		t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
			err := svc.Confirm(token)
			if err != nil {
				t.Errorf("expected no error for token %q, got: %v", token, err)
			}
		})
	}
}

func TestConfirmInvalidToken(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for _, tt := range invalidTokens {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Confirm(tt.token)
			if _, ok := errors.AsType[*apperror.ErrInvalidResource](err); !ok {
				t.Errorf("expected *ErrInvalidResource for %s, got: %T", tt.token, err)
			}
		})
	}
}

func TestConfirmTokenNotFound(t *testing.T) {
	repo := &mock.MockRepository{
		ConfirmSubscriptionFunc: func(confirmToken string) error { return apperror.ErrNotFound },
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Confirm(validTokens[0])
	if _, ok := errors.AsType[*apperror.ErrResourceNotFound](err); !ok {
		t.Errorf("expected *ErrResourceNotFound, got: %T", err)
	}
}

// Unsubscribe tests
func TestUnsubscribeSuccess(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for i, token := range validTokens {
		t.Run(fmt.Sprintf("valid token %d", i+1), func(t *testing.T) {
			err := svc.Unsubscribe(token)
			if err != nil {
				t.Errorf("expected no error for token %q, got: %v", token, err)
			}
		})
	}
}

func TestUnsubscribeInvalidToken(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for _, tt := range invalidTokens {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Unsubscribe(tt.token)
			if _, ok := errors.AsType[*apperror.ErrInvalidResource](err); !ok {
				t.Errorf("expected *ErrInvalidResource for %s, got: %T", tt.token, err)
			}
		})
	}
}

func TestUnsubscribeTokenNotFound(t *testing.T) {
	repo := &mock.MockRepository{
		DeleteSubscriptionFunc: func(unsubscribeToken string) error { return apperror.ErrNotFound },
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	err := svc.Unsubscribe(validTokens[0])
	if _, ok := errors.AsType[*apperror.ErrResourceNotFound](err); !ok {
		t.Errorf("expected *ErrResourceNotFound, got: %T", err)
	}
}

// GetSubscriptions tests
func TestGetSubscriptionsSuccess(t *testing.T) {
	repo := &mock.MockRepository{
		GetSubscriptionsByEmailFunc: func(email string) ([]model.Subscription, error) {
			return []model.Subscription{
				{Email: "user@example.com", Repo: "owner1/repo1", Confirmed: true, LastSeenTag: "v.1.0.0"},
				{Email: "user@example.com", Repo: "owner2/repo2", Confirmed: false, LastSeenTag: "v.1.0.0"},
				{Email: "user@example.com", Repo: "owner3/repo3", Confirmed: false, LastSeenTag: "v.1.0.0"},
			}, nil
		},
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	subs, err := svc.GetSubscriptions("user@example.com")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("expected 3 subscription, got %d", len(subs))
	}
}

func TestGetSubscriptionsInvalidEmail(t *testing.T) {
	svc := newService(mock.IdealRepository(), mock.IdealGithubClient(), mock.IdealNotifier())

	for _, tt := range invalidEmails {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetSubscriptions(tt.email)
			if _, ok := errors.AsType[*apperror.ErrInvalidResource](err); !ok {
				t.Errorf("expected *ErrInvalidResource for %s, got: %T", tt.email, err)
			}
		})
	}
}

func TestGetSubscriptionsRepoError(t *testing.T) {
	repo := &mock.MockRepository{
		GetSubscriptionsByEmailFunc: func(email string) ([]model.Subscription, error) {
			return nil, errors.New("db error")
		},
	}

	svc := newService(repo, mock.IdealGithubClient(), mock.IdealNotifier())
	_, err := svc.GetSubscriptions("user@example.com")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
