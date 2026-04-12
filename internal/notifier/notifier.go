package notifier

import (
	"fmt"

	gomail "gopkg.in/mail.v2"
)

type Notifier interface {
	SendConfirmation(recipient, repo, confirmToken, unsubscribeToken string) error
	SendNotification(recipient, repo, tag, unsubscribeToken string) error
}

type SMTPNotifier struct {
	dialer        *gomail.Dialer
	serverBaseURL string
}

type SMTPNotifierConfig struct {
	Host          string
	Port          int
	Username      string
	Password      string
	ServerBaseURL string
}

func NewSMTPNotifier(c SMTPNotifierConfig) Notifier {
	return &SMTPNotifier{dialer: gomail.NewDialer(c.Host, c.Port, c.Username, c.Password), serverBaseURL: c.ServerBaseURL}
}

func (n *SMTPNotifier) SendConfirmation(recipient, repo, confirmToken, unsubscribeToken string) error {
	subject := fmt.Sprintf("Confirm your subscription to %s", repo)
	repoURL := fmt.Sprintf("https://github.com/%s", repo)
	confirmURL := fmt.Sprintf("%s/api/confirm/%s", n.serverBaseURL, confirmToken)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", n.serverBaseURL, unsubscribeToken)
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

func (n *SMTPNotifier) SendNotification(recipient, repo, tag, unsubscribeToken string) error {
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

func (n *SMTPNotifier) sendEmail(recipient, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.dialer.Username)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return n.dialer.DialAndSend(m)
}
