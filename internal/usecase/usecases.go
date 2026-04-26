package usecase

import "github.com/reqlane/github-releases-notifier/internal/model"

type SubscriptionUseCase interface {
	Subscribe(input *SubscribeInput) error
	Confirm(token string) error
	Unsubscribe(token string) error
	GetSubscriptions(email string) ([]model.Subscription, error)
}
