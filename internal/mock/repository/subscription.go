package mockrepository

import (
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/mock"
)

type SubscriptionRepo struct {
	mock.Mock
}

func (m *SubscriptionRepo) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *SubscriptionRepo) CreateSubscription(email string, repoID uint, tokens model.SubscriptionTokens) error {
	args := m.Called(email, repoID, tokens)
	return args.Error(0)
}

func (m *SubscriptionRepo) SubscriptionExists(email string, repoName string) (bool, error) {
	args := m.Called(email, repoName)
	return args.Bool(0), args.Error(1)
}

func (m *SubscriptionRepo) ConfirmSubscription(confirmToken string) error {
	args := m.Called(confirmToken)
	return args.Error(0)
}

func (m *SubscriptionRepo) DeleteSubscription(unsubscribeToken string) error {
	args := m.Called(unsubscribeToken)
	return args.Error(0)
}

func (m *SubscriptionRepo) GetOrCreateRepo(repoName string, lastSeenTag *string) (model.Repo, error) {
	args := m.Called(repoName, lastSeenTag)
	return args.Get(0).(model.Repo), args.Error(1)
}

func (m *SubscriptionRepo) GetSubscribedRepos() ([]model.Repo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Repo), args.Error(1)
}

func (m *SubscriptionRepo) GetNotificationTargetsByRepo(repoID uint) ([]model.NotificationTarget, error) {
	args := m.Called(repoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.NotificationTarget), args.Error(1)
}

func (m *SubscriptionRepo) UpdateLastSeenTag(repoID uint, tag *string) error {
	args := m.Called(repoID, tag)
	return args.Error(0)
}
