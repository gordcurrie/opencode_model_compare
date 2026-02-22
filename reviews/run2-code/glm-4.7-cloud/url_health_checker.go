package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HealthCheckResult struct {
	URL          string
	Status       string
	ResponseTime time.Duration
	Error        error
}

func checkURL(url string, timeout time.Duration, results chan<- HealthCheckResult, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	client := &http.Client{Timeout: timeout}
	result := HealthCheckResult{URL: url}

	resp, err := client.Get(url)
	if err != nil {
		result.Status = "failed"
		result.Error = err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			result.Status = "reachable"
		} else {
			result.Status = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	result.ResponseTime = time.Since(start)
	results <- result
}

func main() {
	urls := []string{
		"http://example.com",
		"http://httpbin.org/status/200",
		"http://httpbin.org/status/404",
		"http://httpbin.org/delay/2",
		"http://example.org",
	}

	results := make(chan HealthCheckResult, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, 5*time.Second, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("URL Health Check Results:")
	fmt.Println("========================")

	for result := range results {
		if result.Error != nil {
			fmt.Printf("URL: %s\nStatus: %s\nError: %v\n", result.URL, result.Status, result.Error)
		} else {
			fmt.Printf("URL: %s\nStatus: %s\nResponse Time: %v\n", result.URL, result.Status, result.ResponseTime)
		}
		fmt.Println()
	}
}
