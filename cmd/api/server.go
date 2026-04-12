package main

import (
	"fmt"
	"log"
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

func main() {
	// Logger
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger()

	// .env (if not using Docker)
	err := godotenv.Load(".env")
	if err != nil {
		logger.Warn().Str("error message", err.Error()).Msg("error loading .env file, assuming docker is being used")
	}

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Err(err).Msg("error loading config")
		os.Exit(1)
	}

	// DB connection
	dbConn, err := db.ConnectDB(cfg.DSN())
	if err != nil {
		logger.Err(err).Msg("error connecting to db")
		os.Exit(1)
	}
	defer dbConn.Close()

	// Migrations
	err = db.RunMigrations(dbConn)
	if err != nil {
		logger.Err(err).Msg("error running migrations")
		os.Exit(1)
	}

	port := cfg.ServerPort
	githubApiToken := cfg.GithubToken

	// Dependencies
	client := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewHTTPGithubClient(&client, githubApiToken)

	notif := notifier.NewSMTPNotifier(notifier.SMTPNotifierConfig{
		Host:          cfg.SMTPHost,
		Port:          cfg.SMTPPort,
		Username:      cfg.SMTPUsername,
		Password:      cfg.SMTPPassword,
		ServerBaseURL: cfg.ServerBaseURL,
	})

	repository := repository.NewMariaDBRepository(dbConn)
	subscriptionService := service.NewSubcriptionService(repository, githubClient, notif)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionService, logger)

	// Start scanner
	scan := scanner.NewFixedRateScanner(repository, githubClient, notif, logger)
	if githubApiToken == "" {
		scan.SetRequestsPerMin(1)
	}
	go scan.Run()

	app := router.NewApp(subscriptionHandler)
	mux := app.Router()

	// Server
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("Server is running on port:", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln("Error starting the server:", err)
	}
}
