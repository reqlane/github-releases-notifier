package service

import (
	"fmt"

	"github.com/reqlane/github-releases-notifier/internal/api/repository"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubcriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Subscribe(req *model.SubscribeRequest) error {
	if err := validate.Struct(req); err != nil {
		return validationError(err)
	}
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

	subscriptions, err := s.repo.GetByEmail(filter.Email)
	if err != nil {
		return nil, fmt.Errorf("service.GetSubscriptions: %w", err)
	}
	return subscriptions, nil
}
