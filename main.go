package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ModelResult holds the comprehensive test results for a single model
type ModelResult struct {
	ModelName          string
	GenerationTime     time.Duration
	GeneratedCode      string
	GenerationError    string
	CompileSuccess     bool
	CompileTime        time.Duration
	CompileErrors      string
	ExecutionSuccess   bool
	ExecutionTime      time.Duration
	ExecutionOutput    string
	ExecutionError     string
	CodeQualityMetrics CodeQualityMetrics
}

// CodeQualityMetrics holds code quality analysis results
type CodeQualityMetrics struct {
	LinesOfCode      int
	FormattingIssues string
	VetIssues        string
	HasComments      bool
}

// Config holds configuration for the test run
type Config struct {
	GenerationTimeout time.Duration
	CompileTimeout    time.Duration
	ExecutionTimeout  time.Duration
	OutputDir         string
	ResultsDir        string
	PromptFile        string
}

func main() {
	fmt.Println("🚀 OpenCode Model Comparison Framework")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Configuration
	config := Config{
		GenerationTimeout: 15 * time.Minute, // Allow time for models to iterate and fix errors
		CompileTimeout:    30 * time.Second,
		ExecutionTimeout:  10 * time.Second,
		OutputDir:         "output",
		ResultsDir:        "results",
		PromptFile:        "prompt.txt",
	}

	// Read the prompt
	prompt, err := os.ReadFile(config.PromptFile)
	if err != nil {
		fmt.Printf("❌ Error reading prompt file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📝 Loaded prompt from %s (%d bytes)\n\n", config.PromptFile, len(prompt))

	// Get available models
	models, err := getAvailableModels()
	if err != nil {
		fmt.Printf("❌ Error getting models: %v\n", err)
		os.Exit(1)
	}

	// Filter models if command-line arguments provided
	if len(os.Args) > 1 {
		requestedModels := os.Args[1:]
		var filteredModels []string
		for _, requested := range requestedModels {
			for _, available := range models {
				if available == requested {
					filteredModels = append(filteredModels, available)
					break
				}
			}
		}
		if len(filteredModels) == 0 {
			fmt.Printf("❌ None of the requested models are available\n")
			fmt.Printf("   Requested: %v\n", requestedModels)
			fmt.Printf("   Available: %v\n", models)
			os.Exit(1)
		}
		models = filteredModels
		fmt.Printf("🎯 Testing specific models: %v\n\n", models)
	}

	fmt.Printf("🔍 Found %d models:\n", len(models))
	for _, model := range models {
		fmt.Printf("   - %s\n", model)
	}
	fmt.Println()

	// Create output directories
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("❌ Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(config.ResultsDir, 0755); err != nil {
		fmt.Printf("❌ Error creating results directory: %v\n", err)
		os.Exit(1)
	}

	// Test each model
	var results []ModelResult
	for i, model := range models {
		fmt.Printf("\n[%d/%d] Testing model: %s\n", i+1, len(models), model)
		fmt.Println(strings.Repeat("-", 60))

		result := testModel(model, string(prompt), config)
		results = append(results, result)

		// Print quick summary
		status := "✅"
		if !result.CompileSuccess {
			status = "❌ COMPILE FAILED"
		} else if !result.ExecutionSuccess {
			status = "⚠️  RUN FAILED"
		}
		fmt.Printf("   %s Generated in %v, Compiled in %v\n", status, result.GenerationTime.Round(time.Millisecond), result.CompileTime.Round(time.Millisecond))
	}

	// Generate report
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Generating comparison report...")

	reportPath, err := generateReport(results, config)
	if err != nil {
		fmt.Printf("❌ Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Report saved to: %s\n", reportPath)
	fmt.Println()
	fmt.Println("🎉 All tests complete!")
}

// getAvailableModels fetches the list of available Ollama models
func getAvailableModels() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'ollama list': %w", err)
	}

	// Parse the output
	var models []string
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header line
	if scanner.Scan() {
		// First line is header
	}

	// Parse each model line
	re := regexp.MustCompile(`^(\S+)\s+`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			modelName := matches[1]
			// Include all models (both local and cloud)
			models = append(models, modelName)
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}

	return models, nil
}

