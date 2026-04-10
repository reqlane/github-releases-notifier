package githubapi

import (
	"encoding/json"
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
	req, err := g.githubRequest(http.MethodGet, url)
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
	case http.StatusNotFound:
		return &apperror.ErrGithubRepoNotFound{Repo: repo}
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return &apperror.ErrGithubAPIRateLimited{RetryAfter: retryAfter}
	default:
		return fmt.Errorf("unexpected github api response code: %d", resp.StatusCode)
	}
}

func (g *GithubClient) GetLatestRelease(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := g.githubRequest(http.MethodGet, url)
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
	case http.StatusNotFound:
		return "", nil
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return "", &apperror.ErrGithubAPIRateLimited{RetryAfter: retryAfter}
	default:
		return "", fmt.Errorf("unexpected github api response code: %d", resp.StatusCode)
	}
}

func (g *GithubClient) githubRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if g.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiToken)
	}

	return req, nil
}
