package contract

import "github.com/reqlane/github-releases-notifier/internal/model"

type SubscriptionRepo interface {
	GetSubscriptionsByEmail(email string) ([]model.Subscription, error)
	CreateSubscription(email string, repoID int, confirmToken, unsubscribeToken string) error
	SubscriptionExists(email string, repoName string) (bool, error)
	ConfirmSubscription(confirmToken string) error
	DeleteSubscription(unsubscribeToken string) error
	GetRepoByName(repoName string) (model.Repo, error)
	CreateRepo(repo model.Repo) (model.Repo, error)
	GetSubscribedRepos() ([]model.Repo, error)
	GetNotificationTargetsByRepo(repoID int) ([]model.NotificationTarget, error)
	UpdateLastSeenTag(repoID int, tag string) error
}
