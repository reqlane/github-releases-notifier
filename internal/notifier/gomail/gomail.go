package gomail

import (
	"fmt"

	"github.com/reqlane/github-releases-notifier/internal/contract"
	"github.com/reqlane/github-releases-notifier/internal/model"
	gomailv2 "gopkg.in/mail.v2"
)

type gomailNotifier struct {
	dialer        *gomailv2.Dialer
	serverBaseURL string
}

type GomailNotifierConfig struct {
	Host          string
	Port          int
	Username      string
	Password      string
	ServerBaseURL string
}

func NewNotifier(c GomailNotifierConfig) contract.Notifier {
	return &gomailNotifier{dialer: gomailv2.NewDialer(c.Host, c.Port, c.Username, c.Password), serverBaseURL: c.ServerBaseURL}
}

func (n *gomailNotifier) SendConfirmation(recipient, repo string, tokens model.SubscriptionTokens) error {
	subject := fmt.Sprintf("Confirm your subscription to %s", repo)
	repoURL := fmt.Sprintf("https://github.com/%s", repo)
	confirmURL := fmt.Sprintf("%s/api/confirm/%s", n.serverBaseURL, tokens.ConfirmToken)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", n.serverBaseURL, tokens.UnsubscribeToken)
	body := fmt.Sprintf(
		"You requested to subscribe to new releases notifications for the Github repository:\n"+
			"%s\n\n"+
			"To confirm your subscription, click the link below:\n"+
			"%s\n\n"+
			"If you did not request this, simply ignore this email.\n\n"+
			"To unsubscribe at any time, click here:\n"+
			"%s",
		repoURL, confirmURL, unsubscribeURL,
	)
	return n.sendEmail(recipient, subject, body)
}

func (n *gomailNotifier) SendNotification(recipient, repo, tag, unsubscribeToken string) error {
	subject := fmt.Sprintf("New release: %s (%s)", repo, tag)
	repoURL := fmt.Sprintf("https://github.com/%s", repo)
	releaseURL := fmt.Sprintf("%s/releases/tag/%s", repoURL, tag)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", n.serverBaseURL, unsubscribeToken)
	body := fmt.Sprintf(
		"A new release %s is available for %s.\n\n"+
			"Repository: %s\n"+
			"Release: %s\n\n"+
			"To cancel your subscription from these notifications, click here:\n"+
			"%s",
		tag, repo, repoURL, releaseURL, unsubscribeURL,
	)
	return n.sendEmail(recipient, subject, body)
}

func (n *gomailNotifier) sendEmail(recipient, subject, body string) error {
	m := gomailv2.NewMessage()
	m.SetHeader("From", n.dialer.Username)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return n.dialer.DialAndSend(m)
}
