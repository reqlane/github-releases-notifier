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
	return nil
}

func (s *SubscriptionService) Confirm(token string) error {
	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	return nil
}

func (s *SubscriptionService) GetSubscriptions(email string) ([]model.Subscription, error) {
	// TODO validate email
	subscriptions, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service.GetSubscriptions: %w", err)
	}
	return subscriptions, nil
}
