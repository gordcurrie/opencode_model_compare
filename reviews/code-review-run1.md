# Code Review: Run 1

**Run 1 date:** February 21, 2026  
**Report:** [comparison-report-run1.md](comparison-report-run1.md)  
**Prompt Task:** Single-file Go concurrent URL health checker using goroutines, channels, `sync.WaitGroup`, a 5-second per-request timeout, and a `main()` demonstrating 5 sample URLs.

---

## Test Hardware & Stack

| Component | Details |
|-----------|---------|
| Machine | Apple Mac mini (Mac16,11) |
| Chip | Apple M4 Pro |
| CPU Cores | 14 (10P + 4E) |
| GPU Cores | 20 (Metal 4) |
| Unified Memory | 64 GB |
| OS | macOS 26.3 (25D125) |
| Ollama | 0.16.2 (Metal backend; MLX unavailable) |
| Go | 1.26.0 darwin/arm64 |

Local model generation times are Metal-backend throughput. MLX (unavailable) would likely be faster.

---

## Run 1 Quick Scorecard

| Rank | Model | Gen Time | LOC | Compiled | Executed | Correct Output | Quality |
|------|-------|----------|-----|----------|----------|----------------|---------|
| 1 | `qwen3-coder-next:q4_K_M` | 10m59s | 157 | ✅ | ✅ | ✅ | ⭐⭐⭐⭐⭐ |
| 1 | `glm-4.7:cloud` | 17.718s | 135 | ✅ | ✅ | ✅ | ⭐⭐⭐⭐⭐ |
| 3 | `qwen3-coder:30b` | 2m26s | 105 | ✅ | ✅ | ✅ | ⭐⭐⭐⭐ |
| 4 | `glm-4.7-flash:latest` | 7m7s | 117 | ✅ | ✅ | ✅ | ⭐⭐⭐ |
| 4 | `glm-5:cloud` | 21.867s | 93 | ✅ | ✅ | ✅ | ⭐⭐⭐ |
| 6 | `gpt-oss:20b` | 1m46s | 0 | ❌ | ❌ | ❌ | ⭐ |

**5/6 models passed** (compiled, executed, and produced correct output).

---

## Run 1 Individual Code Reviews

### 1. `qwen3-coder-next:q4_K_M` ⭐⭐⭐⭐⭐

**Metrics:** 10m59.618s generation · 157 LOC · compiled ✅ · executed ✅

This is the most architecturally complete submission. It uses a **worker pool pattern** (3 workers consuming from a URL channel) rather than the simpler one-goroutine-per-URL approach every other model used. The result struct carries both a string `Status` field ("healthy"/"error"/"failed") and an integer `StatusCode`, giving callers richer information than raw status codes alone. HTTP requests are created via `http.NewRequestWithContext` with an explicit `context.Background()` — following current Go best practice for cancellable requests.

**Strengths:**
- Worker pool with configurable `workerCount` — demonstrates understanding of goroutine management beyond the minimum requirement
- `context`-aware HTTP requests
- Emoji-decorated output (✅/❌) with response times
- Summary section (total/healthy/failed counts)
- Clean `go vet` output

**Weaknesses:**
- 11-minute generation time is by far the longest of the run — the worker pool abstraction clearly required more reasoning
- The `Status` string field duplicates information already available from `StatusCode`; one or the other would suffice

---

### 2. `glm-4.7:cloud` ⭐⭐⭐⭐⭐

**Metrics:** 17.718s generation · 135 LOC · compiled ✅ · executed ✅

The most well-structured code in this run. Uses a `URLHealthChecker` struct with a constructor (`NewURLHealthChecker`) and a method receiver (`checkURL`), making it the only submission that cleanly encapsulates its HTTP client. The constructor hard-codes the 5-second timeout, keeping `main()` clean. Output includes a section separator, emoji markers (✅/✓/⚠/❌), and a final summary table showing reachable vs. failed counts. Notably, it also prints the total wall-clock time of the concurrent run.

**Strengths:**
- Object-oriented design; HTTP client encapsulated in struct
- Separate `printResults` function isolates I/O from logic
- Timing of the entire concurrent run printed at end
- Handles `status != 200` as a warning (⚠) rather than an error, which is more nuanced than most models

**Weaknesses:**
- Generation was fastest-correct but still a cloud model — no local comparison at this speed tier
- Minor: hardcodes `httpbin.org` URLs; a nonexistent domain would better demonstrate error path handling (though `httpbin.org/status/404` and `/status/500` cover non-200 cases well)

