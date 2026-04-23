package contract

type Notifier interface {
	SendConfirmation(recipient, repo, confirmToken, unsubscribeToken string) error
	SendNotification(recipient, repo, tag, unsubscribeToken string) error
}
