package mockusecase

import (
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/stretchr/testify/mock"
)

type SubscriptionUseCase struct {
	mock.Mock
}

func (m *SubscriptionUseCase) Subscribe(input *usecase.SubscribeInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *SubscriptionUseCase) Confirm(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *SubscriptionUseCase) Unsubscribe(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *SubscriptionUseCase) GetSubscriptions(email string) ([]model.Subscription, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Subscription), args.Error(1)
}
