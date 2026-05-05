package contract

type GithubClient interface {
	RepoExists(repo string) error
	GetLatestRelease(repo string) (string, error)
}
