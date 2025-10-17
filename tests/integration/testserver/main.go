package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Get port from command line or use default
	port := "8765"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Handle /ping endpoint
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "pong")
	})

	// Handle root for basic info
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Test MCP Server")
	})

	addr := ":" + port
	log.Printf("Test server starting on port %s (PID: %d)", port, os.Getpid())
	log.Fatal(http.ListenAndServe(addr, nil))
}
