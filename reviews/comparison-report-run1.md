# OpenCode Model Comparison Report

**Generated:** February 21, 2026 at 8:02 PM

**Models Tested:** 6

## Summary

| Model | Generation Time | Compiled | Executed | LOC | Has Comments |
|-------|----------------|----------|----------|-----|-------------|
| qwen3-coder-next:q4_K_M | 10m59.618s | ✅ | ✅ | 157 | ✅ |
| glm-4.7-flash:latest | 7m7.945s | ✅ | ✅ | 117 | ✅ |
| glm-5:cloud | 21.867s | ✅ | ✅ | 93 | ✅ |
| glm-4.7:cloud | 17.718s | ✅ | ✅ | 135 | ✅ |
| qwen3-coder:30b | 2m26.557s | ✅ | ✅ | 105 | ✅ |
| gpt-oss:20b | 1m46.862s | ❌ | ❌ | 0 | ❌ |

## Detailed Results

### 1. qwen3-coder-next:q4_K_M

**Metrics:**
- Generation Time: 10m59.618s
- Compilation: ✅ (813ms)
- Execution: ✅ (1.052s)
- Lines of Code: 157
- Has Comments: ✅

**Execution Output:**
```
=== URL Health Checker ===
Checking 5 URLs with 3 workers (5s timeout)

--- Results ---

✅ https://www.google.com
   Status: healthy (HTTP 200)
   Response Time: 157.216542ms

❌ https://nonexistent-domain-12345.com
   Status: FAILED
   Error: Get "https://nonexistent-domain-12345.com": dial tcp: lookup nonexistent-domain-12345.com: no such host

✅ https://www.github.com
   Status: healthy (HTTP 200)
   Response Time: 311.263042ms

✅ https://www.microsoft.com
   Status: healthy (HTTP 200)
... (truncated)
```

**Generated Code:**
```go
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
... (108 more lines)

```

---

### 2. glm-4.7-flash:latest

**Metrics:**
- Generation Time: 7m7.945s
- Compilation: ✅ (273ms)
- Execution: ✅ (741ms)
- Lines of Code: 117
- Has Comments: ✅

**Execution Output:**
```
URL                                                    Status   Response Time      Error
-------------------------------------------------------------------------------------
https://badurl12345.com                                  FAIL       38.0605ms Get "htt...
https://github.com                                         OK    184.333584ms (http.StatusOK)
https://httpbin.org/status/404                           FAIL    210.682708ms (non-200 status)
https://httpbin.org/status/200                
... (truncated)
```

**Generated Code:**
```go
package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Result holds the health check result for a single URL
type Result struct {
	URL          string
	Status       bool   // true if reachable (HTTP 200), false if failed
	ResponseTime time.Duration
	Error        error
}

// checkURL performs a concurrent health check for a single URL
func checkURL(url string, timeout time.Duration, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()
	
	start := time.Now()
	
	// Create a client with timeout
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		results <- Result{
			URL:          url,
			Status:       false,
			ResponseTime: time.Since(start),
			Error:        err,
		}
		return
	}
	defer resp.Body.Close()
	
	responseTime := time.Since(start)
	isReachable := resp.StatusCode == http.StatusOK
	
	results <- Result{
		URL:          url,
		Status:       isReachable,
		ResponseTime: responseTime,
		Error:        nil,
	}
... (68 more lines)

```

---

### 3. glm-5:cloud

**Metrics:**
- Generation Time: 21.867s
- Compilation: ✅ (263ms)
- Execution: ✅ (720ms)
- Lines of Code: 93
- Has Comments: ✅

**Execution Output:**
```
Checking URLs concurrently...
============================
URL: https://www.google.com
  Status: 200
  Response Time: 252.498958ms

URL: https://www.github.com
  Status: 200
  Response Time: 328.062708ms

URL: https://www.stackoverflow.com
  Status: 200
  Response Time: 472.509041ms

URL: https://nonexistent.invalid
  Status: FAILED
  Error: Get "https://nonexistent.invalid": dial tcp: lookup nonexistent.invalid: no such host
  Response Time: 3.321875ms

URL: https://httpbin.org/status/200
  Sta
... (truncated)
```

**Generated Code:**
```go
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
... (44 more lines)

```

---

### 4. glm-4.7:cloud

**Metrics:**
- Generation Time: 17.718s
- Compilation: ✅ (263ms)
- Execution: ✅ (3.458s)
- Lines of Code: 135
- Has Comments: ✅

