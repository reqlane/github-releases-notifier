package mocknotifier

import (
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/stretchr/testify/mock"
)

type Notifier struct {
	mock.Mock
}

func (m *Notifier) SendConfirmation(recipient, repo string, tokens model.SubscriptionTokens) error {
	args := m.Called(recipient, repo, tokens)
	return args.Error(0)
}

func (m *Notifier) SendNotification(recipient, repo, tag, unsubscribeToken string) error {
	args := m.Called(recipient, repo, tag, unsubscribeToken)
	return args.Error(0)
}
