# Code Review: Run 3 + Three-Run Comparison

**Run 3 date:** February 22, 2026  
**Report:** [comparison-report-run3.md](comparison-report-run3.md)  
**Prompt Task:** Single-file Go concurrent URL health checker using goroutines, channels, `sync.WaitGroup`, a 5-second per-request timeout, and a `main()` demonstrating 5 sample URLs.  
**Prior runs:** [code-review-run1.md](code-review-run1.md) · [code-review-run2.md](code-review-run2.md)

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

---

## Run 3 Quick Scorecard

Benchmark results alongside corrected assessment (accounting for false negatives):

| Rank | Model | Gen Time | LOC | Benchmark | Actual | Notes |
|------|-------|----------|-----|-----------|--------|-------|
| 1 | `qwen3-coder-next:q4_K_M` | 7m58s | 161 | ✅ ✅ | ✅ ✅ | Best architecture in run |
| 2 | `qwen3-coder:30b` | 1m44s | 108 | ✅ ✅ | ✅ ✅ | Best version of this model yet |
| 2 | `gpt-oss:20b` | 58.916s | 87 | ✅ ✅ | ✅ ✅ | Consistent second pass; added `context.Context` |
| 4 | `glm-4.7-flash:latest` | 8m5s | ~130 | ❌ ❌ | ✅ ✅ | **FALSE NEGATIVE** — wrote to `./output/` subdir, fixed compile error |
| 4 | `glm-5:cloud` | 19.187s | ~115 | ❌ ❌ | ✅ ✅ | **FALSE NEGATIVE** — compiled and ran, then `rm -f`'d all artifacts |
| 4 | `glm-4.7:cloud` | 25.696s | ~80 | ❌ ❌ | ✅ ✅ | **FALSE NEGATIVE** — stale `go.mod` prevented `go mod init`; code is valid |

**Benchmark score: 3/6. Corrected score: 6/6 — every model generated functional code.**

This is the highest corrected pass rate across all three runs. Three separate false negative conditions all triggered simultaneously: wrong output directory, self-cleanup, and stale go.mod from a prior run.

---

## Run 3 Individual Code Reviews

### 1. `qwen3-coder-next:q4_K_M` ⭐⭐⭐⭐⭐

**Metrics:** 7m58.136s generation · 161 LOC · compiled ✅ · executed ✅

The strongest technical submission in Run 3. The model continued and refined its Run 1 worker-pool-inspired architecture into a cleaner OOP design: a `HealthChecker` struct holds the URL list, buffered channel, timeout, and an embedded `sync.WaitGroup`. A constructor (`NewHealthChecker`), a method wrapper (`checkSingleURL`), and a standalone pure function (`checkURL`) cleanly separate concerns. The most notable addition is `formatStatus`, a helper that maps HTTP status code ranges to human-readable strings ("OK", "Redirect", "Client Error", "Server Error") rather than embedding formatting logic in the output loop.

The package-level comment (`// URL Health Checker - Concurrent Go Program`) and function-level comments are present throughout. The summary section correctly identifies 4xx responses as "failed," which is a reasonable design choice for a health checker (a 404 means the endpoint is broken).

Generation time dropped from 10m59s in Run 1 to 7m58s — likely because the structural pattern had been established and this run produced a cleaner result with less iteration.

**Strengths:**
- `formatStatus` range-based helper is the best output formatting of any submission across all three runs
- OOP design with interface-like separation: `NewHealthChecker` → `CheckAll` → `checkSingleURL` → `checkURL`
- `HealthChecker.wg` is embedded in the struct — avoids WaitGroup parameter passing
- Sample URLs include both real domains (google.com, github.com) and intentionally failing cases (httpstat.us/404, nonexistent domain)
- 7-minute generation for 161 LOC of genuinely high-quality code

**Weaknesses:**
- Output order is non-deterministic (no index-preserving mechanism despite having a struct-based design that could support it)
- Timeout is hardcoded at call site rather than configurable via the struct constructor

---

### 2. `qwen3-coder:30b` ⭐⭐⭐⭐

**Metrics:** 1m44.553s generation · 108 LOC · compiled ✅ · executed ✅

