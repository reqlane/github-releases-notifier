package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/reqlane/github-releases-notifier/internal/api/repository"
	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

type SubscriptionService struct {
	repo         *repository.SubscriptionRepository
	githubClient *githubapi.GithubClient
}

func NewSubcriptionService(repo *repository.SubscriptionRepository, githubClient *githubapi.GithubClient) *SubscriptionService {
	return &SubscriptionService{
		repo:         repo,
		githubClient: githubClient,
	}
}

func (s *SubscriptionService) Subscribe(req *model.SubscribeRequest) error {
	if err := validate.Struct(req); err != nil {
		return validationError(err)
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

	// TODO Send confirmation email

	return nil
}

func (s *SubscriptionService) Confirm(token string) error {
	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	return nil
}

func (s *SubscriptionService) GetSubscriptions(filter *model.SubscriptionFilter) ([]model.Subscription, error) {
	if err := validate.Struct(filter); err != nil {
		return nil, validationError(err)
	}

	subscriptions, err := s.repo.GetSubscriptionsByEmail(filter.Email)
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
