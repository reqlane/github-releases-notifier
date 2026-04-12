package mock

import "github.com/reqlane/github-releases-notifier/internal/model"

type MockRepository struct {
	getSubscriptionsByEmail      func(email string) ([]model.Subscription, error)
	createSubscription           func(email string, repoID int, confirmToken, unsubscribeToken string) error
	subscriptionExists           func(email string, repoName string) (bool, error)
	confirmSubscription          func(confirmToken string) error
	deleteSubscription           func(unsubscribeToken string) error
	getRepoByName                func(repoName string) (model.Repo, error)
	createRepo                   func(repo model.Repo) (model.Repo, error)
	getSubscribedRepos           func() ([]model.Repo, error)
	getNotificationTargetsByRepo func(repoID int) ([]model.NotificationTarget, error)
	updateLastSeenTag            func(repoID int, tag string) error
}

func (m *MockRepository) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
	return m.getSubscriptionsByEmail(email)
}

func (m *MockRepository) CreateSubscription(email string, repoID int, confirmToken, unsubscribeToken string) error {
	return m.createSubscription(email, repoID, confirmToken, unsubscribeToken)
}

func (m *MockRepository) SubscriptionExists(email string, repoName string) (bool, error) {
	return m.subscriptionExists(email, repoName)
}

func (m *MockRepository) ConfirmSubscription(confirmToken string) error {
	return m.confirmSubscription(confirmToken)
}

func (m *MockRepository) DeleteSubscription(unsubscribeToken string) error {
	return m.deleteSubscription(unsubscribeToken)
}

func (m *MockRepository) GetRepoByName(repoName string) (model.Repo, error) {
	return m.getRepoByName(repoName)
}

func (m *MockRepository) CreateRepo(repo model.Repo) (model.Repo, error) {
	return m.createRepo(repo)
}

func (m *MockRepository) GetSubscribedRepos() ([]model.Repo, error) {
	return m.getSubscribedRepos()
}

func (m *MockRepository) GetNotificationTargetsByRepo(repoID int) ([]model.NotificationTarget, error) {
	return m.getNotificationTargetsByRepo(repoID)
}

func (m *MockRepository) UpdateLastSeenTag(repoID int, tag string) error {
	return m.updateLastSeenTag(repoID, tag)
}
