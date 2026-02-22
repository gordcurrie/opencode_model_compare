# Code Review: Run 2

**Run 2 date:** February 21, 2026  
**Report:** [comparison-report-run2.md](comparison-report-run2.md)  
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

---

## Run 2 Quick Scorecard

Benchmark results shown alongside corrected assessment (accounting for false negatives):

| Rank | Model | Gen Time | LOC | Benchmark | Actual | Notes |
|------|-------|----------|-----|-----------|--------|-------|
| 1 | `gpt-oss:20b` | 1m4s | 91 | ✅ ✅ | ✅ ✅ | First success after prompt fix |
| 2 | `qwen3-coder:30b` | 1m53s | ~100 | ❌ ❌ | ✅ ✅ | **FALSE NEGATIVE** — wrote to `./output/` subdir |
| 3 | `glm-5:cloud` | 34.507s | 115 | ✅ ✅ | ✅ ✅ | Correct |
| 4 | `glm-4.7-flash:latest` | 6m57s | 87 | ✅ ✅ | ✅ ✅ | Correct |
| 4 | `glm-4.7:cloud` | 17.839s | 74 | ✅ ✅ | ✅ ✅ | Used `http://` instead of `https://` |
| 6 | `qwen3-coder-next:q4_K_M` | 8m20s | 0 | ❌ ❌ | ❌ ❌ | Genuine failure — tried unavailable `Read` tool |

**Benchmark score: 4/6 pass. Corrected score: 5/6 pass.**

The false negative is a benchmark tool limitation: the `.go` file scanner is non-recursive and misses files written to subdirectories of the model's working directory.

---

## Run 2 Individual Code Reviews

### 1. `gpt-oss:20b` ⭐⭐⭐⭐

**Metrics:** 1m4.01s generation · 91 LOC · compiled ✅ · executed ✅

The prompt augmentation fix worked perfectly. This is the first successful run for `gpt-oss:20b`. The code is notably well-documented — the `Result` struct has a multi-line block comment with a usage example, and `checker` and `checkURLs` both have doc comments. The `checkURLs` function cleanly separates the concurrency orchestration from `main()`.

The channel is buffered to `len(urls)`, so the `wg.Wait(); close(ch)` call in `checkURLs` is safe with no separate goroutine needed — the goroutines will never block. Results are appended from the channel in non-deterministic order, but that is acceptable for a health checker.

**Strengths:**
- Best documentation in this run — multi-line struct doc comment with example is exceptional
- Clean separation of `checkURLs` (orchestration) and `checker` (per-URL logic) and `main` (I/O)
- Sample URLs include a `httpbin.org/delay/3` to demonstrate timeout behavior
- Compact, readable output format

**Weaknesses:**
- Output order is non-deterministic (no index-preserving mechanism)
- No summary section

---

### 2. `qwen3-coder:30b` ⭐⭐⭐⭐ (FALSE NEGATIVE)

**Metrics:** 1m53.85s generation · ~100 LOC · benchmark recorded ❌ · actual: ✅

The model wrote its output to `cd output && cat > url_health_checker.go` — a nested subdirectory of its working directory. The benchmark's non-recursive `.go` file scanner missed it, recording zero LOC and a failure. The code itself compiled and ran correctly when tested manually.

This submission uses fully exported functions (`CheckURLHealth`, `CheckURLs`) — the cleanest public API of any model in this run. The goroutine correctly captures the loop variable via a parameter `go func(url string) {...}(url)`, avoiding the classic closure-capture bug. Sample URLs include httpbin 404 and 500 status codes, making output verification straightforward.

**Strengths:**
- Exported API design (`CheckURLHealth`, `CheckURLs`) is the most library-friendly approach
- Goroutine closure variable capture done correctly
- `go func() { wg.Wait(); close(results) }()` pattern is correct
- httpbin 404/500 sample URLs enable real verification

**Weaknesses:**
- Wrote output file to wrong directory, triggering false negative
- No summary section
- Output order is non-deterministic

---

### 3. `glm-5:cloud` ⭐⭐⭐

**Metrics:** 34.507s generation · 115 LOC · compiled ✅ · executed ✅

