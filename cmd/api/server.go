package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/api/repository"
	"github.com/reqlane/github-releases-notifier/internal/api/router"
	"github.com/reqlane/github-releases-notifier/internal/api/service"
	"github.com/reqlane/github-releases-notifier/internal/config"
	"github.com/reqlane/github-releases-notifier/internal/db"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("Error loading config:", err)
	}

	dbConn, err := db.ConnectDB(cfg.DSN())
	if err != nil {
		log.Fatalln("Error connecting to db:", err)
	}
	defer dbConn.Close()

	err = db.RunMigrations(dbConn)
	if err != nil {
		log.Fatalln("Error running migratrions:", err)
	}

	port := cfg.ServerPort
	githubApiToken := cfg.GithubToken

	client := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewGithubClient(&client, githubApiToken)

	subscriptionRepository := repository.NewSubcriptionRepository(dbConn)
	subscriptionService := service.NewSubcriptionService(subscriptionRepository, githubClient)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionService)

	app := router.NewApp(subscriptionHandler)
	mux := app.Router()

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
