package mock

type MockGithubClient struct {
	repoExists       func(repo string) error
	getLatestRelease func(repo string) (string, error)
}

func (m *MockGithubClient) RepoExists(repo string) error {
	return m.repoExists(repo)
}

func (m *MockGithubClient) GetLatestRelease(repo string) (string, error) {
	return m.getLatestRelease(repo)
}
