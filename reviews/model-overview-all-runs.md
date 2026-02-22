# Three-Run Model Overview

**Runs conducted:** February 21–22, 2026  
**Task:** Single-file Go concurrent URL health checker — goroutines, channels, `sync.WaitGroup`, 5-second per-request timeout, `main()` with 5 sample URLs.  
**Benchmark tool fixes applied after Run 3:** recursive `.go` scanner, idempotent `go mod init`.

---

## At a Glance

| Model | Type | Size | Functional Passes | Avg Gen Time | Avg Quality | Verdict |
|-------|------|------|-------------------|--------------|-------------|---------|
| `qwen3-coder:30b` | Local | 18 GB | 3/3 | ~2m | ⭐⭐⭐⭐ | **Best overall local model** |
| `glm-4.7:cloud` | Cloud | — | 3/3 | ~20s | ⭐⭐⭐⭐ | **Fastest; quality regressed** |
| `glm-5:cloud` | Cloud | — | 3/3 | ~25s | ⭐⭐⭐ | **Consistent but self-deletes** |
| `gpt-oss:20b` | Local | 13 GB | 2/3¹ | ~1m | ⭐⭐⭐⭐ | **Solved with prompt fix** |
| `glm-4.7-flash:latest` | Local | 19 GB | 3/3 | ~7m | ⭐⭐⭐ | **Reliable but slow and verbose** |
| `qwen3-coder-next:q4_K_M` | Local | 51 GB | 2/3 | ~9m | ⭐⭐⭐⭐⭐ | **Best ceiling; worst consistency** |

¹ Run 1 failure was a pre-fix tool-schema bug, not a code quality issue. Post-fix: 2/2.

---

## Model-by-Model

### `qwen3-coder:30b`

**Pros:**
- Perfect 3/3 functional reliability across all runs — the only local model that never failed or regressed
- Code quality improved each run; Run 3 was the most thoroughly commented submission in the entire dataset, including a rare `// Pass url as argument to avoid closure issues` note
- Exported API design (`CheckURLHealth`, `CheckURLs`) — the most library-friendly approach
- Run 3 intentionally included a `delay/6` URL to trigger the timeout path — the only model that actively tested the timeout

**Cons:**
- No summary section in any run
- Output order is non-deterministic across all three runs
- Occasionally wrote to unexpected paths (Run 2 false negative)

**Bottom line:** The most dependable local model. If you need a local model you can trust to produce working, well-commented Go code in ~2 minutes, this is it.

---

### `glm-4.7:cloud`

**Pros:**
- Fastest consistent output across the dataset — ~17–25s per run
- Run 1 produced the best-structured code of any cloud model: OOP design with `URLHealthChecker` struct, constructor, method receivers, clean I/O separation, and a total-run timer
- 3/3 functional reliability (all three false negatives were infrastructure issues, not model failures)

**Cons:**
- Sharp quality regression after Run 1: the OOP design was abandoned, `https://` became `http://`, and `Status bool` dropped the raw HTTP status code
- High run-to-run variance in code structure — you cannot predict what architecture you'll get
- Run 3 false negative caused by stale `go.mod` (now fixed in benchmark tool)

**Bottom line:** Fastest when you need a result quickly. But the quality spread is wide — Run 1's output is outstanding, Runs 2 and 3 are mediocre. Not a model to depend on for consistent code architecture.

---

### `glm-5:cloud`

**Pros:**
- Most consistent generation time of any model: 19–34s across all three runs
- The index-preserving channel pattern (anonymous struct `struct{ index int; result Result }`) maintains output order — a pattern none of the other models used
- Summary section with reachable/failed counts present in all runs
- Run 2 added a proper summary with counts; consistent functional quality

**Cons:**
- Zero inline comments in every run — the least documented code in the dataset
- Run 3: deleted its own source file and binary after a successful run, causing a false negative and leaving no auditable artifact (now partially mitigated by infrastructure fixes, but self-deletion cannot be fully prevented)
- Summary only counts HTTP 200 as "reachable," ignoring 2xx range

**Bottom line:** Fast, consistent, and functionally correct — but unauditable when it decides to clean up after itself. The self-deletion behavior is the single biggest ongoing risk for benchmark use.

---

### `gpt-oss:20b`

**Pros:**
- Best documentation quality post-fix: multi-line struct doc comment with usage example, doc comments on all functions — standards that every other model fell short of
- Clean separation of `checker` (per-URL), `checkURLs` (orchestration), and `main` (I/O)
- Improving generation time each run: 1m46s → 1m4s → 58s
- Run 3 added `context.Context` parameter, indicating awareness of Go cancellation idioms

