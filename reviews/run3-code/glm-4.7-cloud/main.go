package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type CheckResult struct {
	URL         string
	Status      bool
	ResponseTime time.Duration
	Error       error
}

func checkURL(url string, results chan<- CheckResult, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	result := CheckResult{URL: url}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Error = err
		result.Status = false
	} else {
		defer resp.Body.Close()
		result.Status = resp.StatusCode == http.StatusOK
		if !result.Status {
			result.Error = fmt.Errorf("status code: %d", resp.StatusCode)
		}
	}

	results <- result
}

func main() {
	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/500",
		"https://httpbin.org/delay/1",
		"https://nonexistent.example.com",
	}

	results := make(chan CheckResult, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Concurrent URL Health Checker Results:")
	fmt.Println("======================================")

	for result := range results {
		if result.Status {
			fmt.Printf("✓ %s - OK (%.2fms)\n", result.URL, float64(result.ResponseTime.Microseconds())/1000)
		} else {
			fmt.Printf("✗ %s - FAILED (%.2fms): %v\n", result.URL, float64(result.ResponseTime.Microseconds())/1000, result.Error)
		}
	}

	fmt.Println("\nAll checks completed.")
}
