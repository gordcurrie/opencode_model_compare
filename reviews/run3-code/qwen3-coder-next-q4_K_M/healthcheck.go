// URL Health Checker - Concurrent Go Program
// Checks multiple URLs simultaneously and reports their status

package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Result holds the outcome of checking a single URL
type Result struct {
	URL         string
	Status      int
	RespTime    time.Duration
	Err         error
}

// checkURL sends a GET request to the given URL and returns the result
func checkURL(url string, timeout time.Duration) Result {
	start := time.Now()
	
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Get(url)
	
	result := Result{
		URL: url,
	}
	
	if err != nil {
		result.Err = err
		result.RespTime = time.Since(start)
		return result
	}
	defer resp.Body.Close()
	
	result.Status = resp.StatusCode
	result.RespTime = time.Since(start)
	
	return result
}

// HealthChecker manages concurrent URL checking
type HealthChecker struct {
	urls     []string
	results  chan Result
	timeout  time.Duration
	wg       sync.WaitGroup
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(urls []string, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		urls:    urls,
		results: make(chan Result, len(urls)),
		timeout: timeout,
	}
}

// checkSingleURL is the worker function that checks one URL
func (hc *HealthChecker) checkSingleURL(url string) {
	defer hc.wg.Done()
	result := checkURL(url, hc.timeout)
	hc.results <- result
}

// CheckAll concurrently checks all URLs and returns results
func (hc *HealthChecker) CheckAll() []Result {
	// Start goroutines for each URL
	for _, url := range hc.urls {
		hc.wg.Add(1)
		go hc.checkSingleURL(url)
	}
	
	// Wait for all goroutines to complete in a separate goroutine
	go func() {
		hc.wg.Wait()
		close(hc.results)
	}()
	
	// Collect all results
	var results []Result
	for result := range hc.results {
		results = append(results, result)
	}
	
	return results
}

// formatStatus converts HTTP status code to human-readable string
func formatStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "OK"
	case code >= 300 && code < 400:
		return "Redirect"
	case code >= 400 && code < 500:
		return "Client Error"
	case code >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

// main demonstrates the health checker with 5 sample URLs
func main() {
	// Sample URLs to check
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://httpstat.us/200",
		"https://httpstat.us/404",
		"https://nonexistent-domain-12345.com",
	}
	
	// 5 second timeout per request
	timeout := 5 * time.Second
	
	fmt.Println(" URL Health Checker")
	fmt.Println("==================")
	fmt.Printf(" Checking %d URLs with %v timeout...\n\n", len(urls), timeout)
	
	// Create and run health checker
	checker := NewHealthChecker(urls, timeout)
	results := checker.CheckAll()
	
	// Print results
	fmt.Println("Results:")
	fmt.Println("--------")
	
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("%-50s FAILED (%v)\n", r.URL, r.Err)
		} else {
			fmt.Printf("%-50s %d %s (%v)\n", r.URL, r.Status, formatStatus(r.Status), r.RespTime)
		}
	}
	
	// Summary
	fmt.Println("\n--------")
	fmt.Println("Summary:")
	
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Err != nil || (r.Status >= 400 && r.Status < 500) {
			failed++
		} else {
			passed++
		}
	}
	
	fmt.Printf("Successful: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
}
