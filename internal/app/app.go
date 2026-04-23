package app

import (
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/api/router"
	"github.com/reqlane/github-releases-notifier/internal/api/service"
	"github.com/reqlane/github-releases-notifier/internal/config"
	"github.com/reqlane/github-releases-notifier/internal/db"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
	"github.com/reqlane/github-releases-notifier/internal/notifier"
	"github.com/reqlane/github-releases-notifier/internal/repository"
	"github.com/reqlane/github-releases-notifier/internal/scanner"
	"github.com/rs/zerolog"
)

func Run() (err error) {
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger()

	err = godotenv.Load(".env")
	if err != nil {
		logger.Warn().Str("error message", err.Error()).Msg("error loading .env file, assuming docker is being used")
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Err(err).Msg("error loading config")
		return
	}

	dbConn, err := db.ConnectDB(cfg.DSN())
	if err != nil {
		logger.Err(err).Msg("error connecting to db")
		return
	}
	defer func() {
		if err = dbConn.Close(); err != nil {
			logger.Err(err).Msg("error closing db connection")
		}
	}()

	err = db.RunMigrations(dbConn)
	if err != nil {
		logger.Err(err).Msg("error running migrations")
		return
	}

	githubApiToken := cfg.GithubToken

	client := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewHTTPGithubClient(&client, logger, githubApiToken)

	notif := notifier.NewSMTPNotifier(notifier.SMTPNotifierConfig{
		Host:          cfg.SMTPHost,
		Port:          cfg.SMTPPort,
		Username:      cfg.SMTPUsername,
		Password:      cfg.SMTPPassword,
		ServerBaseURL: cfg.ServerBaseURL,
	})

	repository := repository.NewMariaDBRepository(dbConn, logger)
	subscriptionService := service.NewSubcriptionService(repository, githubClient, notif)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionService, logger)

	scan := scanner.NewFixedRateScanner(repository, githubClient, notif, logger)
	if githubApiToken == "" {
		scan.SetRequestsPerMin(1)
	}
	go scan.Run()

	rt := router.NewRouter(subscriptionHandler)
	mux := rt.Build()

	server := http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	logger.Info().Msg("Server is running on port: " + cfg.ServerPort)
	err = server.ListenAndServe()
	if err != nil {
		logger.Err(err).Msg("error starting the server")
		return
	}

	return
}
