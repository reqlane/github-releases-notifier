package router

import (
	"database/sql"
	"net/http"

	"github.com/reqlane/github-releases-notifier/internal/githubapi"
)

type app struct {
	db           *sql.DB
	githubClient *githubapi.GithubClient
}

func NewApp(db *sql.DB, githubClient *githubapi.GithubClient) *app {
	return &app{
		db:           db,
		githubClient: githubClient,
	}
}

func (a *app) Router() *http.ServeMux {
	mux := http.NewServeMux()

	a.subscriptionRouter(mux)

	return mux
}