The best version of this model across all three runs. Key improvements over Runs 1 and 2: extensive inline comments throughout (including one explicitly noting `// Pass url as argument to avoid closure issues`), which is rare and educational. The `delay/6` sample URL intentionally exceeds the 5-second timeout, demonstrating that the timeout actually fires — this is the only submission in all three runs that actively tests the timeout path.

The exported `CheckURLHealth` function persists from Run 2 but is now implemented with an inline goroutine closure rather than as a standalone function. `resp.Body.Close()` is called inline (not deferred) — safe here since there's no early return path after a successful response, and the comment explains the intent.

**Strengths:**
- `delay/6` URL explicitly triggers timeout — best test coverage of any submission
- Comment `// Pass url as argument to avoid closure issues` is pedagogically valuable
- Exported `CheckURLHealth` function maintains clean API across runs
- Most thoroughly commented Go code in the entire dataset

**Weaknesses:**
- No summary section
- Output order is non-deterministic
- `resp.Body.Close()` inline rather than deferred — not incorrect, but less idiomatic

---

### 3. `gpt-oss:20b` ⭐⭐⭐⭐

**Metrics:** 58.916s generation · 87 LOC · compiled ✅ · executed ✅

Consistent with Run 2 in structure and quality, with one notable addition: `context.Context` is now passed to `checkURL`. However, the context is not actually used for the HTTP request — `client.Get(url)` is called directly rather than `http.NewRequestWithContext(ctx, "GET", url, nil)`. This is the correct shape of the pattern but an incomplete implementation; the context's cancellation signal would not propagate to the in-flight request.

Documentation remains high-quality. The expanded `Result` struct comment from Run 2 is trimmed slightly but still present. Generation time dropped to 58s, the fastest this model has produced code.

**Strengths:**
- Fastest generation time for `gpt-oss:20b` across all runs (58s, vs. 1m4s in Run 2)
- `context.Context` parameter signals awareness of Go's cancellation patterns
- Sample URLs include `httpstat.us/200?sleep=2000` (2-second simulated delay)
- Consistent doc comment quality

**Weaknesses:**
- `context.Context` is passed but never used in `http.NewRequestWithContext` — incomplete implementation
- No summary section
- Output order non-deterministic

---

### 4. `glm-4.7-flash:latest` ⭐⭐⭐⭐ (FALSE NEGATIVE)

**Metrics:** 8m5.035s generation · ~130 LOC (estimated) · benchmark ❌ · actual: ✅

This is the most technically interesting false negative in the series. The model wrote its output to `./output/health_checker.go` (a subdirectory of its working directory), encountered a compile error (`errors` package imported but unused), fixed it autonomously with `sed -i '' '/\"errors\"/d'`, recompiled successfully, and ran the program — all before the benchmark scanner looked for files. The non-recursive scanner missed the nested file.

The code itself shows the most significant quality improvement of any model across runs. The `URLHealthResult` struct correctly has both `IsReachable bool` and `StatusCode int`. The summary section computes average response times for successful and failed checks separately, plus an overall success rate percentage — by far the most comprehensive statistics output of any submission. An `init()` function configures the default `http.Transport` with `MaxIdleConns`, `IdleConnTimeout`, and `DisableCompression`.

The one structural oddity is a `time.NewTicker(500ms)` created inside `healthCheck` with `ticker.Stop()` called in the result collection loop — the ticker is never actually used for its intended tick purpose and the Stop call is a no-op in this flow. It appears to be vestigial scaffolding from an iterative generation pass.

**Strengths:**
- Autonomous compile error detection and repair via `sed` — the only self-healing submission
- Summary statistics include success rate percentage and separate average response times for pass/fail
- `init()` function for transport configuration shows awareness of HTTP connection pooling
- `IsReachable bool` + `StatusCode int` is the most useful dual-field struct design

**Weaknesses:**
- All sample URLs are expected-reachable real sites (google, github, bing) — no intentional error cases tested
- `time.NewTicker` created but never meaningfully used
- Redundant second `wg.Wait()` call after the result channel loop has already drained (WaitGroup is already at zero)
- Wrote to wrong directory, triggering false negative

---

### 5. `glm-5:cloud` ⭐⭐⭐ (FALSE NEGATIVE)

**Metrics:** 19.187s generation · ~115 LOC (estimated from Run 2 comparison) · benchmark ❌ · actual: ✅

