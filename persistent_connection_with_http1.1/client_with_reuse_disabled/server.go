package main

import (
	"fmt"
	"net/http"
	"time"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("Response at %v\n", time.Now().Format(time.RFC3339))
	w.Write([]byte(response))
}

func main() {
	http.HandleFunc("/", handleRequest)

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
