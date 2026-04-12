package mock

import "github.com/reqlane/github-releases-notifier/internal/model"

type MockRepository struct {
	GetSubscriptionsByEmailFunc      func(email string) ([]model.Subscription, error)
	CreateSubscriptionFunc           func(email string, repoID int, confirmToken, unsubscribeToken string) error
	SubscriptionExistsFunc           func(email string, repoName string) (bool, error)
	ConfirmSubscriptionFunc          func(confirmToken string) error
	DeleteSubscriptionFunc           func(unsubscribeToken string) error
	GetRepoByNameFunc                func(repoName string) (model.Repo, error)
	CreateRepoFunc                   func(repo model.Repo) (model.Repo, error)
	GetSubscribedReposFunc           func() ([]model.Repo, error)
	GetNotificationTargetsByRepoFunc func(repoID int) ([]model.NotificationTarget, error)
	UpdateLastSeenTagFunc            func(repoID int, tag string) error
}

func (m *MockRepository) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
	return m.GetSubscriptionsByEmailFunc(email)
}

func (m *MockRepository) CreateSubscription(email string, repoID int, confirmToken, unsubscribeToken string) error {
	return m.CreateSubscriptionFunc(email, repoID, confirmToken, unsubscribeToken)
}

func (m *MockRepository) SubscriptionExists(email string, repoName string) (bool, error) {
	return m.SubscriptionExistsFunc(email, repoName)
}

func (m *MockRepository) ConfirmSubscription(confirmToken string) error {
	return m.ConfirmSubscriptionFunc(confirmToken)
}

func (m *MockRepository) DeleteSubscription(unsubscribeToken string) error {
	return m.DeleteSubscriptionFunc(unsubscribeToken)
}

func (m *MockRepository) GetRepoByName(repoName string) (model.Repo, error) {
	return m.GetRepoByNameFunc(repoName)
}

func (m *MockRepository) CreateRepo(repo model.Repo) (model.Repo, error) {
	return m.CreateRepoFunc(repo)
}

func (m *MockRepository) GetSubscribedRepos() ([]model.Repo, error) {
	return m.GetSubscribedReposFunc()
}

func (m *MockRepository) GetNotificationTargetsByRepo(repoID int) ([]model.NotificationTarget, error) {
	return m.GetNotificationTargetsByRepoFunc(repoID)
}

func (m *MockRepository) UpdateLastSeenTag(repoID int, tag string) error {
	return m.UpdateLastSeenTagFunc(repoID, tag)
}

func IdealRepository() *MockRepository {
	return &MockRepository{
		GetSubscriptionsByEmailFunc:      func(email string) ([]model.Subscription, error) { return []model.Subscription{}, nil },
		CreateSubscriptionFunc:           func(email string, repoID int, confirmToken, unsubscribeToken string) error { return nil },
		SubscriptionExistsFunc:           func(email, repoName string) (bool, error) { return false, nil },
		ConfirmSubscriptionFunc:          func(confirmToken string) error { return nil },
		DeleteSubscriptionFunc:           func(unsubscribeToken string) error { return nil },
		GetRepoByNameFunc:                func(repoName string) (model.Repo, error) { return model.Repo{}, nil },
		CreateRepoFunc:                   func(repo model.Repo) (model.Repo, error) { return model.Repo{}, nil },
		GetSubscribedReposFunc:           func() ([]model.Repo, error) { return []model.Repo{}, nil },
		GetNotificationTargetsByRepoFunc: func(repoID int) ([]model.NotificationTarget, error) { return []model.NotificationTarget{}, nil },
		UpdateLastSeenTagFunc:            func(repoID int, tag string) error { return nil },
	}
}
