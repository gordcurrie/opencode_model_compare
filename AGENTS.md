# AGENTS.md

Instructions for AI coding agents working in this repository.

---

## What This Project Is

A Go benchmarking tool that runs `opencode` (an AI coding CLI) against multiple LLM models, has each model generate a Go program from a fixed prompt, then compiles and executes the output. Results are aggregated into a Markdown report. The purpose is to compare code generation quality, correctness, and consistency across models.

---

## Tech Stack

- **Language:** Go 1.26 (`go.mod` module: `opencode_model_compare`)
- **Build tool:** `make` (see `Makefile` for all targets)
- **Runtime deps:** `ollama` (local models), `opencode` (AI coding CLI), `go`
- **Platform:** macOS (darwin/arm64) — Apple M4 Pro, 64 GB unified memory

---

## Repository Layout

```
main.go              # Entrypoint — orchestrates model comparison
prompt.txt           # The prompt sent to every model (do not modify mid-run)
Makefile             # Build/run/clean targets
go.mod               # Go module file
bin/                 # Compiled binary (git-ignored)
output/              # Generated code per model — wiped on `make clean`
results/             # Auto-generated Markdown reports — wiped on `make clean`
reviews/             # Manually curated reviews and archived runs (DO NOT wipe)
  comparison-report-run1-2026-02-20.md
  comparison-report-run2-2026-02-21.md
  comparison-report-run3-2026-02-21.md
  code-review-run1-2026-02-20.md
  code-review-run2-and-comparison-2026-02-21.md
  code-review-run3-and-3run-comparison-2026-02-21.md
  run2-code/          # Archived Go files from run 2
  run3-code/          # Archived Go files from run 3
```

---

## Common Commands

```bash
make run             # Build and run the full comparison (all models)
make clean run       # Fresh run — wipes output/ and results/ first
make build           # Compile the tool only
make clean           # Remove output/, results/, bin/
make models          # List available Ollama models
make status          # Check if a run is in progress
make report          # Print the latest auto-generated report
make diff            # Show LOC per model's generated output
make deps            # Verify ollama, opencode, and go are on PATH
```

> **Warning:** `make clean` deletes `output/` and `results/`. Back up anything you want to keep into `reviews/` first.

---

## Running a New Comparison

1. Back up current results if needed:
   ```bash
   cp results/*.md reviews/
   find output -name "*.go" | while read f; do
     dest="reviews/runN-code/$(echo $f | sed 's|output/||')"
     mkdir -p "$(dirname $dest)"
     cp "$f" "$dest"
   done
   ```
2. Start fresh:
   ```bash
   make clean run
   ```
3. The run is long (30–60 min). Monitor with `make status`.
4. When done, the report is in `results/comparison-report-<timestamp>.md`.

---

## Key Source File: `main.go`

The main orchestration loop. Key functions:

| Function | Purpose |
|----------|---------|
| `testModel()` | Run the full pipeline for one model |
| `generateCode()` | Invoke `opencode` CLI with 15-min timeout |
| `compileCode()` | Run `go build` on the output with 30s timeout |
| `executeCode()` | Run the compiled binary with 10s timeout |
| `analyzeCodeQuality()` | Run `go vet`, count LOC, check comments |
| `generateReport()` | Produce a Markdown report to `results/` |

Do not increase the `generateCode` timeout beyond 15 minutes without testing — the tool is designed to fail gracefully when a model gets stuck.

---

## The Test Prompt

`prompt.txt` is the single prompt sent to every model. It asks for a single-file Go concurrent URL health checker. **Do not modify `prompt.txt`** between runs — the whole point is every model sees the same stimulus. If you want to test a different prompt, copy the repo first.

---

## Models Under Test

| Model | Type | Size | Notes |
|-------|------|------|-------|
| `qwen3-coder-next:q4_K_M` | Local (Ollama) | 51 GB | Fits in 64 GB unified memory |
| `qwen3-coder:30b` | Local (Ollama) | 18 GB | Best local model in current runs |
| `glm-4.7-flash:latest` | Local (Ollama) | 19 GB | Highly variable generation time |
| `glm-5:cloud` | Cloud API | — | Fastest cloud; normally 26–29s |
| `glm-4.7:cloud` | Cloud API | — | Most consistent; always ~13s |
| `gpt-oss:20b` | Local (Ollama) | 13 GB | Consistently fails — tool-use bug |

The model list is read at runtime from `ollama list`. To add a model, `ollama pull <model>` and re-run.

---

## Reviews Directory Conventions

When writing code reviews into `reviews/`, follow this file-naming and structure convention (established across runs 1–3):

**File naming:**
- Code review: `code-review-run<N>-<date>.md` (or `code-review-run<N>-and-comparison-<date>.md` for multi-run comparisons)
- Raw report archive: `comparison-report-run<N>-<date>.md`
- Backed-up Go files: `run<N>-code/<model>_<filename>.go`

**Review file header format:**
```markdown
# Code Review: Run N [+ Cross-Run Comparison]

**Run N date:** Month DD, YYYY  
**Report:** [comparison-report-runN-<date>.md](comparison-report-runN-<date>.md)  
**Prompt Task:** Single-file Go concurrent URL health checker using goroutines, channels, `sync.WaitGroup`, a 5-second per-request timeout, and a `main()` demonstrating 5 sample URLs.
```

**Required sections (in order):**
1. `## Test Hardware & Stack` — hardware table + notes
2. `## Run N Quick Scorecard` — ranked table: model, gen time, LOC, compiled, executed, correct output, code quality (stars)
3. `## [Cross-run comparison tables]` — if reviewing multiple runs
4. `## Run N Individual Code Reviews` — one `###` subsection per model
5. `## [Summary tables]` — code quality evolution, rankings
6. `## Conclusions` — numbered findings

**Star ratings:** Use ⭐ emoji (`⭐⭐⭐⭐⭐` = 5 stars max). Put stars in the section heading: `### 1. \`model-name\` — Description ⭐⭐⭐⭐`

---

## Known Issues / Watch Out For

- **`gpt-oss:20b`** previously never produced a file. It was trained expecting a `write` tool that doesn't exist in opencode's tool set, and it couldn't correctly fall back to writing via `bash`. This is now mitigated by a model-specific prompt augmentation in `main.go` (`testModel`) that pre-explains the bash-only file-write workflow. After this fix, the model produces correct code and compiles/runs successfully. The fix uses `strings.Contains(modelName, "gpt-oss")` so all other models receive the unmodified `prompt.txt`.
- **`glm-4.7-flash:latest`** is non-deterministic in both generation time (3–9 min) and correctness. It has had a different bug in each of the three runs to date. Do not trust a single passing run.
- **`glm-4.7:cloud`** consistently triggers `go vet`: `fmt.Println arg list ends with redundant newline`. Not a compilation failure; just a style issue. It appears in every run.
- **Background load affects local model times.** Do not co-locate heavy compute tasks with a comparison run.
- **`output/` is wiped by `make clean`.** Always archive to `reviews/` before cleaning.

---

## Hardware Context (for timing interpretation)

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

Local model generation times are Metal-backend throughput on this specific hardware. MLX (not available) would likely be faster.
