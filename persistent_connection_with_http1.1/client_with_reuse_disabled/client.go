package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

func makeRequest(client *http.Client, reqNum int) {
	fmt.Printf("\n--- Request %d ---\n", reqNum)

	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	// Force connection close after request
	req.Close = true
	req.Header.Set("Connection", "close")

	var start, connect, dns time.Time
	var dnsDone bool

	// Create a trace to measure detailed timing
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
		},
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			dnsDone = true
			if ddi.Err != nil {
				fmt.Printf("DNS error: %v\n", ddi.Err)
			}
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
		GotConn: func(info httptrace.GotConnInfo) {
			fmt.Printf("Connection was reused: %v\n", info.Reused)
			if info.WasIdle {
				fmt.Printf("Connection was idle for %v\n", info.IdleTime)
			}
		},
	}

	// Add the trace to the request context
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// Record start time
	start = time.Now()

	// Make the request
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
	fmt.Printf("Connection header: %s\n", resp.Header.Get("Connection"))
	fmt.Printf("Total time: %v\n", total)
	fmt.Printf("Server processing time (approximate): %v\n", total-ttfb+time.Since(connect))
}

func main() {
	// Create a client with transport configured to disable connection reuse
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true, // Disable keep-alive, force new connections
		},
	}

	// Make multiple requests to demonstrate new connections each time
	for i := 1; i <= 3; i++ {
		makeRequest(client, i)
		time.Sleep(time.Second) // Wait a bit between requests
	}
}
