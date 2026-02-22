package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Result holds the HTTP result for a single URL check.
// URL   – the URL that was checked.
// Code  – the HTTP status code returned (e.g., 200, 404).
// Time  – how long it took for the request to complete.
// Err   – any error that occurred, such as a timeout.
// The struct is exported so callers can inspect all fields.
// The Error field is of type error to allow nil when the request
// succeeded without network issues, regardless of the status code.
//
// Example:
//  result := Result{URL: "https://example.com", Code: 200, Time: 123 * time.Millisecond}
type Result struct {
    URL  string
    Code int
    Time time.Duration
    Err  error
}

// checker performs the HTTP GET request to the given URL.
// It uses a 5-second timeout per request, as required.
// After obtaining the response, it records the status code and
// elapsed time, sending the Result on the provided channel.
func checker(url string, ch chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    client := &http.Client{Timeout: 5 * time.Second}
    start := time.Now()
    resp, err := client.Get(url)
    elapsed := time.Since(start)

    if err != nil {
        ch <- Result{URL: url, Code: 0, Time: elapsed, Err: err}
        return
    }
    defer resp.Body.Close()
    ch <- Result{URL: url, Code: resp.StatusCode, Time: elapsed, Err: nil}
}

// checkURLs concurrently checks each URL in the slice and returns a slice
// of Result structs in the same order as the input slice.
func checkURLs(urls []string) []Result {
    var wg sync.WaitGroup
    ch := make(chan Result, len(urls))

    for _, u := range urls {
        wg.Add(1)
        go checker(u, ch, &wg)
    }

    // Wait for all goroutines to finish.
    wg.Wait()
    close(ch)

    results := make([]Result, 0, len(urls))
    for r := range ch {
        results = append(results, r)
    }
    return results
}

func main() {
    // Sample URLs for demonstration. Feel free to replace with real endpoints.
    urls := []string{
        "https://example.com",
        "https://httpbin.org/status/200",
        "https://httpbin.org/status/404",
        "https://httpbin.org/delay/3", // 3s delay
        "https://nonexistent.domain.example",
    }

    fmt.Println("Checking URLs concurrently...")
    results := checkURLs(urls)

    fmt.Println("Results:")
    for _, r := range results {
        if r.Err != nil {
            fmt.Printf("%s -> Error: %v (took %v)\n", r.URL, r.Err, r.Time)
            continue
        }
        fmt.Printf("%s -> Status %d, Time %v\n", r.URL, r.Code, r.Time)
    }
}