The model generated code, compiled it, ran it successfully, and then executed `rm -f health_checker health_checker.go` — cleaning up both the binary and the source file. No archived code exists for Run 3 (the benchmark had nothing to scan). Functional assessment is based on the session log showing successful compilation and execution output.

The code is assessed as broadly consistent with Run 2 (same generation time range, same structural pattern based on session log evidence), though without the source file this cannot be verified. If consistent with prior runs, the indexed-channel pattern and summary section would still be present.

**Weaknesses:**
- Self-deletion of source artifacts is the most disruptive behavior in the entire dataset — makes auditability impossible
- No archived code: quality assessment is inference only

---

### 6. `glm-4.7:cloud` ⭐⭐⭐ (FALSE NEGATIVE)

**Metrics:** 25.696s generation · ~80 LOC · benchmark ❌ · actual: ✅

A stale `go.mod` from a prior run survived `make clean` (the clean target did not wipe the model's working directory fully), causing `go mod init` to fail in the benchmark setup phase. The code was complete and valid — archived as [run3-code/glm-4.7-cloud/main.go](run3-code/glm-4.7-cloud/main.go).

The code is structurally similar to Run 3's current output (as attached by the user): a `CheckResult` struct with `Status bool`, `checkURL` accepting WaitGroup and channel as parameters, and the same `fmt.Errorf("status code: %d")` for non-200 statuses. This is a regression from Run 1's OOP design (`URLHealthChecker` struct with constructor) and repeats Run 2's simpler procedural approach.

Sample URLs include httpbin 404/500/delay and a nonexistent domain — good coverage. The `✓`/`✗` output markers and millisecond precision (`%.2fms`) are the same compact format seen in Run 1's OOP version.

**Strengths:**
- Good test URL coverage (200, 404, 500, delay, nonexistent domain)
- Millisecond-precision output with `✓`/`✗` markers
- Correct `go func() { wg.Wait(); close(results) }()` pattern

**Weaknesses:**
- `Status bool` loses raw HTTP status code — same regression as Run 2
- OOP design from Run 1 abandoned without evident reason
- Stale go.mod caused false negative (benchmark infrastructure issue)

---

## Three-Run Comparison

### Generation Time by Run

| Model | Run 1 | Run 2 | Run 3 | Trend |
|-------|-------|-------|-------|-------|
| `qwen3-coder-next:q4_K_M` | 10m59s ✅ | 8m20s ❌ | 7m58s ✅ | Faster each run; inconsistent pass |
| `glm-4.7-flash:latest` | 7m7s ✅ | 6m57s ✅ | 8m5s ✅* | Stable ~7–8 min; 3/3 functional |
| `glm-5:cloud` | 21.867s ✅ | 34.507s ✅ | 19.187s ✅* | Fastest cloud; 3/3 functional |
| `glm-4.7:cloud` | 17.718s ✅ | 17.839s ✅ | 25.696s ✅* | Very consistent; 3/3 functional |
| `qwen3-coder:30b` | 2m26s ✅ | 1m53s ✅* | 1m44s ✅ | Improving speed; 3/3 functional |
| `gpt-oss:20b` | 1m46s ❌ | 1m4s ✅ | 58.916s ✅ | Run 1 pre-fix; improving speed |

✅* = functionally passed despite benchmark false negative

### Pass/Fail Consistency (Corrected)

| Model | Run 1 | Run 2 | Run 3 | Functional Passes |
|-------|-------|-------|-------|-------------------|
| `qwen3-coder-next:q4_K_M` | ✅ | ❌ | ✅ | 2/3 |
| `glm-4.7-flash:latest` | ✅ | ✅ | ✅ | 3/3 |
| `glm-5:cloud` | ✅ | ✅ | ✅ | 3/3 |
| `glm-4.7:cloud` | ✅ | ✅ | ✅ | 3/3 |
| `qwen3-coder:30b` | ✅ | ✅ | ✅ | 3/3 |
| `gpt-oss:20b` | ❌¹ | ✅ | ✅ | 2/3² |

¹ Run 1 failure was pre-fix (tool-schema bug). Not a code generation quality issue.  
² If only post-fix runs are counted, `gpt-oss:20b` is 2/2 (100%).

### Code Quality Evolution (Stars)

| Model | Run 1 | Run 2 | Run 3 | Trajectory |
|-------|-------|-------|-------|------------|
| `qwen3-coder-next:q4_K_M` | ⭐⭐⭐⭐⭐ | — (failed) | ⭐⭐⭐⭐⭐ | Consistently excellent when it passes |
| `glm-4.7:cloud` | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Sharp regression after Run 1 |
| `qwen3-coder:30b` | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Stable; best version in Run 3 |
| `gpt-oss:20b` | — (failed) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Consistent quality post-fix |
| `glm-4.7-flash:latest` | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | Improving; best version in Run 3 |
| `glm-5:cloud` | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Consistent; self-deletion is ongoing |

### Overall Three-Run Rankings

Rankings weight: reliability (3/3 passes) > code quality average > generation speed.

| Rank | Model | Reliability | Avg Quality | Avg Speed | Summary |
|------|-------|-------------|-------------|-----------|---------|
| 1 | `qwen3-coder:30b` | 3/3 | ⭐⭐⭐⭐ | ~2m | Best balance of reliability, quality, and speed for a local model |
| 2 | `glm-4.7:cloud` | 3/3 | ⭐⭐⭐⭐ | ~20s | Fastest consistent output; quality regressed after Run 1 |
| 2 | `glm-5:cloud` | 3/3 | ⭐⭐⭐ | ~25s | Fastest and most consistent cloud model; self-deletion is a risk |
| 4 | `gpt-oss:20b` | 2/3¹ | ⭐⭐⭐⭐ | ~1m | High quality post-fix; pre-fix run inflates failure count |
| 5 | `glm-4.7-flash:latest` | 3/3 | ⭐⭐⭐ | ~7m | Reliable but slow; improving quality trend |
| 6 | `qwen3-coder-next:q4_K_M` | 2/3 | ⭐⭐⭐⭐⭐ | ~9m | Highest ceiling; lowest consistency |

¹ Only Run 1 failed (pre-fix). Post-fix reliability is 2/2.

---

## Conclusions

1. **Run 3 has the highest corrected pass rate: 6/6.** Every model generated functional, compiling, executing code. The benchmark's 3/6 score is entirely an artifact of tool limitations (non-recursive scanner, self-cleanup, stale go.mod).

2. **Three root causes for false negatives were identified across the three runs:**
   - Non-recursive `.go` file scanner misses subdirectory writes (`qwen3-coder:30b` Run 2, `glm-4.7-flash` Run 3)
   - Model self-cleanup deletes artifacts before scan (`glm-5:cloud` Run 3)
   - Stale `go.mod` from prior run survives `make clean` (`glm-4.7:cloud` Run 3)

3. **`qwen3-coder:30b` is the most reliable local model.** 3/3 functional passes across all runs, improving quality each run, and consistent generation times around 2 minutes. Run 3 produced the most educational comments of any submission.

4. **`glm-4.7:cloud` regressed sharply after Run 1.** The OOP design and constructor it produced in Run 1 (⭐⭐⭐⭐⭐) was not reproduced. Cloud model non-determinism is real: the same model, same prompt, same hardware produced meaningfully different architecture choices across runs.

5. **`gpt-oss:20b` is a solved problem.** The prompt augmentation fix works. Post-fix, the model consistently produces well-documented, correct code in under 1 minute. The only ongoing quality gap is the `context.Context` parameter that is passed but not wired to the HTTP client.

6. **`qwen3-coder-next:q4_K_M` has the highest ceiling but the lowest consistency.** When it passes, it produces the best code in the run. When it fails, it produces nothing. At 8–11 minutes per run, it is also the most expensive to retry. It is not suitable for deadline-sensitive tasks.

7. **`glm-4.7-flash:latest` showed the most improvement across runs.** Run 3 produced the most statistically comprehensive output (success rate, average response times) and the only self-healing model behavior (autonomous compile error fix via `sed`). Its 3/3 functional reliability is underrated by the benchmark.

8. **Benchmark quality gap:** Three false negatives in one run indicates the benchmark tool needs improvement before future runs. At minimum: recursive `.go` file scanning, preservation of model working directories across detection, and a `make clean` that explicitly wipes model-specific go.mod files.