// testModel runs the complete test pipeline for a single model
func testModel(modelName, prompt string, config Config) ModelResult {
	result := ModelResult{
		ModelName: modelName,
	}

	// Sanitize model name for directory
	dirName := sanitizeModelName(modelName)
	modelDir := filepath.Join(config.OutputDir, dirName)

	// Create model directory
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		result.GenerationError = fmt.Sprintf("Failed to create directory: %v", err)
		return result
	}

	// Step 1: Generate code with OpenCode
	fmt.Printf("   ⏳ Generating code...\n")
	generatedCode, codeFile, genTime, err := generateCode(modelName, prompt, modelDir, config.GenerationTimeout)
	result.GenerationTime = genTime
	result.GeneratedCode = generatedCode

	if err != nil {
		result.GenerationError = err.Error()
		fmt.Printf("   ❌ Generation failed: %v\n", err)
		return result
	}

	fmt.Printf("   ✅ Code generated (%d bytes) -> %s\n", len(generatedCode), filepath.Base(codeFile))

	// Step 2: Initialize Go module in model directory
	if err := initGoModule(modelDir, dirName); err != nil {
		result.CompileErrors = fmt.Sprintf("Failed to init module: %v", err)
		return result
	}

	// Step 3: Compile the code
	fmt.Printf("   ⏳ Compiling...\n")
	compileSuccess, compileTime, compileErrors := compileCode(modelDir, config.CompileTimeout)
	result.CompileSuccess = compileSuccess
	result.CompileTime = compileTime
	result.CompileErrors = compileErrors

	if !compileSuccess {
		fmt.Printf("   ❌ Compilation failed\n")
		return result
	}
	fmt.Printf("   ✅ Compiled successfully\n")

	// Step 4: Run code quality checks
	result.CodeQualityMetrics = analyzeCodeQuality(codeFile, modelDir)

	// Step 5: Execute the binary
	fmt.Printf("   ⏳ Running program...\n")
	execSuccess, execTime, execOutput, execError := executeCode(modelDir, config.ExecutionTimeout)
	result.ExecutionSuccess = execSuccess
	result.ExecutionTime = execTime
	result.ExecutionOutput = execOutput
	result.ExecutionError = execError

	if execSuccess {
		fmt.Printf("   ✅ Executed successfully\n")
	} else {
		fmt.Printf("   ⚠️  Execution had issues\n")
	}

	return result
}

// sanitizeModelName converts a model name into a safe directory name
func sanitizeModelName(modelName string) string {
	// Replace colons and slashes with dashes
	safe := strings.ReplaceAll(modelName, ":", "-")
	safe = strings.ReplaceAll(safe, "/", "-")
	return safe
}

// normalizeModelName converts Ollama model name to OpenCode's expected format
// OpenCode doesn't recognize :latest tags, so we strip them
func normalizeModelName(modelName string) string {
	// Remove :latest suffix if present (OpenCode expects model names without :latest)
	return strings.TrimSuffix(modelName, ":latest")
}

