package router

import (
	"database/sql"
	"net/http"
)

type app struct {
	db *sql.DB
}

func NewApp(db *sql.DB) *app {
	return &app{db: db}
}

func (a *app) Router() *http.ServeMux {
	mux := http.NewServeMux()

	a.subscriptionRouter(mux)

	return mux
}