**Execution Output:**
```
Checking 5 URLs concurrently (5-second timeout each)...

Total time: 3.2059455s (concurrent execution)

URL Health Check Results:
========================

⚠ https://httpbin.org/status/404
   Status: 404 (expected 200 OK)
   Response Time: 205.584292ms

✓ https://www.google.com
   Status: 200 OK
   Response Time: 262.498042ms

✓ https://httpbin.org/status/200
   Status: 200 OK
   Response Time: 367.444959ms

⚠ https://httpbin.org/status/500
   Status: 500 (expected 200 OK)
   Response Ti
... (truncated)
```

**Generated Code:**
```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// URLHealthResult holds the result of checking a single URL
type URLHealthResult struct {
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

// URLHealthChecker handles concurrent URL checking
type URLHealthChecker struct {
	client *http.Client
}

// NewURLHealthChecker creates a new checker with a 5-second timeout
func NewURLHealthChecker() *URLHealthChecker {
	return &URLHealthChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// checkURL performs a health check on a single URL and sends the result to the channel
func (c *URLHealthChecker) checkURL(url string, results chan<- URLHealthResult, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	resp, err := c.client.Get(url)
	responseTime := time.Since(startTime)

	result := URLHealthResult{
		URL:          url,
		ResponseTime: responseTime,
	}

	if err != nil {
		result.Error = err
		result.StatusCode = 0
	} else {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
... (86 more lines)

```

---

### 5. qwen3-coder:30b

**Metrics:**
- Generation Time: 2m26.557s
- Compilation: ✅ (263ms)
- Execution: ✅ (3.459s)
- Lines of Code: 105
- Has Comments: ✅

**Execution Output:**
```
Starting concurrent URL health check...

Results:
========
URL: https://httpbin.org/status/500
Status: 500
Response Time: 202.505ms

URL: https://httpbin.org/status/404
Status: 404
Response Time: 206.364708ms

URL: https://httpbin.org/status/200
Status: 200
Response Time: 235.540542ms

URL: https://httpbin.org/delay/1
Status: 200
Response Time: 1.207008958s

URL: https://httpbin.org/delay/3
Status: 200
Response Time: 3.203073084s


```

**Generated Code:**
```go
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
... (56 more lines)

```

---

### 6. gpt-oss:20b

**Metrics:**
- Generation Time: 1m46.862s
- Compilation: ❌
- Execution: ❌
- Lines of Code: 0
- Has Comments: ❌

**Generation Error:**
```
no Go files generated
Stdout: {"type":"step_start","timestamp":1771725667820,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c83143de9001I7YzCJvO3We5rr","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c8313d22a001S9AF2VA29t7xF6","type":"step-start","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}}
{"type":"step_finish","timestamp":1771725698323,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c8314b4ef001scBt1evKGmzyO2","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c8313d22a001S9AF2VA29t7xF6","type":"step-finish","reason":"unknown","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904","cost":0,"tokens":{"input":0,"output":0,"reasoning":0,"cache":{"read":0,"write":0}}}}
{"type":"step_start","timestamp":1771725698664,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c8314b665001J5lQa8b0rstOIR","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c8314b5320010rCMubwdPLF3nj","type":"step-start","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}}
{"type":"step_finish","timestamp":1771725726438,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c831522ca001SzOj1cH74pMAg2","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c8314b5320010rCMubwdPLF3nj","type":"step-finish","reason":"unknown","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904","cost":0,"tokens":{"input":0,"output":0,"reasoning":0,"cache":{"read":0,"write":0}}}}
{"type":"step_start","timestamp":1771725746238,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c8315703d001Foy0VU4nb3glB7","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c83152304001XbmrGwR6xIylT8","type":"step-start","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}}
{"type":"step_finish","timestamp":1771725746266,"sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","part":{"id":"prt_c8315703e001eV0ZLwF4PoS6Cx","sessionID":"ses_37cec2df9ffeoGiLkGkV6Yw1ni","messageID":"msg_c83152304001XbmrGwR6xIylT8","type":"step-finish","reason":"stop","snapshot":"4b825dc642cb6eb9a060e54bf8d69288fbee4904","cost":0,"tokens":{"total":11591,"input":10858,"output":733,"reasoning":0,"cache":{"read":0,"write":0}}}}

Stderr: 
```

---