The index-preserving channel pattern from Run 1 is repeated here, and a summary section has been added (reachable vs. failed counts). This is a welcome improvement over the Run 1 submission. The `Status` field is an `int` (correct — raw status code), and `checkURL` is a standalone function returning a `Result` value rather than sending to a channel directly, which is a cleaner design than passing channels as parameters.

**Strengths:**
- Index-preserving output order (anonymous struct `struct{ index int; result Result }`)
- Summary section with reachable/failed counts
- `checkURL` returns a value rather than side-effecting via channel — testable
- Fast generation (34s)

**Weaknesses:**
- Zero comments
- Summary counts only HTTP 200 as "reachable" — debatable but reasonable
- `Err` field name (vs. conventional `Error` or `Err error`) is inconsistent with stdlib patterns

---

### 4. `glm-4.7-flash:latest` ⭐⭐⭐

**Metrics:** 6m57.926s generation · 87 LOC · compiled ✅ · executed ✅

This run improves on Run 1 by giving the `Result` struct both a `Status bool` and a `StatusCode int`, so non-200 status codes are no longer silently dropped. The inline goroutine closure pattern (anonymous func inside the URL loop) is correct and slightly more idiomatic than passing WaitGroup and channel as function parameters. The channel is buffered to `len(urls)`, and `wg.Wait(); close(results)` is called inline in `main()` — safe because the buffer prevents goroutines from blocking.

Output format includes a header row with aligned columns using `%-50s` formatting.

**Strengths:**
- Fixed the `Status bool` issue from Run 1 by adding `StatusCode int`
- Inline goroutine closure captures loop variable correctly via `func(u string)`
- Output includes column headers

**Weaknesses:**
- 7-minute generation for a modest result
- `Status bool` field is now redundant — `StatusCode` alone is sufficient
- No summary section

---

### 5. `glm-4.7:cloud` ⭐⭐⭐

**Metrics:** 17.839s generation · 74 LOC · compiled ✅ · executed ✅

This is a regression from Run 1. The OOP design (`URLHealthChecker` struct with constructor) has been dropped in favor of a simpler procedural approach. More notably, all sample URLs use `http://` instead of `https://` — a functional regression that would cause most modern endpoints to redirect rather than respond. The `Status` field is now a string ("reachable", "HTTP 404", "failed") which conveys semantic meaning but loses the raw numeric code.

**Strengths:**
- Fast generation (17s)
- Correct `go func() { wg.Wait(); close(results) }()` pattern
- String status field is human-readable

**Weaknesses:**
- `http://` instead of `https://` for all sample URLs — correctness regression
- Lost the OOP design and constructor from Run 1
- Code is shorter (74 LOC) but structurally simpler — not an improvement

---

### 6. `qwen3-coder-next:q4_K_M` ⭐ (Genuine Failure)

**Metrics:** 8m20.129s generation · 0 LOC · compiled ❌

This is a genuine failure, not a false negative. The model attempted to use a `Read` tool that does not exist in opencode's tool set. Rather than falling back to bash-based exploration, it spent approximately 8 minutes executing various bash search commands (`find`, `ls`, `cat`) looking for existing Go project structure, then appears to have given up without writing any code. Only `opencode.json` was produced.

This contrasts with Run 1 where this model succeeded with a worker pool design. The failure mode appears to be a tool-schema confusion triggered by the session context.

---

## Conclusions

1. **`gpt-oss:20b` is now fixed.** The prompt augmentation added in `main.go` completely resolves the tool-schema mismatch. The model produced well-documented, correct code in its first successful run.

2. **The benchmark tool has a false negative problem.** Two models (`qwen3-coder:30b` in this run, `glm-5:cloud` in later runs) wrote files to unexpected locations that the non-recursive scanner missed. The corrected pass rate is 5/6, not 4/6.

3. **`glm-4.7:cloud` regressed.** It dropped its clean OOP design from Run 1 and switched from `https://` to `http://`. Cloud models are non-deterministic even with equivalent prompts.

4. **`qwen3-coder-next:q4_K_M` is inconsistent.** Run 1 produced the best code in the run; Run 2 produced nothing. This is a risk pattern for large local models: high ceiling, but sensitive to session context and tool availability.

5. **`glm-4.7-flash:latest` improved incrementally** (added `StatusCode int`) but remains a slow, inconsistent performer — 7 minutes for code that is broadly equivalent to faster cloud models.