// generateCode invokes OpenCode to generate code for the prompt
// OpenCode works by running in a directory and creating/modifying files there
// Returns: code content, filepath of generated file, duration, error
func generateCode(modelName, prompt, workDir string, timeout time.Duration) (string, string, time.Duration, error) {
	start := time.Now()

	// Create opencode.json config to allow file edits (needed for models like qwen3-coder)
	// SECURITY: Ultra-restrictive permissions
	// - Working directory is set via --dir flag to the isolated test directory
	// - Only .go files in the working directory root (no subdirectories, no parent access)
	// - No shell commands, no external directory access
	// - Read limited to .go and go.mod files only
	configContent := `{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "edit": {
      "*.go": "allow",
      "*": "deny"
    },
    "write": {
      "*.go": "allow",
      "*": "deny"
    },
    "read": {
      "*.go": "allow",
      "go.mod": "allow",
      "go.sum": "allow",
      "*": "deny"
    },
    "list": "allow",
    "bash": "deny",
    "external_directory": "deny"
  }
}`
	configPath := filepath.Join(workDir, "opencode.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return "", "", time.Since(start), fmt.Errorf("failed to create opencode.json: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Normalize model name for OpenCode (strip :latest tag as OpenCode doesn't recognize it)
	normalizedModel := normalizeModelName(modelName)

	// Run opencode run in the specified directory
	// Use --format json to get structured output
	cmd := exec.CommandContext(ctx, "opencode", "run",
		"--dir", workDir,
		"-m", fmt.Sprintf("ollama/%s", normalizedModel),
		"--format", "json",
		prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", "", duration, fmt.Errorf("timeout after %v", timeout)
		}
		return "", "", duration, fmt.Errorf("opencode failed: %v\nStderr: %s\nStdout: %s", err, stderr.String(), stdout.String())
	}

	// OpenCode creates files in the directory, so find any .go file
	files, _ := filepath.Glob(filepath.Join(workDir, "*.go"))

	var codeFilePath string
	var code []byte

	if len(files) > 0 {
		// Files were created - use the first one
		codeFilePath = files[0]
		code, err = os.ReadFile(codeFilePath)
		if err != nil {
			return "", "", duration, fmt.Errorf("could not read generated code: %v", err)
		}
	} else {
		// No files created - try to extract code from stdout (some models output as markdown)
		extractedCode, filename := extractCodeFromOutput(stdout.String())
		if extractedCode == "" {
			return "", "", duration, fmt.Errorf("no Go files generated\nStdout: %s\nStderr: %s", stdout.String(), stderr.String())
		}

		// Save the extracted code to a file
		if filename == "" {
			filename = "main.go"
		}
		codeFilePath = filepath.Join(workDir, filename)
		if err := os.WriteFile(codeFilePath, []byte(extractedCode), 0644); err != nil {
			return "", "", duration, fmt.Errorf("could not write extracted code: %v", err)
		}
		code = []byte(extractedCode)
	}

	return string(code), codeFilePath, duration, nil
}

// extractCodeFromOutput tries to extract Go code from OpenCode's JSON output
func extractCodeFromOutput(output string) (string, string) {
	// OpenCode outputs JSON events line by line
	// Parse the JSON to extract text fields
	var allText strings.Builder
	var filename string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Try to parse as JSON
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Check if this is a text event
		if eventType, ok := event["type"].(string); ok && eventType == "text" {
			// Extract the part field
			if part, ok := event["part"].(map[string]interface{}); ok {
				if text, ok := part["text"].(string); ok {
					allText.WriteString(text)
					allText.WriteString("\n")
				}
			}
		}
	}

	text := allText.String()

	// First try XML format (qwen3-coder)
	if strings.Contains(text, "<parameter=content>") {
		var codeBlock strings.Builder
		inContent := false

		for _, line := range strings.Split(text, "\n") {
			// Start capturing after <parameter=content>
			if strings.Contains(line, "<parameter=content>") {
				inContent = true
				// Check if content is on same line
				if idx := strings.Index(line, "<parameter=content>"); idx != -1 {
					after := line[idx+len("<parameter=content>"):]
					if !strings.Contains(after, "</parameter>") && strings.TrimSpace(after) != "" {
						codeBlock.WriteString(after)
						codeBlock.WriteString("\n")
					}
				}
				continue
			}

			// Stop at </parameter>
			if inContent && strings.Contains(line, "</parameter>") {
				break
			}

			// Capture content lines
			if inContent {
				codeBlock.WriteString(line)
				codeBlock.WriteString("\n")
			}

			// Look for filepath
			if strings.Contains(line, "<parameter=filePath>") {
				start := strings.Index(line, ">")
				end := strings.Index(line, "</parameter>")
				if start != -1 && end != -1 && start < end {
					fname := strings.TrimSpace(line[start+1 : end])
					fname = strings.TrimPrefix(fname, "/")
					if fname != "" && strings.HasSuffix(fname, ".go") {
						filename = fname
					}
				}
			}
		}

		if codeBlock.Len() > 0 {
			return strings.TrimSpace(codeBlock.String()), filename
		}
	}

	// Try markdown format (gpt-oss)
	var codeBlock strings.Builder
	inCodeBlock := false

	for _, line := range strings.Split(text, "\n") {
		// Look for filename hints in comments
		if strings.Contains(line, "// ") && strings.Contains(line, ".go") && !inCodeBlock {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, ".go") {
					filename = strings.Trim(part, "//.,`\"")
					break
				}
			}
		}

		trimmed := strings.TrimSpace(line)

		// Start of Go code block
		if strings.HasPrefix(trimmed, "```go") {
			inCodeBlock = true
			continue
		}

		// End of code block - but only if it's just ``` alone
		if inCodeBlock && trimmed == "```" {
			inCodeBlock = false
			// Don't break - there might be multiple code blocks
			continue
		}

		// Collect code
		if inCodeBlock {
			codeBlock.WriteString(line)
			codeBlock.WriteString("\n")
		}
	}

	return strings.TrimSpace(codeBlock.String()), filename
}

