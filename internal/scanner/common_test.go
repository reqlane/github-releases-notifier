package scanner

import (
	"io"
	"time"

	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	mockrepository "github.com/reqlane/github-releases-notifier/internal/mock/repository"
	"github.com/rs/zerolog"
)

// --- Scanner ---
func scannerMocks() (*mockrepository.SubscriptionRepo, *mockgithubapi.GithubClient, *mocknotifier.Notifier) {
	return new(mockrepository.SubscriptionRepo),
		new(mockgithubapi.GithubClient),
		new(mocknotifier.Notifier)
}

func newScanner(r *mockrepository.SubscriptionRepo, g *mockgithubapi.GithubClient, n *mocknotifier.Notifier) *FixedRateScanner {
	return &FixedRateScanner{
		repo:           r,
		githubClient:   g,
		notif:          n,
		logger:         zerolog.New(io.Discard), // no logs in tests
		requestsPerMin: defaultRequestsPerMin,
		sleepOnEmpty:   defaultSleepOnEmpty,
		pauseCh:        make(chan time.Time, 3),
		sleepFn:        func(time.Duration) {}, // no sleep in tests
	}
}
