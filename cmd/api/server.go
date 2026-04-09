package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("main:", err)
		return
	}

	port := os.Getenv("SERVER_PORT")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "placeholder") })

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
