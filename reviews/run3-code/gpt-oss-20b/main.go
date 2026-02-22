package main

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Result holds the outcome of a single URL check.
// URL is the requested address, Status is the HTTP status code
// (0 if the request failed before a status could be obtained),
// ResponseTime records the time elapsed for the request, and
// Error contains any error that occurred.
type Result struct {
    URL          string
    Status       int
    ResponseTime time.Duration
    Error        error
}

// checkURL performs a single HTTP GET on the provided URL with a 5‑second timeout.
// It writes the result to the out channel and signals completion via the WaitGroup.
func checkURL(ctx context.Context, url string, wg *sync.WaitGroup, out chan<- Result) {
    defer wg.Done()
    start := time.Now()
    // Create an HTTP client with a 5sec timeout. We use the context's timeout.
    client := http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    elapsed := time.Since(start)
    res := Result{URL: url, ResponseTime: elapsed}
    if err != nil {
        res.Error = err
        res.Status = 0
        out <- res
        return
    }
    defer resp.Body.Close()
    res.Status = resp.StatusCode
    out <- res
}

func main() {
    urls := []string{
        "https://www.google.com",
        "https://www.github.com",
        "https://www.thisurldoesnotexist.tld",
        "https://httpstat.us/200?sleep=2000", // 2 sec delay
        "https://httpstat.us/404",
    }

    // Channel to receive results. Buffered to avoid goroutine leak.
    out := make(chan Result, len(urls))
    var wg sync.WaitGroup

    ctx := context.Background()
    for _, u := range urls {
        wg.Add(1)
        go checkURL(ctx, u, &wg, out)
    }

    // Close channel after all goroutines finished.
    go func() {
        wg.Wait()
        close(out)
    }()

    // Aggregate results.
    var results []Result
    for r := range out {
        results = append(results, r)
    }

    // Pretty print.
    fmt.Println("URL Health Check Results:")
    for _, r := range results {
        if r.Error != nil {
            fmt.Printf("%s: ERROR (%v), elapsed %v\n", r.URL, r.Error, r.ResponseTime)
        } else if r.Status == http.StatusOK {
            fmt.Printf("%s: OK (200), elapsed %v\n", r.URL, r.ResponseTime)
        } else {
            fmt.Printf("%s: FAIL (%d), elapsed %v\n", r.URL, r.Status, r.ResponseTime)
        }
    }
}

