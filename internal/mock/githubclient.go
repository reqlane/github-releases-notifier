package mock

type MockGithubClient struct {
	RepoExistsFunc       func(repo string) error
	GetLatestReleaseFunc func(repo string) (string, error)
}

func (m *MockGithubClient) RepoExists(repo string) error {
	return m.RepoExistsFunc(repo)
}

func (m *MockGithubClient) GetLatestRelease(repo string) (string, error) {
	return m.GetLatestReleaseFunc(repo)
}

func IdealGithubClient() *MockGithubClient {
	return &MockGithubClient{
		RepoExistsFunc:       func(repo string) error { return nil },
		GetLatestReleaseFunc: func(repo string) (string, error) { return "v1.0.0", nil },
	}
}
