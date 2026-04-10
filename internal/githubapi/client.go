package githubapi

import (
	"fmt"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
)

type GithubClient struct {
	client   *http.Client
	apiToken string
}

func NewGithubClient(client *http.Client, apiToken string) *GithubClient {
	return &GithubClient{client: client, apiToken: apiToken}
}

func (g *GithubClient) RepoExists(repo string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if g.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiToken)
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
	case http.StatusNotFound:
		return &apperror.ErrGithubRepoNotFound{Repo: repo}
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return &apperror.ErrGithubAPIRateLimited{RetryAfter: retryAfter}
	default:
		return fmt.Errorf("unexpected github api response code: %d", resp.StatusCode)
	}
}
