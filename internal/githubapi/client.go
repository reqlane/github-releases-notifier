package githubapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
)

const githubAPIVersion = "2026-03-10"

// Token - 5000/h (429) | No token - 60/h (403)
//
// Calculating points for the secondary rate limit
// Most REST API GET, HEAD, and OPTIONS requests - 1 point
//
// No more than 100 concurrent requests are allowed.
// No more than 900 points per minute are allowed for REST API endpoints.
type GithubClient struct {
	client       *http.Client
	apiToken     string
	blockedUntil time.Time
	mu           sync.RWMutex
}

func NewGithubClient(client *http.Client, apiToken string) *GithubClient {
	return &GithubClient{client: client, apiToken: apiToken}
}

func (g *GithubClient) blockStatus() (bool, time.Time) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return time.Now().Before(g.blockedUntil), g.blockedUntil
}

func (g *GithubClient) blockUntil(t time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blockedUntil = t
}

func (g *GithubClient) RepoExists(repo string) error {
	if blocked, until := g.blockStatus(); blocked {
		return &apperror.ErrGithubAPIRateLimited{ResetTime: until}
	}

	reqURL := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	req, err := g.githubRequest(http.MethodGet, reqURL)
	if err != nil {
		return err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return apperror.ErrInvalidGithubAPIToken
	case http.StatusForbidden:
		// 403 Forbidden is rate limit related when X-RateLimit-Remaining is 0
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			t := g.handleRateLimit(resp)
			return &apperror.ErrGithubAPIRateLimited{ResetTime: t}
		}
		return apperror.ErrGithubForbidden
	case http.StatusNotFound:
		return &apperror.ErrGithubRepoNotFound{Repo: repo}
	case http.StatusTooManyRequests:
		t := g.handleRateLimit(resp)
		return &apperror.ErrGithubAPIRateLimited{ResetTime: t}
	default:
		return fmt.Errorf("unexpected github api response code: %d", resp.StatusCode)
	}
}

func (g *GithubClient) GetLatestRelease(repo string) (string, error) {
	if blocked, until := g.blockStatus(); blocked {
		return "", &apperror.ErrGithubAPIRateLimited{ResetTime: until}
	}

	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := g.githubRequest(http.MethodGet, reqURL)
	if err != nil {
		return "", err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return "", fmt.Errorf("error decoding github api response: %w", err)
		}
		return release.TagName, nil
	case http.StatusUnauthorized:
		return "", apperror.ErrInvalidGithubAPIToken
	case http.StatusForbidden:
		// 403 Forbidden is rate limit related when X-RateLimit-Remaining is 0
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			t := g.handleRateLimit(resp)
			return "", &apperror.ErrGithubAPIRateLimited{ResetTime: t}
		}
		return "", apperror.ErrGithubForbidden
	case http.StatusNotFound:
		return "", apperror.ErrGithubRepoNoReleases
	case http.StatusTooManyRequests:
		t := g.handleRateLimit(resp)
		return "", &apperror.ErrGithubAPIRateLimited{ResetTime: t}
	default:
		return "", fmt.Errorf("unexpected github api response code: %d", resp.StatusCode)
	}
}

// Github API (c)
// If you receive a rate limit error, you should stop making requests temporarily according to these guidelines:
// If the retry-after response header is present, you should not retry your request until after that many seconds has elapsed.
// If the x-ratelimit-remaining header is 0, you should not make another request until after the time specified by the x-ratelimit-reset header.
// The x-ratelimit-reset header is in UTC epoch seconds.
// Otherwise, wait for at least one minute before retrying.
func (g *GithubClient) handleRateLimit(resp *http.Response) time.Time {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		seconds, err := strconv.ParseInt(retryAfter, 10, 64)
		if err == nil {
			t := time.Now().Add(time.Duration(seconds) * time.Second)
			g.blockUntil(t)
			return t
		}
	}
	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
			unix, err := strconv.ParseInt(reset, 10, 64)
			if err == nil {
				t := time.Unix(unix, 0)
				g.blockUntil(t)
				return t
			}
		}
	}
	t := time.Now().Add(time.Minute)
	g.blockUntil(t)
	return t
}

func (g *GithubClient) githubRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	if g.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiToken)
	}

	return req, nil
}
