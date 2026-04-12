package mock

type MockNotifier struct {
	sendConfirmation func(recipient, repo, confirmToken, unsubscribeToken string) error
	sendNotification func(recipient, repo, tag, unsubscribeToken string) error
}

func (m *MockNotifier) SendConfirmation(recipient, repo, confirmToken, unsubscribeToken string) error {
	return m.sendConfirmation(recipient, repo, confirmToken, unsubscribeToken)
}

func (m *MockNotifier) SendNotification(recipient, repo, tag, unsubscribeToken string) error {
	return m.sendNotification(recipient, repo, tag, unsubscribeToken)
}