---

### 3. `qwen3-coder:30b` ⭐⭐⭐⭐

**Metrics:** 2m26.557s generation · 105 LOC · compiled ✅ · executed ✅

Clean, idiomatic Go. The `checkURL` function takes the channel and WaitGroup as parameters (rather than capturing via closure), which is a valid and explicit style. The result struct (`HealthResult`) has the right fields. Includes a nil check on `resp.Body` before closing — technically unnecessary since a non-nil `resp` always has a non-nil `Body`, but demonstrates careful coding. Output includes varied httpbin status codes (200, 404, 500) and delays, making correct output easy to verify.

**Strengths:**
- Clear separation of concerns across `checkURL`, `checkURLs`, `main`
- Consistent channel/WaitGroup management
- Sample URLs include 404 and 500 status codes — best for verification

**Weaknesses:**
- No summary section — output ends after the per-URL results
- Slightly verbose parameter passing (wg and chan as args instead of closures)

---

### 4. `glm-4.7-flash:latest` ⭐⭐⭐

**Metrics:** 7m7.945s generation · 117 LOC · compiled ✅ · executed ✅

Functional and readable, but with a notable design choice: the `Result.Status` field is a `bool` (true = HTTP 200 only), which loses the actual status code for non-200 responses. This means a 404 and a 500 are both represented the same way. It does include helper functions (`truncateURL`, `truncateError`) for formatted output — a nice usability touch. The goroutine pattern passes WaitGroup and channel as parameters to `checkURL`, which is consistent. Uses `http.StatusOK` constant rather than the magic number 200.

**Strengths:**
- `truncateURL` and `truncateError` helper functions for clean tabular output
- Uses `http.StatusOK` constant
- Header row with column labels in output

**Weaknesses:**
- `Status bool` discards non-200 status codes — a meaningful design regression vs. the other models
- 7-minute generation time for a relatively modest result

---

### 5. `glm-5:cloud` ⭐⭐⭐

**Metrics:** 21.867s generation · 93 LOC · compiled ✅ · executed ✅

The most concise correct submission. Uses an anonymous inline struct `struct{ index int; result Result }` to preserve result ordering across the channel — a clever pattern that ensures output order matches input order. The `Result` struct is minimal (no exported names, `Error` field named `Error` not `Err`). Has a summary section at the end showing reachable vs. failed counts but no comments.

**Strengths:**
- Fast generation (21s for a cloud model)
- Index-preserving channel pattern maintains input order
- Summary section

**Weaknesses:**
- Zero comments — no inline documentation
- Unexported struct and function names (`result`, `checkURL`, `CheckURLs` mix of exported/unexported)
- Minimal error output detail (no response time on error paths)

---

### 6. `gpt-oss:20b` ⭐

**Metrics:** 1m46.862s · 0 LOC · compiled ❌

Failed to produce any code. The opencode.json session log shows three near-empty step cycles — the model generated minimal output then stopped, producing only the `opencode.json` config file. No bash commands were attempted, no file was written. This is the `write`-tool bug: the model's training expects a `write` tool that does not exist in opencode's tool set, and it never fell back to bash-based file writing.

**Note:** A prompt augmentation fix (`strings.Contains(modelName, "gpt-oss")`) has since been applied to `main.go` to pre-instruct this model on the correct bash-based workflow. See Run 2 for results with the fix.

---

## Conclusions

1. **5/6 models generated correct, working Go code** in Run 1. All five passing submissions demonstrated goroutines, channels, and `sync.WaitGroup` correctly.

2. **Architecture varied significantly.** `qwen3-coder-next` used a worker pool; `glm-4.7:cloud` used OOP with a struct/constructor; others used simpler per-URL goroutines. All approaches are valid Go.

3. **Speed vs. size trade-off is real.** Cloud models (`glm-5:cloud` at 21s, `glm-4.7:cloud` at 17s) are 20–40× faster than the largest local model (`qwen3-coder-next` at 11 minutes). Local model quality is competitive, but time cost is high.

4. **`gpt-oss:20b` requires a prompt fix** to work with opencode. Without it, the model does nothing. This is a tool-schema mismatch, not a code generation quality issue.

5. **`glm-4.7-flash:latest` has quality variance** — 7 minutes to generate code that discards HTTP status codes in a bool field is a sign of inconsistency in this model's output quality.
