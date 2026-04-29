package contract

import "github.com/reqlane/github-releases-notifier/internal/model"

type Notifier interface {
	SendConfirmation(recipient, repo string, tokens model.SubscriptionTokens) error
	SendNotification(recipient, repo, tag, unsubscribeToken string) error
}
