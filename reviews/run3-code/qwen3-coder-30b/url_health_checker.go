package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthResult represents the result of a single URL health check
type HealthResult struct {
	URL          string
	Status       int
	ResponseTime time.Duration
	Error        error
}

// CheckURLHealth concurrently checks the health of multiple URLs
// It uses goroutines and channels to process URLs concurrently
func CheckURLHealth(urls []string) []HealthResult {
	// Create a channel to receive results
	results := make(chan HealthResult, len(urls))
	
	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	
	// Start a goroutine for each URL
	for _, url := range urls {
		wg.Add(1) // Increment the WaitGroup counter
		
		// Launch goroutine to check this URL
		go func(url string) {
			defer wg.Done() // Decrement the WaitGroup counter when done
			
			// Create HTTP request with 5-second timeout
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			
			start := time.Now()
			
			// Make HTTP GET request
			resp, err := client.Get(url)
			duration := time.Since(start)
			
			// Prepare result
			result := HealthResult{
				URL:          url,
				ResponseTime: duration,
			}
			
			if err != nil {
				result.Error = err
			} else {
				result.Status = resp.StatusCode
				resp.Body.Close() // Close the response body to avoid resource leak
			}
			
			// Send result through channel
			results <- result
		}(url) // Pass url as argument to avoid closure issues
	}
	
	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect all results
	var allResults []HealthResult
	for result := range results {
		allResults = append(allResults, result)
	}
	
	return allResults
}

// demonstrateHealthChecker shows how to use the URL health checker
func main() {
	// Sample URLs to check
	sampleURLs := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/500",
		"https://httpbin.org/delay/6", // This should timeout
	}
	
	fmt.Println("Starting concurrent URL health check...")
	fmt.Println("Checking", len(sampleURLs), "URLs with 5-second timeout per request")
	fmt.Println()
	
	// Check all URLs
	results := CheckURLHealth(sampleURLs)
	
	// Display results
	fmt.Println("Results:")
	fmt.Println("==============================")
	
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("URL: %s\nError: %v\nResponse Time: %v\n\n", result.URL, result.Error, result.ResponseTime)
		} else {
			fmt.Printf("URL: %s\nStatus: %d\nResponse Time: %v\n\n", result.URL, result.Status, result.ResponseTime)
		}
	}
}
