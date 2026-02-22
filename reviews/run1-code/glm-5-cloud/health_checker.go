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
	Error        error
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
		Error:        err,
	}

	if err == nil {
		result.Status = resp.StatusCode
		resp.Body.Close()
	}

	return result
}

func CheckURLs(urls []string, timeout time.Duration) []Result {
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
		"https://www.stackoverflow.com",
		"https://nonexistent.invalid",
		"https://httpbin.org/status/200",
	}

	fmt.Println("Checking URLs concurrently...")
	fmt.Println("============================")

	timeout := 5 * time.Second
	results := CheckURLs(urls, timeout)

	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("URL: %s\n  Status: FAILED\n  Error: %v\n  Response Time: %v\n\n", r.URL, r.Error, r.ResponseTime)
		} else {
			fmt.Printf("URL: %s\n  Status: %d\n  Response Time: %v\n\n", r.URL, r.Status, r.ResponseTime)
		}
	}
}
