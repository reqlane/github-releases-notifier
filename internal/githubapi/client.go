package githubapi

import "net/http"

type GithubClient struct {
	client   *http.Client
	apiToken string
}

func NewGithubClient(client *http.Client, apiToken string) *GithubClient {
	return &GithubClient{client: client, apiToken: apiToken}
}

func (g *GithubClient) RepoExists(repo string) error {
	return nil
}