// initGoModule initializes a Go module in the specified directory
func initGoModule(dir, moduleName string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod init failed: %v\nOutput: %s", err, output)
	}
	return nil
}

// compileCode attempts to compile the Go code in the specified directory
func compileCode(dir string, timeout time.Duration) (bool, time.Duration, string) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", "program")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		return false, duration, string(output)
	}

	return true, duration, ""
}

// executeCode runs the compiled binary
func executeCode(dir string, timeout time.Duration) (bool, time.Duration, string, string) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	binaryPath := filepath.Join(dir, "program")
	// Convert to absolute path to avoid PATH lookup issues
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return false, time.Since(start), "", fmt.Sprintf("failed to resolve path: %v", err)
	}

	cmd := exec.CommandContext(ctx, absPath)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(start)

	if err != nil {
		errorMsg := err.Error()
		if stderr.Len() > 0 {
			errorMsg = stderr.String()
		}
		return false, duration, stdout.String(), errorMsg
	}

	return true, duration, stdout.String(), ""
}

// analyzeCodeQuality performs static analysis on the generated code
func analyzeCodeQuality(codeFile, dir string) CodeQualityMetrics {
	metrics := CodeQualityMetrics{}

	// Count lines of code
	if content, err := os.ReadFile(codeFile); err == nil {
		metrics.LinesOfCode = strings.Count(string(content), "\n")
		metrics.HasComments = strings.Contains(string(content), "//") || strings.Contains(string(content), "/*")
	}

	// Check formatting
	cmd := exec.Command("gofmt", "-d", codeFile)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		metrics.FormattingIssues = string(output)
	}

	// Run go vet
	cmd = exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil && len(output) > 0 {
		metrics.VetIssues = string(output)
	}

	return metrics
}

