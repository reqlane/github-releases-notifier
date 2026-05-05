package usecase

import (
	"strings"

	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	mockrepository "github.com/reqlane/github-releases-notifier/internal/mock/repository"
)

var (
	invalidEmails = []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"missing @", "userexample.com"},
		{"missing domain", "user@"},
		{"missing local part", "@example.com"},
		{"spaces", "user @example.com"},
	}
	invalidRepos = []struct {
		name string
		repo string
	}{
		{"missing slash", "ownerrepo"},
		{"owner too long", strings.Repeat("a", 40) + "/repo"},
		{"repo name too long", "owner/" + strings.Repeat("a", 101)},
		{"starts with hyphen", "-owner/repo"},
		{"double slash", "owner//repo"},
	}
	validTokens = []string{
		"32d49477a4752b36bcaeed3a25249c4333eb04333971f5ddd5fa568337d038f1",
		"aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
		"aabbcca4752b3a1f5ddd5cddaabbccdb1f5fa568335fa56833eeeeeeeebbccdd",
	}
	invalidTokens = []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"too short", "aabbccdd"},
		{"too long", "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd00"},
		{"non-hex characters", "zzbbaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabb"},
		{"correct length but spaces", "aabbccddaabbccdd aabbccddaabbccddaabbccddaabbccddaabbccddaabbcc"},
	}
)

// --- SubscriptionUseCase ---
func subscriptionUseCaseMocks() (*mockrepository.SubscriptionRepo, *mockgithubapi.GithubClient, *mocknotifier.Notifier) {
	return new(mockrepository.SubscriptionRepo),
		new(mockgithubapi.GithubClient),
		new(mocknotifier.Notifier)
}

func newSubscriptionUseCase(r *mockrepository.SubscriptionRepo, g *mockgithubapi.GithubClient, n *mocknotifier.Notifier) SubscriptionUseCase {
	return NewSubscriptionUseCase(r, g, n)
}
