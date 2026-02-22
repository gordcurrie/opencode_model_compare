package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthResult represents the result of checking a URL
type HealthResult struct {
	URL         string
	Status      int
	ResponseTime time.Duration
	Error       error
}

// checkURL checks the health of a single URL with a 5-second timeout
func checkURL(url string, resultChan chan<- HealthResult, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create a HTTP request with a 5-second timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(url)
	responseTime := time.Since(start)

	// Prepare the result
	result := HealthResult{
		URL:         url,
		ResponseTime: responseTime,
	}

	if err != nil {
		result.Error = err
		result.Status = 0
	} else {
		result.Status = resp.StatusCode
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	resultChan <- result
}

// checkURLs concurrently checks the health of all URLs in the provided slice
func checkURLs(urls []string) []HealthResult {
	// Channel to receive results
	resultChan := make(chan HealthResult, len(urls))
	
	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Start a goroutine for each URL
	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, resultChan, &wg)
	}
	
	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	var results []HealthResult
	for result := range resultChan {
		results = append(results, result)
	}
	
	return results
}

func main() {
	// Sample URLs to check
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/500",
		"https://httpbin.org/delay/3",
	}
	
	fmt.Println("Starting concurrent URL health check...")
	
	// Check all URLs
	results := checkURLs(urls)
	
	// Display results
	fmt.Println("\nResults:")
	fmt.Println("========")
	
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("URL: %s\nError: %v\nResponse Time: %v\n\n", result.URL, result.Error, result.ResponseTime)
		} else {
			fmt.Printf("URL: %s\nStatus: %d\nResponse Time: %v\n\n", result.URL, result.Status, result.ResponseTime)
		}
	}
}
