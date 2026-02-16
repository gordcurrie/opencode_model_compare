.PHONY: all build run clean test help

# Binary name and location
BINARY_NAME=opencode_model_compare
BIN_DIR=bin
BINARY_PATH=$(BIN_DIR)/$(BINARY_NAME)

# Default target
all: build

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BINARY_PATH) .
	@echo "✅ Build complete"

# Run the comparison (builds if needed)
run: build
	@echo "🚀 Running model comparison..."
	@./$(BINARY_PATH)

# Clean all generated files
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf output/
	@rm -rf results/
	@rm -rf $(BIN_DIR)
	@echo "✅ Cleanup complete"

# Run a single model test (usage: make test-one MODEL=gpt-oss:20b)
test-one:
	@echo "Testing single model: $(MODEL)"
	@mkdir -p test-$(MODEL)
	@opencode run --dir test-$(MODEL) -m "ollama/$(MODEL)" "$$(cat prompt.txt)"

# Show available models
models:
	@echo "📋 Available Ollama models:"
	@ollama list

# Check if the comparison is running
status:
	@echo "🔍 Checking process status..."
	@ps aux | grep $(BINARY_NAME) | grep -v grep || echo "No processes running"
	@echo ""
	@echo "📁 Generated outputs:"
	@ls -lh output/ 2>/dev/null || echo "No output directory yet"

# View the latest report
report:
	@echo "📊 Opening latest report..."
	@LATEST=$$(ls -t results/*.md 2>/dev/null | head -1); \
	if [ -n "$$LATEST" ]; then \
		cat "$$LATEST"; \
	else \
		echo "No reports found. Run 'make run' first."; \
	fi

# Quick comparison of all generated code
diff:
	@echo "🔎 Comparing generated code..."
	@for dir in output/*/; do \
		echo "=== $$(basename $$dir) ==="; \
		wc -l $$dir/*.go 2>/dev/null || echo "No Go files"; \
		echo ""; \
	done

# Install/check dependencies
deps:
	@echo "Checking dependencies..."
	@which ollama > /dev/null || echo "⚠️  ollama not found"
	@which opencode > /dev/null || echo "⚠️  opencode not found"
	@which go > /dev/null || echo "⚠️  go not found"
	@echo "✅ Dependencies check complete"

# Show help
help:
	@echo "OpenCode Model Comparison - Available targets:"
	@echo ""
	@echo "  make build      - Build the comparison tool"
	@echo "  make run        - Build and run the full comparison"
	@echo "  make clean      - Remove all generated files"
	@echo "  make models     - List available Ollama models"
	@echo "  make status     - Check if comparison is running"
	@echo "  make report     - Show the latest comparison report"
	@echo "  make diff       - Quick comparison of generated code"
	@echo "  make deps       - Check if dependencies are installed"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run                    # Run full comparison"
	@echo "  make clean && make run      # Fresh run"
	@echo "  make status                 # Check progress"
