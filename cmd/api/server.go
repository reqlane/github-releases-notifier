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
	// .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("Error loading config:", err)
	}

	// DB connection
	dbConn, err := db.ConnectDB(cfg.DSN())
	if err != nil {
		log.Fatalln("Error connecting to db:", err)
	}
	defer dbConn.Close()

	// Migrations
	err = db.RunMigrations(dbConn)
	if err != nil {
		log.Fatalln("Error running migratrions:", err)
	}

	port := cfg.ServerPort
	githubApiToken := cfg.GithubToken

	// Dependencies
	client := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewGithubClient(&client, githubApiToken)

	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger()

	notif := notifier.New(notifier.Config{
		Host:          cfg.SMTPHost,
		Port:          cfg.SMTPPort,
		Username:      cfg.SMTPUsername,
		Password:      cfg.SMTPPassword,
		ServerBaseURL: cfg.ServerBaseURL,
	})

	repository := repository.NewRepository(dbConn)
	subscriptionService := service.NewSubcriptionService(repository, githubClient, notif)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionService, logger)

	// Start scanner
	scan := scanner.New(repository, githubClient, notif, logger)
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
