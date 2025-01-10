package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

func main() {
	// Create a new request
	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	var start, connect, dns time.Time
	var dnsDone bool

	// Create a trace to measure detailed timing
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
		},
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			dnsDone = true
		},
		ConnectStart: func(network, addr string) {
			if dnsDone {
				fmt.Printf("DNS lookup time: %v\n", time.Since(dns))
			}
			connect = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				fmt.Printf("Connect error: %v\n", err)
				return
			}
			fmt.Printf("TCP handshake time: %v\n", time.Since(connect))
		},
	}

	// Add the trace to the request context
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// Record start time
	start = time.Now()

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Calculate time to first byte
	ttfb := time.Since(start)
	fmt.Printf("Time to first byte: %v\n", ttfb)

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	// Calculate total time
	total := time.Since(start)

	fmt.Printf("Server response: %s", body)
	fmt.Printf("Total time: %v\n", total)
	fmt.Printf("Server processing time (approximate): %v\n", total-ttfb)
}
