package main

import (
	"fmt"
	"net/http"
	"time"
	"sync"
)

// Result holds the health check result for a single URL
type Result struct {
	URL         string
	Status      bool   // true if reachable (HTTP 200), false otherwise
	StatusCode  int    // HTTP status code
	ResponseTime time.Duration
	Error       error
}

func main() {
	// Sample URLs to check
	urls := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.openai.com",
		"https://httpbin.org/200",
		"https://nonexistent.example.com",
	}

	fmt.Printf("Checking %d URLs...\n\n", len(urls))

	// Channel to collect results from goroutines
	results := make(chan Result, len(urls))

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Process each URL
	for _, url := range urls {
		wg.Add(1)
		
		go func(u string) {
			defer wg.Done()
			
			result := Result{URL: u}
			
			// Start timer
			start := time.Now()
			
			// Make HTTP GET request with timeout
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			
			resp, err := client.Get(u)
			
			// Calculate response time
			result.ResponseTime = time.Since(start)
			
			if err != nil {
				result.Status = false
				result.Error = err
			} else {
				result.Status = resp.StatusCode == 200
				result.StatusCode = resp.StatusCode
				resp.Body.Close()
			}
			
			results <- result
		}(url)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Display results
	for result := range results {
		fmt.Printf("%-50s | Status: %v | Code: %d | Time: %v\n", 
			result.URL, result.Status, result.StatusCode, result.ResponseTime)
		
		if result.Error != nil {
			fmt.Printf("             Error: %v\n", result.Error)
		}
		fmt.Println()
	}
}

