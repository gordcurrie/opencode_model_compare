# Active Task: Fresh 3-Run Benchmark with gpt-oss Fix

## Background

Runs 1–3 (Feb 20–21, 2026) were completed before a critical fix was applied to `main.go`.
`gpt-oss:20b` failed in all three runs due to a tool-use schema bug — it tried to call a `write`
tool that doesn't exist in opencode's tool set. A prompt augmentation fix is now in place
(see `main.go` around line 208–220, `strings.Contains(modelName, "gpt-oss")` block).

A solo test of gpt-oss with the fix confirmed it now produces correct, compiled, executed code
(89 LOC, 1m50s, run Feb 21 2026 at 7:29 PM).

The goal is to wipe the pre-fix runs and do 3 clean comparative runs.

---

## ~~Step 1: Wipe Pre-Fix Data~~ ✅ DONE

`reviews/` and `results/*.md` have been deleted. Start from Step 2.

---

## Step 2: Run Trial 1

```bash
make clean run
```

Wait for completion (30–60 min). When done:

```bash
# Archive the raw report and generated code
mkdir -p reviews/run1-code
cp results/comparison-report-*.md reviews/comparison-report-run1.md
find output -name "*.go" | while read f; do
  dest="reviews/run1-code/$(echo $f | sed 's|output/||')"
  mkdir -p "$(dirname $dest)"
  cp "$f" "$dest"
done
```

---

## Step 3: Run Trial 2

```bash
make clean run
```

Archive results:

```bash
mkdir -p reviews/run2-code
cp results/comparison-report-*.md reviews/comparison-report-run2.md
find output -name "*.go" | while read f; do
  dest="reviews/run2-code/$(echo $f | sed 's|output/||')"
  mkdir -p "$(dirname $dest)"
  cp "$f" "$dest"
done
```

---

## Step 4: Run Trial 3

```bash
make clean run
```

Archive results:

```bash
mkdir -p reviews/run3-code
cp results/comparison-report-*.md reviews/comparison-report-run3.md
find output -name "*.go" | while read f; do
  dest="reviews/run3-code/$(echo $f | sed 's|output/||')"
  mkdir -p "$(dirname $dest)"
  cp "$f" "$dest"
done
```

---

## Step 5: Write Code Reviews

After all 3 runs are complete and archived, write the following files using the
format conventions in AGENTS.md:

### `reviews/code-review-run1.md`
- Header: `# Code Review: Run 1`
- Fields: `**Run 1 date:**`, `**Report:**`, `**Prompt Task:**`
- Sections (in order): Hardware & Stack, Run 1 Quick Scorecard, Run 1 Individual Code Reviews, Conclusions
- Read the archived Go files in `reviews/run1-code/` to do the review
- Read `reviews/comparison-report-run1.md` for metrics (gen time, LOC, compiled, executed, vet issues)
- Hardware info is in AGENTS.md (copy the same table — hasn't changed)

### `reviews/code-review-run2.md`
- Same structure as run 1

### `reviews/code-review-run3-and-comparison.md`
- Header: `# Code Review: Run 3 + Three-Run Comparison`
- Fields: `**Run 3 date:**`, `**Run 1 date:** | **Run 2 date:**`, `**Report:**`, `**Prompt Task:**`
- Sections: Hardware & Stack, Run 3 Quick Scorecard, Three-Run Generation Time Comparison,
  Three-Run Pass/Fail Consistency, Run 3 Individual Code Reviews,
  Three-Run Code Quality Evolution, Overall 3-Run Rankings, Conclusions

---

## Key Context to Carry Forward

**The fix in `main.go`:** `gpt-oss:20b` gets a prompt suffix explaining that `write`
doesn't exist and showing correct bash tool usage. All other models get unmodified `prompt.txt`.

**Models under test** (from `ollama list` at runtime):
- `qwen3-coder-next:q4_K_M` — 51 GB local
- `qwen3-coder:30b` — 18 GB local
- `glm-4.7-flash:latest` — 19 GB local
- `glm-5:cloud` — cloud API
- `glm-4.7:cloud` — cloud API
- `gpt-oss:20b` — 13 GB local (**should now pass** with the fix)

**Known model quirks to watch for:**
- `glm-4.7:cloud`: always triggers `go vet` redundant newline warning — not a failure
- `glm-4.7-flash:latest`: highly non-deterministic, has had a different bug every run
- `gpt-oss:20b`: if it still fails despite the fix, document exactly where in its opencode.json
  session log it got stuck and what tool it tried to call

**Star rating scale:** ⭐ = 1, ⭐⭐⭐⭐⭐ = 5. Put stars in `###` subsection heading.

**Scorecard column for gpt-oss:** If it passes, it should get a real quality rating (not ⭐ by default).
Evaluate the code in `reviews/runN-code/gpt-oss-20b_*.go` like any other model.
If it fails again, note which step it failed at and what the new error pattern was.

**Correct output definition:** A model's output is "correct" if it compiles, executes,
and produces distinct results showing some URLs reachable and some not (not all passing,
not all failing silently).


