package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/reqlane/github-releases-notifier/internal/notifier"
	"github.com/reqlane/github-releases-notifier/internal/repository"
)

type SubscriptionService struct {
	repo         *repository.Repository
	githubClient *githubapi.GithubClient
	notif        *notifier.Notifier
}

func NewSubcriptionService(r *repository.Repository, g *githubapi.GithubClient, n *notifier.Notifier) *SubscriptionService {
	return &SubscriptionService{
		repo:         r,
		githubClient: g,
		notif:        n,
	}
}

func (s *SubscriptionService) Subscribe(req *model.SubscribeRequest) error {
	if err := validate.Struct(req); err != nil {
		return structValidationError(err)
	}

	// Check if email already subscribed on repo
	exists, err := s.repo.SubscriptionExists(req.Email, req.Repo)
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}
	if exists {
		return apperror.ErrSubscriptionAlreadyExists
	}

	// Check repo existence
	err = s.githubClient.RepoExists(req.Repo)
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}

	// Get latest release
	lastSeenTag, err := s.githubClient.GetLatestRelease(req.Repo)
	if err != nil && !errors.Is(err, apperror.ErrGithubRepoNoReleases) {
		return fmt.Errorf("service.Subscribe: %w", err)
	}

	// Create repo record if not tracked yet
	repo, err := s.repo.GetRepoByName(req.Repo)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return fmt.Errorf("service.Subscribe: %w", err)
	}
	if errors.Is(err, apperror.ErrNotFound) {
		repo, err = s.repo.CreateRepo(model.Repo{Repo: req.Repo, LastSeenTag: lastSeenTag})
		if err != nil {
			// possible race condition
			if errors.Is(err, apperror.ErrAlreadyExists) {
				repo, err = s.repo.GetRepoByName(req.Repo)
				if err != nil {
					return fmt.Errorf("service.Subscribe: %w", err)
				}
			} else {
				return fmt.Errorf("service.Subscribe: %w", err)
			}
		}
	}

	// Create subscription
	confirmToken, err := generateToken()
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}
	unsubscribeToken, err := generateToken()
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}
	err = s.repo.CreateSubscription(req.Email, repo.ID, confirmToken, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}

	// Send confirmation email
	err = s.notif.SendConfirmation(req.Email, req.Repo, confirmToken, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("service.Subscribe: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Confirm(token string) error {
	if !isValidToken(token) {
		return &apperror.ErrInvalidResource{Resource: "Token"}
	}

	err := s.repo.ConfirmSubscription(token)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return &apperror.ErrResourceNotFound{Resource: "Token"}
		}
		return fmt.Errorf("service.Confirm: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	if !isValidToken(token) {
		return &apperror.ErrInvalidResource{Resource: "Token"}
	}

	err := s.repo.DeleteSubscription(token)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return &apperror.ErrResourceNotFound{Resource: "Token"}
		}
		return fmt.Errorf("service.Unsubscribe: %w", err)
	}

	return nil
}

func (s *SubscriptionService) GetSubscriptions(email string) ([]model.Subscription, error) {
	if err := validate.Var(email, "email"); err != nil {
		return nil, &apperror.ErrInvalidResource{Resource: "Email"}
	}

	subscriptions, err := s.repo.GetSubscriptionsByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service.GetSubscriptions: %w", err)
	}

	return subscriptions, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isValidToken(token string) bool {
	b, err := hex.DecodeString(token)
	return err == nil && len(b) == 32
}
