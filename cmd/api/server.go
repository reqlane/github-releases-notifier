package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/reqlane/github-releases-notifier/internal/api/router"
	"github.com/reqlane/github-releases-notifier/internal/db"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	db, err := db.ConnectDB()
	if err != nil {
		log.Fatalln("Error connecting to db:", err)
	}
	defer db.Close()

	port := os.Getenv("SERVER_PORT")

	app := router.NewApp(db)
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
