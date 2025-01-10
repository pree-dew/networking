package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

type ConnTracker struct {
	connCount int
}

func (ct *ConnTracker) trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			fmt.Printf("Getting connection for %s\n", hostPort)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			fmt.Printf("Got connection: reused=%v, wasIdle=%v, idleTime=%v\n",
				info.Reused, info.WasIdle, info.IdleTime)
		},
		PutIdleConn: func(err error) {
			if err != nil {
				fmt.Printf("Connection not returned to pool: %v\n", err)
			} else {
				fmt.Printf("Connection returned to pool successfully\n")
			}
		},
		ConnectStart: func(network, addr string) {
			fmt.Printf("Starting new connection to %s\n", addr)
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				fmt.Printf("Connection error: %v\n", err)
			} else {
				fmt.Printf("Connection established to %s\n", addr)
			}
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err != nil {
				fmt.Printf("Error writing request: %v\n", info.Err)
			} else {
				fmt.Printf("Wrote request successfully\n")
			}
		},
	}
}

func makeRequest(client *http.Client, url string, config requestConfig) {
	tracker := &ConnTracker{}

	req, err := http.NewRequest(config.method, url, config.body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	if config.closeHeader {
		req.Header.Set("Connection", "close")
	}

	// Add the connection trace
	ctx := httptrace.WithClientTrace(context.Background(), tracker.trace())
	req = req.WithContext(ctx)

	fmt.Printf("\n=== Making request to %s ===\n", url)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// if config doesn't have skip body
	if !config.skipBody {
		// discard body to complete the request and reponse cycle
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			fmt.Println("Not able to read response body")
			return
		}
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Response Connection header: %s\n", resp.Header.Get("Connection"))

	fmt.Println("=== Request complete ===\n")
}

type requestConfig struct {
	method      string
	body        io.Reader
	closeHeader bool
	skipBody    bool
}

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
		},
	}

	baseURL := "http://localhost:8080"

	// Case 1: Don't read response body
	fmt.Println("\n--- Case 1: Not reading response body ---")
	makeRequest(client, baseURL+"/case1", requestConfig{
		method:   "GET",
		skipBody: true,
	})
	// Make another request to show connection not reused
	makeRequest(client, baseURL+"/case1", requestConfig{
		method: "GET",
	})

	// Case 2: Client requests connection close
	fmt.Println("\n--- Case 2: Client requesting connection close ---")
	makeRequest(client, baseURL+"/case2", requestConfig{
		method:      "GET",
		closeHeader: true,
	})
	// Make another request to show new connection
	makeRequest(client, baseURL+"/case2", requestConfig{
		method: "GET",
	})

	// Case 3: Server forces connection close
	fmt.Println("\n--- Case 3: Server forcing connection close ---")
	makeRequest(client, baseURL+"/case3", requestConfig{
		method: "GET",
	})
	// Make another request to show new connection
	makeRequest(client, baseURL+"/case3", requestConfig{
		method: "GET",
	})

	// Case 4: Server closes connection mid-response
	fmt.Println("\n--- Case 4: Server closing connection mid-response ---")
	makeRequest(client, baseURL+"/case4", requestConfig{
		method: "GET",
	})
	// Make another request to show new connection
	makeRequest(client, baseURL+"/case4", requestConfig{
		method: "GET",
	})
}
