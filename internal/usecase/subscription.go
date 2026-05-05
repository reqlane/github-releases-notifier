package usecase

import (
	"encoding/hex"
	"errors"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/contract"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/reqlane/github-releases-notifier/pkg/tokengen"
)

type subscriptionUseCase struct {
	repo         contract.SubscriptionRepo
	githubClient contract.GithubClient
	notif        contract.Notifier
	tokenGen     tokengen.Generator
}

func NewSubscriptionUseCase(r contract.SubscriptionRepo, g contract.GithubClient, n contract.Notifier) SubscriptionUseCase {
	return &subscriptionUseCase{
		repo:         r,
		githubClient: g,
		notif:        n,
		tokenGen:     tokengen.NewRandGenerator(32),
	}
}

func (s *subscriptionUseCase) Subscribe(input *SubscribeInput) error {
	if err := validate.Struct(input); err != nil {
		return structValidationError(err)
	}

	// Check if email already subscribed on repo
	exists, err := s.repo.SubscriptionExists(input.Email, input.Repo)
	if err != nil {
		return err
	}
	if exists {
		return apperror.ErrSubscriptionAlreadyExists
	}

	// Check repo existence
	err = s.githubClient.RepoExists(input.Repo)
	if err != nil {
		return err
	}

	// Get latest release
	lastSeenTag, err := s.githubClient.GetLatestRelease(input.Repo)
	if err != nil {
		return err
	}

	// Create repo record if not tracked yet
	trackedRepo, err := s.repo.GetOrCreateRepo(input.Repo, lastSeenTag)
	if err != nil {
		return err
	}

	// Create subscription
	tokens := model.SubscriptionTokens{
		ConfirmToken:     s.tokenGen.Generate(),
		UnsubscribeToken: s.tokenGen.Generate(),
	}
	err = s.repo.CreateSubscription(input.Email, trackedRepo.ID, tokens)
	if err != nil {
		return err
	}

	// Send confirmation email
	err = s.notif.SendConfirmation(input.Email, input.Repo, tokens)
	if err != nil {
		return err
	}

	return nil
}

func (s *subscriptionUseCase) Confirm(token string) error {
	if !isValidToken(token) {
		return &apperror.ErrInvalidResource{Resource: "Token"}
	}

	err := s.repo.ConfirmSubscription(token)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return &apperror.ErrResourceNotFound{Resource: "Token"}
		}
		return err
	}

	return nil
}

func (s *subscriptionUseCase) Unsubscribe(token string) error {
	if !isValidToken(token) {
		return &apperror.ErrInvalidResource{Resource: "Token"}
	}

	err := s.repo.DeleteSubscription(token)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return &apperror.ErrResourceNotFound{Resource: "Token"}
		}
		return err
	}

	return nil
}

func (s *subscriptionUseCase) GetSubscriptions(email string) ([]model.Subscription, error) {
	if err := validate.Var(email, "email"); err != nil {
		return nil, &apperror.ErrInvalidResource{Resource: "Email"}
	}

	subscriptions, err := s.repo.GetSubscriptionsByEmail(email)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func isValidToken(token string) bool {
	b, err := hex.DecodeString(token)
	return err == nil && len(b) == 32
}
