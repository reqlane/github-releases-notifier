package mock

type MockNotifier struct {
	SendConfirmationFunc func(recipient, repo, confirmToken, unsubscribeToken string) error
	SendNotificationFunc func(recipient, repo, tag, unsubscribeToken string) error
}

func (m *MockNotifier) SendConfirmation(recipient, repo, confirmToken, unsubscribeToken string) error {
	return m.SendConfirmationFunc(recipient, repo, confirmToken, unsubscribeToken)
}

func (m *MockNotifier) SendNotification(recipient, repo, tag, unsubscribeToken string) error {
	return m.SendNotificationFunc(recipient, repo, tag, unsubscribeToken)
}

func IdealNotifier() *MockNotifier {
	return &MockNotifier{
		SendConfirmationFunc: func(recipient, repo, confirmToken, unsubscribeToken string) error { return nil },
		SendNotificationFunc: func(recipient, repo, tag, unsubscribeToken string) error { return nil },
	}
}
