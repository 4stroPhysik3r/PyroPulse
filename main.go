package main

import (
	"log"
	"net/http"

	"bomberman-dom/server"
)

func main() {

	hub := server.NewHub()
	go hub.Run()

	// Register handlers
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(w, r, hub) // Pass the game instance to HandleWebSocket
	})

	// Create a file server to serve static files
	fs := http.FileServer(http.Dir("./client"))
	http.Handle("/", fs)

	// Start the server
	log.Println("Starting server at http://localhost:5050")
	err := http.ListenAndServe(":5050", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
