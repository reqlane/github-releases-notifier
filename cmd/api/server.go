package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/reqlane/github-releases-notifier/internal/api/router"
	"github.com/reqlane/github-releases-notifier/internal/db"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	dbConn, err := db.ConnectDB()
	if err != nil {
		log.Fatalln("Error connecting to db:", err)
	}
	defer dbConn.Close()

	err = db.RunMigrations(dbConn)
	if err != nil {
		log.Fatalln("Error running migratrions:", err)
	}

	port := os.Getenv("SERVER_PORT")
	githubApiToken := os.Getenv("GITHUB_API_TOKEN")

	client := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewGithubClient(&client, githubApiToken)
	app := router.NewApp(dbConn, githubClient)
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
