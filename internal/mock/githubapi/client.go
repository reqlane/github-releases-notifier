package mockgithubapi

import "github.com/stretchr/testify/mock"

type GithubClient struct {
	mock.Mock
}

func (m *GithubClient) RepoExists(repo string) error {
	args := m.Called(repo)
	return args.Error(0)
}

func (m *GithubClient) GetLatestRelease(repo string) (string, error) {
	args := m.Called(repo)
	return args.String(0), args.Error(1)
}
