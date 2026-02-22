package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Result holds the health check result for a single URL
type Result struct {
	URL          string
	Status       bool   // true if reachable (HTTP 200), false if failed
	ResponseTime time.Duration
	Error        error
}

// checkURL performs a concurrent health check for a single URL
func checkURL(url string, timeout time.Duration, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()
	
	start := time.Now()
	
	// Create a client with timeout
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		results <- Result{
			URL:          url,
			Status:       false,
			ResponseTime: time.Since(start),
			Error:        err,
		}
		return
	}
	defer resp.Body.Close()
	
	responseTime := time.Since(start)
	isReachable := resp.StatusCode == http.StatusOK
	
	results <- Result{
		URL:          url,
		Status:       isReachable,
		ResponseTime: responseTime,
		Error:        nil,
	}
}

// main demonstrates concurrent URL health checking with sample URLs
func main() {
	// Sample URLs to check
	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://google.com",
		"https://github.com",
		"https://badurl12345.com",
	}
	
	// Configuration
	timeout := 5 * time.Second
	
	// Channels for results
	results := make(chan Result, len(urls))
	
	// WaitGroup to track all goroutines
	var wg sync.WaitGroup
	
	// Spawn goroutines for each URL
	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, timeout, &wg, results)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect and display results
	fmt.Printf("%-50s %10s %15s %10s\n", "URL", "Status", "Response Time", "Error")
	fmt.Println(strings.Repeat("-", 85))
	
	for result := range results {
		if result.Error != nil {
			fmt.Printf("%-50s %10s %15v %10s\n", 
			 truncateURL(result.URL, 50), "FAIL", result.ResponseTime, truncateError(result.Error.Error(), 8))
		} else if result.Status {
			fmt.Printf("%-50s %10s %15v %10s\n", 
			 truncateURL(result.URL, 50), "OK", result.ResponseTime, "(http.StatusOK)")
		} else {
			fmt.Printf("%-50s %10s %15v %10s\n", 
			 truncateURL(result.URL, 50), "FAIL", result.ResponseTime, "(non-200 status)")
		}
	}
}

// truncateURL limits URL length for display
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

// truncateError limits error message length for display
func truncateError(err string, maxLen int) string {
	if len(err) <= maxLen {
		return err
	}
	return err[:maxLen] + "..."
}
