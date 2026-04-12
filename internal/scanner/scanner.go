package scanner

import (
	"time"

	"github.com/reqlane/github-releases-notifier/internal/githubapi"
	"github.com/reqlane/github-releases-notifier/internal/notifier"
	"github.com/reqlane/github-releases-notifier/internal/repository"
)

const (
	defaultRequestsPerMinute = 60
	defaultSleepOnEmpty      = 1 * time.Minute
)

// Calculating points for the secondary rate limit
// Most REST API GET, HEAD, and OPTIONS requests - 1 point
//
// No more than 100 concurrent requests are allowed.
// No more than 900 points per minute are allowed for REST API endpoints.
type Scanner struct {
	repo           *repository.Repository
	githubClient   *githubapi.GithubClient
	notif          *notifier.Notifier
	requestsPerMin int
	sleepOnEmpty   time.Duration
}

func New(r *repository.Repository, g *githubapi.GithubClient, n *notifier.Notifier) *Scanner {
	return &Scanner{
		repo:           r,
		githubClient:   g,
		notif:          n,
		requestsPerMin: defaultRequestsPerMinute,
		sleepOnEmpty:   defaultSleepOnEmpty,
	}
}

func (s *Scanner) SetRequestsPerMin(requestsPerMin int) *Scanner {
	s.requestsPerMin = requestsPerMin
	return s
}

func (s *Scanner) SetSleepOnEmpty(sleepOnEmpty time.Duration) *Scanner {
	s.sleepOnEmpty = sleepOnEmpty
	return s
}

func (s *Scanner) Run() {
	for {
		s.scan()
	}
}

func (s *Scanner) scan() {

}