// generateReport creates a markdown comparison report
func generateReport(results []ModelResult, config Config) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	reportPath := filepath.Join(config.ResultsDir, fmt.Sprintf("comparison-report-%s.md", timestamp))

	var buf bytes.Buffer

	// Header
	buf.WriteString("# OpenCode Model Comparison Report\n\n")
	buf.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("January 2, 2006 at 3:04 PM")))
	buf.WriteString(fmt.Sprintf("**Models Tested:** %d\n\n", len(results)))

	// Summary Table
	buf.WriteString("## Summary\n\n")
	buf.WriteString("| Model | Generation Time | Compiled | Executed | LOC | Has Comments |\n")
	buf.WriteString("|-------|----------------|----------|----------|-----|-------------|\n")

	for _, r := range results {
		compiled := "✅"
		if !r.CompileSuccess {
			compiled = "❌"
		}
		executed := "✅"
		if !r.ExecutionSuccess {
			executed = "❌"
		}
		hasComments := "✅"
		if !r.CodeQualityMetrics.HasComments {
			hasComments = "❌"
		}

		buf.WriteString(fmt.Sprintf("| %s | %v | %s | %s | %d | %s |\n",
			r.ModelName,
			r.GenerationTime.Round(time.Millisecond),
			compiled,
			executed,
			r.CodeQualityMetrics.LinesOfCode,
			hasComments,
		))
	}
	buf.WriteString("\n")

	// Detailed Results
	buf.WriteString("## Detailed Results\n\n")

	for i, r := range results {
		buf.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, r.ModelName))

		// Metrics
		buf.WriteString("**Metrics:**\n")
		buf.WriteString(fmt.Sprintf("- Generation Time: %v\n", r.GenerationTime.Round(time.Millisecond)))
		buf.WriteString(fmt.Sprintf("- Compilation: %s", statusIcon(r.CompileSuccess)))
		if r.CompileSuccess {
			buf.WriteString(fmt.Sprintf(" (%v)\n", r.CompileTime.Round(time.Millisecond)))
		} else {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("- Execution: %s", statusIcon(r.ExecutionSuccess)))
		if r.ExecutionSuccess {
			buf.WriteString(fmt.Sprintf(" (%v)\n", r.ExecutionTime.Round(time.Millisecond)))
		} else {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("- Lines of Code: %d\n", r.CodeQualityMetrics.LinesOfCode))
		buf.WriteString(fmt.Sprintf("- Has Comments: %s\n", statusIcon(r.CodeQualityMetrics.HasComments)))
		buf.WriteString("\n")

		// Errors/Issues
		if r.GenerationError != "" {
			buf.WriteString("**Generation Error:**\n```\n")
			buf.WriteString(r.GenerationError)
			buf.WriteString("\n```\n\n")
		}

		if !r.CompileSuccess && r.CompileErrors != "" {
			buf.WriteString("**Compilation Errors:**\n```\n")
			buf.WriteString(r.CompileErrors)
			buf.WriteString("\n```\n\n")
		}

		if r.CodeQualityMetrics.FormattingIssues != "" {
			buf.WriteString("**Formatting Issues:**\n```diff\n")
			buf.WriteString(r.CodeQualityMetrics.FormattingIssues)
			buf.WriteString("\n```\n\n")
		}

		if r.CodeQualityMetrics.VetIssues != "" {
			buf.WriteString("**Vet Issues:**\n```\n")
			buf.WriteString(r.CodeQualityMetrics.VetIssues)
			buf.WriteString("\n```\n\n")
		}

		// Execution output (truncated)
		if r.ExecutionSuccess && r.ExecutionOutput != "" {
			buf.WriteString("**Execution Output:**\n```\n")
			if len(r.ExecutionOutput) > 500 {
				buf.WriteString(r.ExecutionOutput[:500])
				buf.WriteString("\n... (truncated)")
			} else {
				buf.WriteString(r.ExecutionOutput)
			}
			buf.WriteString("\n```\n\n")
		}

		if !r.ExecutionSuccess && r.ExecutionError != "" {
			buf.WriteString("**Execution Error:**\n```\n")
			buf.WriteString(r.ExecutionError)
			buf.WriteString("\n```\n\n")
		}

		// Generated Code (first 50 lines)
		if r.GeneratedCode != "" {
			buf.WriteString("**Generated Code:**\n```go\n")
			lines := strings.Split(r.GeneratedCode, "\n")
			maxLines := 50
			if len(lines) > maxLines {
				buf.WriteString(strings.Join(lines[:maxLines], "\n"))
				buf.WriteString(fmt.Sprintf("\n... (%d more lines)\n", len(lines)-maxLines))
			} else {
				buf.WriteString(r.GeneratedCode)
			}
			buf.WriteString("\n```\n\n")
		}

		buf.WriteString("---\n\n")
	}

	// Write report
	if err := os.WriteFile(reportPath, buf.Bytes(), 0644); err != nil {
		return "", err
	}

	return reportPath, nil
}

func statusIcon(success bool) string {
	if success {
		return "✅"
	}
	return "❌"
}
