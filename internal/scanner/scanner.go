package scanner

import (
	"errors"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/contract"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"github.com/rs/zerolog"
)

const (
	defaultRequestsPerMin = 60
	defaultSleepOnEmpty   = 1 * time.Minute
)

// Github API (c)
// Calculating points for the secondary rate limit
// Most REST API GET, HEAD, and OPTIONS requests - 1 point
//
// No more than 100 concurrent requests are allowed.
// No more than 900 points per minute are allowed for REST API endpoints.
type FixedRateScanner struct {
	repo           contract.SubscriptionRepo
	githubClient   contract.GithubClient
	notif          contract.Notifier
	logger         zerolog.Logger
	requestsPerMin int
	sleepOnEmpty   time.Duration
	pauseCh        chan time.Time
	sleepFn        func(time.Duration)
}

func NewFixedRateScanner(r contract.SubscriptionRepo, g contract.GithubClient, n contract.Notifier, l zerolog.Logger) *FixedRateScanner {
	return &FixedRateScanner{
		repo:           r,
		githubClient:   g,
		notif:          n,
		logger:         l,
		requestsPerMin: defaultRequestsPerMin,
		sleepOnEmpty:   defaultSleepOnEmpty,
		// size 3 to prevent blocked "checkRepo" goroutines
		pauseCh: make(chan time.Time, 3),
		sleepFn: time.Sleep,
	}
}

func (s *FixedRateScanner) SetRequestsPerMin(requestsPerMin int) {
	s.requestsPerMin = requestsPerMin
}

func (s *FixedRateScanner) Run() {
	for {
		s.scan()
	}
}

func (s *FixedRateScanner) scan() {
	// Get subscribed repos
	repos, err := s.repo.GetSubscribedRepos()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting subscribed repos from database")
		s.sleepFn(s.sleepOnEmpty)
		return
	}
	if len(repos) == 0 {
		s.logger.Info().Msg("no subscribed repos to check, scanner sleeps")
		s.sleepFn(s.sleepOnEmpty)
		return
	}

	delay := time.Minute / time.Duration(s.requestsPerMin)

	for _, repo := range repos {
		// Check for rate limit pause signal
		select {
		case resetTime := <-s.pauseCh:
			drained := false
			// Free goroutines if concurrent ones present
			for !drained {
				select {
				case <-s.pauseCh:
				default:
					drained = true
				}
			}
			s.logger.Warn().Time("reset time", resetTime).Msg("rate limit pause signal received, pausing scanner")
			s.sleepFn(time.Until(resetTime))
		default:
		}

		// Start checking repos in separate goroutines to maintain consistent rate
		go s.checkRepo(repo)
		s.sleepFn(delay)
	}
}

func (s *FixedRateScanner) checkRepo(repo model.Repo) {
	tag, err := s.githubClient.GetLatestRelease(repo.Repo)
	if err != nil && !errors.Is(err, apperror.ErrGithubRepoNoReleases) {
		if e, ok := errors.AsType[*apperror.ErrGithubAPIRateLimited](err); ok {
			s.logger.Warn().Time("reset time", e.ResetTime).Msg("rate limited, sent pause signal to scanner")
			s.pauseCh <- e.ResetTime
			return
		}
		s.logger.Err(err).Str("repo", repo.Repo).Msg("error getting latest release of a repo")
		return
	}

	if tag == nil || *tag == repo.LastSeenTag {
		s.logger.Info().Str("repo", repo.Repo).Msg("no new releases for repo")
		return
	}

	err = s.repo.UpdateLastSeenTag(repo.ID, tag)
	if err != nil {
		s.logger.Err(err).Str("repo", repo.Repo).Msg("error updating last seen tag of a repo")
		return
	}

	targets, err := s.repo.GetNotificationTargetsByRepo(repo.ID)
	if err != nil {
		s.logger.Err(err).Str("repo", repo.Repo).Msg("error getting notification targets of a repo")
		return
	}

	for _, target := range targets {
		err = s.notif.SendNotification(target.Email, repo.Repo, *tag, target.UnsubscribeToken)
		if err != nil {
			s.logger.Err(err).Str("email", target.Email).Msg("error sending notification to email")
		}
	}
}
