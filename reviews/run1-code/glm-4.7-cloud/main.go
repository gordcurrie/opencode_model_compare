package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// URLHealthResult holds the result of checking a single URL
type URLHealthResult struct {
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

// URLHealthChecker handles concurrent URL checking
type URLHealthChecker struct {
	client *http.Client
}

// NewURLHealthChecker creates a new checker with a 5-second timeout
func NewURLHealthChecker() *URLHealthChecker {
	return &URLHealthChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// checkURL performs a health check on a single URL and sends the result to the channel
func (c *URLHealthChecker) checkURL(url string, results chan<- URLHealthResult, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	resp, err := c.client.Get(url)
	responseTime := time.Since(startTime)

	result := URLHealthResult{
		URL:          url,
		ResponseTime: responseTime,
	}

	if err != nil {
		result.Error = err
		result.StatusCode = 0
	} else {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	results <- result
}

// CheckURLs checks multiple URLs concurrently and returns the results
func (c *URLHealthChecker) CheckURLs(urls []string) []URLHealthResult {
	var wg sync.WaitGroup
	results := make(chan URLHealthResult, len(urls))

	// Start a goroutine for each URL
	for _, url := range urls {
		wg.Add(1)
		go c.checkURL(url, results, &wg)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var allResults []URLHealthResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

// printResults displays the health check results in a formatted way
func printResults(results []URLHealthResult) {
	fmt.Println("URL Health Check Results:")
	fmt.Println("========================")
	fmt.Println()

	reachable := 0
	failed := 0

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s\n", result.URL)
			fmt.Printf("   Error: %v\n", result.Error)
			fmt.Printf("   Response Time: %v\n", result.ResponseTime)
			failed++
		} else if result.StatusCode == http.StatusOK {
			fmt.Printf("✓ %s\n", result.URL)
			fmt.Printf("   Status: %d OK\n", result.StatusCode)
			fmt.Printf("   Response Time: %v\n", result.ResponseTime)
			reachable++
		} else {
			fmt.Printf("⚠ %s\n", result.URL)
			fmt.Printf("   Status: %d (expected 200 OK)\n", result.StatusCode)
			fmt.Printf("   Response Time: %v\n", result.ResponseTime)
			failed++
		}
		fmt.Println()
	}

	fmt.Println("========================")
	fmt.Printf("Total URLs: %d\n", len(results))
	fmt.Printf("Reachable: %d\n", reachable)
	fmt.Printf("Failed: %d\n", failed)
}

func main() {
	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/500",
		"https://www.google.com",
		"https://httpbin.org/delay/3",
	}

	checker := NewURLHealthChecker()
	fmt.Printf("Checking %d URLs concurrently (5-second timeout each)...\n\n", len(urls))

	startTime := time.Now()
	results := checker.CheckURLs(urls)
	totalTime := time.Since(startTime)

	fmt.Printf("Total time: %v (concurrent execution)\n\n", totalTime)
	printResults(results)
}
