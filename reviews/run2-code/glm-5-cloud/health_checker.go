package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	URL          string
	Status       int
	ResponseTime time.Duration
	Err          error
}

func checkURL(url string, timeout time.Duration) Result {
	start := time.Now()
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	elapsed := time.Since(start)

	result := Result{
		URL:          url,
		ResponseTime: elapsed,
		Err:          err,
	}

	if err == nil {
		result.Status = resp.StatusCode
		resp.Body.Close()
	}

	return result
}

func checkURLs(urls []string, timeout time.Duration) []Result {
	results := make([]Result, len(urls))
	var wg sync.WaitGroup
	ch := make(chan struct {
		index int
		result Result
	}, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			r := checkURL(u, timeout)
			ch <- struct {
				index  int
				result Result
			}{idx, r}
		}(i, url)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for item := range ch {
		results[item.index] = item.result
	}

	return results
}

func main() {
	urls := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.example.com",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
	}

	fmt.Println("Checking URLs concurrently...")
	fmt.Println("================================")

	timeout := 5 * time.Second
	results := checkURLs(urls, timeout)

	for _, r := range results {
		fmt.Printf("\nURL: %s\n", r.URL)
		if r.Err != nil {
			fmt.Printf("  Status: FAILED\n")
			fmt.Printf("  Error: %v\n", r.Err)
		} else {
			fmt.Printf("  Status: %d\n", r.Status)
			if r.Status == 200 {
				fmt.Printf("  Reachable: YES\n")
			} else {
				fmt.Printf("  Reachable: NO (non-200 status)\n")
			}
		}
		fmt.Printf("  Response Time: %v\n", r.ResponseTime)
	}

	fmt.Println("\n================================")
	fmt.Println("Summary:")
	reachable := 0
	failed := 0
	for _, r := range results {
		if r.Err == nil && r.Status == 200 {
			reachable++
		} else {
			failed++
		}
	}
	fmt.Printf("Reachable (HTTP 200): %d\n", reachable)
	fmt.Printf("Failed: %d\n", failed)
}
