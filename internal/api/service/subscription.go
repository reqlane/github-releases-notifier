package service

import "github.com/reqlane/github-releases-notifier/internal/api/repository"

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubcriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}