**Cons:**
- Run 1: complete failure due to tool-schema mismatch (`write` tool not available). Required a prompt augmentation workaround that only benefits this model
- The `context.Context` added in Run 3 is passed to `checkURL` but never wired to `http.NewRequestWithContext` — the pattern is correct in shape but incomplete in execution
- No summary section in any run

**Bottom line:** A solved problem post-fix, and the best-documented model in the set. The context.Context gap is the only remaining quality issue. Would benefit from one more run to confirm consistency.

---

### `glm-4.7-flash:latest`

**Pros:**
- 3/3 functional reliability — never genuinely failed
- Most improved across runs: Run 1 used `Status bool` (bad); Run 2 added `StatusCode int`; Run 3 added an `init()` transport configuration, separate average response time statistics for pass/fail, and a success rate percentage — the most comprehensive statistics output of any submission
- Run 3's autonomous self-repair (detected `"errors"` unused import, fixed with `sed`, recompiled) is a remarkable behavior no other model demonstrated

**Cons:**
- Slowest model in the dataset at ~7–8 minutes per run — worse per-token throughput than `qwen3-coder:30b` for equivalent or lower quality output
- All three runs used only expected-reachable sample URLs (google.com, github.com, bing.com) — never probed error paths in `main()`
- Wrote to a subdirectory in both Run 2 and Run 3, causing false negatives (partially mitigated by the recursive scanner fix)
- `time.NewTicker` in Run 3 was created but served no functional purpose — vestigial generation artifact

**Bottom line:** Reliable and improving, but the cost (7–8 minutes per run) is hard to justify against models that produce comparable or better code in a fraction of the time. The self-repair behavior is genuinely impressive but also a sign the model generates errors it then has to fix.

---

### `qwen3-coder-next:q4_K_M`

**Pros:**
- Highest code quality ceiling in the entire dataset — both passing runs produced the best architecture of their respective runs
- Run 1: worker pool pattern with `context`-aware HTTP requests and emoji output with summary
- Run 3: full OOP design (`HealthChecker` struct, `NewHealthChecker` constructor, `formatStatus` range helper mapping status code ranges to strings) — the most idiomatic and extensible design across all 18 submissions
- Generation time improved each passing run: 10m59s → 7m58s

**Cons:**
- Run 2: complete failure — spent 8+ minutes trying to use a non-existent `Read` tool, then did `find`/`grep` searches and produced nothing. Same prompt, same hardware as Run 1 which passed
- At 51 GB and ~8–11 minutes per run, retrying a failure is enormously expensive
- Output order is non-deterministic despite having a struct design that could easily support ordering
- No summary section in Run 1; summary logic in Run 3 counts 4xx as "failed" (reasonable choice, but different from other models)

**Bottom line:** When it works, it produces the best code. When it doesn't, you've burned 8 minutes and gotten nothing. Not suitable for pipelines where reliability matters more than peak quality.

---

## Key Findings Across All Runs

1. **Cloud models are 15–30× faster than local models.** The fastest local model (`qwen3-coder:30b`) averages ~2 minutes; the slowest cloud model (`glm-5:cloud`) averages ~25 seconds. For iteration speed, cloud wins decisively.

2. **Local model quality can match or exceed cloud model quality** when the local model succeeds. `qwen3-coder-next:q4_K_M` in Run 3 produced better-designed code than any cloud model in any run. But reliability is lower.

3. **Benchmark tool false negatives were pervasive.** Of the 18 model-run combinations, at least 4 were miscounted failures. The corrected pass rates:
   - Run 1: 5/6 (benchmark: 5/6 — accurate)
   - Run 2: 5/6 (benchmark: 4/6 — one false negative)
   - Run 3: 6/6 (benchmark: 3/6 — three false negatives)
   
   The recursive `.go` scanner and idempotent `go mod init` fixes eliminate two of the three failure modes.

4. **The `gpt-oss:20b` tool-schema bug is a solved problem.** The prompt augmentation fix works completely and reliably. This model should be treated as a 2/2 post-fix performer, not a 1/3 overall performer.

5. **Code quality is more variable than pass/fail.** All models can write a health checker that compiles and runs. The spread is in architecture choices, documentation, error handling nuance, and test URL selection. These differences matter in production code review but not in a pass/fail benchmark.

6. **No model consistently produced output in the same order.** Only `glm-5:cloud` used an index-preserving channel pattern. All others produce non-deterministic output order, which is functionally acceptable but worth noting in a code review context.
