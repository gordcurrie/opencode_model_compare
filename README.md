# OpenCode Model Comparison Framework

A comprehensive testing framework for comparing the performance of different Ollama models using OpenCode CLI.

## Overview

This tool automatically:

- Detects available Ollama models on your system
- Runs the same coding prompt through each model via OpenCode
- Generates isolated output for each model (avoiding Go package conflicts)
- Compiles and executes the generated code
- Performs code quality analysis
- Produces a detailed comparison report

## Features

- **Fully Automated**: One command runs all tests
- **Security-First**: Each model runs in an isolated directory with minimal permissions (no shell access, no external directory access)
- **Comprehensive Metrics**: Generation time, compilation success, execution results, LOC, code quality
- **Multi-Format Support**: Handles both file-writing models (GLM) and markdown output models (gpt-oss)
- **Smart Extraction**: Automatically extracts code from JSON/XML/markdown output formats

## Requirements

- Go 1.18+
- [Ollama](https://ollama.ai/) installed and running
- [OpenCode](https://opencode.ai/) CLI configured with Ollama
- Local Ollama models downloaded

## Quick Start

1. **Run the comparison:**

   ```bash
   make run
   ```

   Or just:

   ```bash
   make
   ```

2. **View results:**

   ```bash
   make report              # View the comparison report
   make status              # Check progress
   make diff                # Quick comparison of generated code
   ```

3. **Clean up:**

   ```bash
   make clean
   ```

4. **See all commands:**

   ```bash
   make help
   ```

## Results

Current test results show **4 out of 5 models (80%) successfully generate, compile, and execute** working Go programs:

| Model | Status | LOC | Time |
|-------|--------|-----|------|
| glm-4.7-flash | ✅ | ~130 | 3-7min |
| glm-5 | ✅ | ~115 | 15-90s |
| glm-4.7 | ✅ | ~100 | 8-14s |
| gpt-oss:20b | ✅ | ~100 | 40-100s |
| qwen3-coder | ⚠️ Inconsistent | 0-110 | 55-100s |

**Note**: qwen3-coder sometimes outputs code via XML text format (which works), other times tries to write files directly and gets permission rejected (fails).

## How It Works

### Workflow

1. **Discovery**: Scans `ollama list` for available local models
2. **Security Setup**: Creates `opencode.json` in each test directory with minimal permissions:
   - Only .go file creation/editing
   - No shell command execution
   - No external directory access
   - Read-only access to .go, go.mod, go.sum files
3. **Generation**: For each model, invokes `opencode run --dir <dir> -m "ollama/<model>" --format json` with the prompt
4. **Extraction**: Smart extraction from multiple formats:
   - Direct file creation (GLM models)
   - Markdown code blocks (gpt-oss)
   - XML parameter content (qwen3-coder)
5. **Isolation**: Saves generated code to `output/<model-name>/` with its own `go.mod`
6. **Compilation**: Attempts to build the code with `go build`
7. **Execution**: Runs the compiled binary with absolute path resolution
8. **Analysis**: Checks formatting (gofmt), linting (go vet), and code metrics
9. **Reporting**: Generates markdown report comparing all models

### Directory Structure

```txt
.
├── bin/                          # Compiled binary (gitignored)
│   └── opencode_model_compare
├── prompt.txt                    # Input prompt for code generation
├── main.go                       # Test orchestrator
├── Makefile                      # Build automation
├── output/                       # Generated code (gitignored)
│   ├── model-name-1/
│   │   ├── opencode.json        # Security config
│   │   ├── go.mod
│   │   ├── main.go              # Model 1's generated code
│   │   └── program              # Compiled binary
│   └── model-name-2/
│       ├── opencode.json        # Security config
│       ├── go.mod
│       └── main.go              # Model 2's generated code
└── results/                      # Test reports (gitignored)
    └── comparison-report-*.md
```

## Security

Each model runs with **ultra-restrictive permissions** defined in `opencode.json`:

```json
{
  "permission": {
    "edit": { "*.go": "allow", "*": "deny" },
    "write": { "*.go": "allow", "*": "deny" },
    "read": { "*.go": "allow", "go.mod": "allow", "go.sum": "allow", "*": "deny" },
    "bash": "deny",
    "external_directory": "deny"
  }
}
```

This ensures models can **ONLY**:
- Create/edit .go files in their isolated test directory
- Read .go, go.mod, go.sum files
- List directory contents

Models **CANNOT**:
- Execute shell commands
- Access parent directories or other project files
- Create non-.go files
- Access external directories

## Configuration

Edit the `Config` struct in `main.go` to adjust:

- `GenerationTimeout`: Max time for code generation (default: 8 minutes)
- `CompileTimeout`: Max time for compilation (default: 30 seconds)
- `ExecutionTimeout`: Max time for program execution (default: 10 seconds)
- `PromptFile`: Path to prompt file (default: `prompt.txt`)

## Metrics Collected

### Performance

- Code generation time
- Compilation time
- Execution time

### Correctness

- Compilation success/failure
- Execution success/failure
- Compiler error messages
- Runtime error messages

### Code Quality

- Lines of code
- Presence of comments
- Formatting compliance (gofmt)
- Static analysis issues (go vet)

## Report Format

The generated markdown report includes:

- **Summary table**: Quick comparison of all models
- **Detailed results**: Per-model metrics, errors, and generated code snippets
- **Timestamps**: When the test was run

## Example Usage

```bash
# Run comparison with default settings
make run

# Check what models are available
make models

# Monitor progress while it's running
make status

# View the latest report without opening a file
make report

# Clean and start fresh
make clean run

# View a specific model's output
cat output/qwen3-coder-30b/main.go

# Compare two model outputs directly
diff output/model-1/main.go output/model-2/main.go

# Quick comparison of all generated code
make diff
```

## Makefile Targets

- `make build` - Build the comparison tool
- `make run` - Build and run the full comparison  
- `make clean` - Remove all generated files
- `make models` - List available Ollama models
- `make status` - Check if comparison is running and show output
- `make report` - Display the latest comparison report
- `make diff` - Quick comparison of generated code sizes
- `make deps` - Verify all dependencies are installed
- `make help` - Show all available targets

## Notes

- **Cloud models are skipped**: Models with `:cloud` suffix are excluded (they require remote API access)
- **Each model gets isolated directory**: Prevents Go package naming conflicts
- **Generated code is preserved**: You can manually inspect, test, or modify any model's output
- **Sequential execution**: Models are tested one at a time to avoid resource contention and ensure accurate timing

## Troubleshooting

**No models found:**

```bash
ollama list  # Verify models are installed
ollama pull codellama  # Download a model
```

**OpenCode not found:**

```bash
which opencode  # Verify OpenCode is installed and in PATH
```

**Permission denied on scripts:**

```bash
chmod +x run-comparison.sh clean-outputs.sh
```

## License

MIT
