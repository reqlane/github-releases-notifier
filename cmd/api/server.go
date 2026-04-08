package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := "3000"

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "placeholder") })

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("Error starting the server:", err)
	}
}
