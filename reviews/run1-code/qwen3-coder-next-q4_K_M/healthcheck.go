package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthResult holds the result of a single URL health check
type HealthResult struct {
	URL         string
	Status      string
	StatusCode  int
	ResponseTime time.Duration
	Error       error
}

// checkURL performs a health check on a single URL with timeout
func checkURL(url string, timeout time.Duration) HealthResult {
	start := time.Now()

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return HealthResult{
			URL:     url,
			Status:  "failed",
			Error:   err,
		}
	}

	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return HealthResult{
			URL:          url,
			Status:       "failed",
			Error:        err,
			ResponseTime: responseTime,
		}
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		status = "error"
	}

	return HealthResult{
		URL:          url,
		Status:       status,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
	}
}

// worker processes URLs from the input channel and sends results to the output channel
func worker(id int, urls <-chan string, results chan<- HealthResult, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
		result := checkURL(url, timeout)
		results <- result
	}
}

// healthCheck demonstrates concurrent URL health checking
func healthCheck(urls []string, workerCount int, timeout time.Duration) []HealthResult {
	urlChan := make(chan string, len(urls))
	resultChan := make(chan HealthResult, len(urls))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, urlChan, resultChan, timeout, &wg)
	}

	// Send URLs to workers
	go func() {
		for _, url := range urls {
			urlChan <- url
		}
		close(urlChan)
	}()

	// Close results channel when all workers are done
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
	testURLs := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.microsoft.com",
		"https://nonexistent-domain-12345.com",
		"https://www.ibm.com",
	}

	workerCount := 3
	timeout := 5 * time.Second

	fmt.Println("=== URL Health Checker ===")
	fmt.Printf("Checking %d URLs with %d workers (5s timeout)\n\n", len(testURLs), workerCount)

	results := healthCheck(testURLs, workerCount, timeout)

	fmt.Println("--- Results ---")
	fmt.Println()

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s\n", result.URL)
			fmt.Printf("   Status: FAILED\n")
			fmt.Printf("   Error: %v\n", result.Error)
		} else {
			fmt.Printf("✅ %s\n", result.URL)
			fmt.Printf("   Status: %s (HTTP %d)\n", result.Status, result.StatusCode)
			fmt.Printf("   Response Time: %v\n", result.ResponseTime)
		}
		fmt.Println()
	}

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}

	fmt.Printf("--- Summary ---\n")
	fmt.Printf("Total: %d URLs checked\n", len(results))
	fmt.Printf("Healthy: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(results)-successCount)
}
