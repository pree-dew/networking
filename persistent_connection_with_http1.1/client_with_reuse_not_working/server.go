package main

import (
	"fmt"
	"net/http"
)

func case1Handler(w http.ResponseWriter, r *http.Request) {
	// Case 1: Normal response, client won't read body
	fmt.Println("\nCase 1: Sending response that client won't read")
	w.Write([]byte("This response body won't be read by client\n"))
}

func case2Handler(w http.ResponseWriter, r *http.Request) {
	// Case 2: Client requests connection close
	fmt.Printf("\nCase 2: Client requested connection close: %v\n",
		r.Header.Get("Connection"))
	w.Write([]byte("Responding to client that requested connection close\n"))
}

func case3Handler(w http.ResponseWriter, r *http.Request) {
	// Case 3: Server forces connection close
	fmt.Println("\nCase 3: Server forcing connection close")
	w.Header().Set("Connection", "close")
	w.Write([]byte("Server is closing this connection\n"))
}

func case4Handler(w http.ResponseWriter, r *http.Request) {
	// Case 5: Server closes connection mid-response
	fmt.Println("\nCase 5: Server will close connection mid-response")
	w.Header().Set("Content-Length", "1000") // Lie about content length
	w.Write([]byte("Starting response..."))
	// Close connection prematurely
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	hj, _ := w.(http.Hijacker)
	conn, _, _ := hj.Hijack()
	conn.Close()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/case1", case1Handler)
	mux.HandleFunc("/case2", case2Handler)
	mux.HandleFunc("/case3", case3Handler)
	mux.HandleFunc("/case4", case4Handler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server listening on :8080")
	server.ListenAndServe()
}
