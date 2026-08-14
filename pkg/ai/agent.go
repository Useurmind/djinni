package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/useurmind/djinni/pkg/config"
)

func (a *Agent) Execute() (string, error) {
	ctx := context.Background()

	gitTool := NewGitChangedFilesTool(a.WorkingDir)

	gitOutput, err := gitTool.Call(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get changed files: %w", err)
	}

	if gitOutput == "No changes detected." {
		return "", fmt.Errorf("no changes detected in repository")
	}

	var fileContents []string
	for _, path := range a.ReadPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		files, err := a.listFilesInPath(path)
		if err != nil {
			return "", fmt.Errorf("failed to list files in '%s': %w", path, err)
		}
		fileContents = append(fileContents, fmt.Sprintf("\n=== Files in %s ===", path))
		for _, file := range files {
			if stringsHasSuffix(file, ".go") || stringsHasSuffix(file, ".md") || stringsHasSuffix(file, ".yaml") || stringsHasSuffix(file, ".yml") || stringsHasSuffix(file, ".json") {
				content, err := a.readFile(filepath.Join(path, file))
				if err == nil {
					fileContents = append(fileContents, fmt.Sprintf("\n--- %s ---\n%s", file, content))
				}
			}
		}
	}

	allFileContents := stringsJoin(fileContents, "\n")

	prompt := strings.TrimSpace(`
You are an expert software engineer reviewing code changes. Analyze the code changes provided and generate a commit message following best practices.

Format requirements:
1. First line: Short, imperative mood summary (50 chars or less)
2. Empty line
3. Detailed description of the changes
4. Use bullet points for multiple changes if needed

Changes to analyze:
` + gitOutput + allFileContents + `

Generate the commit message:`)

	input := map[string]interface{}{
		"input": prompt,
	}

	result, err := a.executePrompt(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate code: %w", err)
	}

	return result, nil
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func stringsJoin(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return s[0]
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += sep + s[i]
	}
	return result
}

func (a *Agent) executePrompt(ctx context.Context, input map[string]interface{}) (string, error) {
	if prompt, ok := input["input"].(string); ok {
		return a.simplePrompt(ctx, prompt)
	}
	return "", fmt.Errorf("invalid input format")
}

func (a *Agent) simplePrompt(ctx context.Context, prompt string) (string, error) {
	provider := a.Provider
	if provider == nil {
		return "", fmt.Errorf("provider is not set")
	}

	llm, err := a.createLLM(provider)
	if err != nil {
		return "", err
	}

	return llms.GenerateFromSinglePrompt(ctx, llm, prompt)
}

func (a *Agent) createLLM(provider *config.ModelProvider) (llms.Model, error) {
	opts := []openai.Option{
		openai.WithModel(a.ModelID),
	}

	if provider.APIKey != "" {
		opts = append(opts, openai.WithToken(provider.APIKey))
	}

	if provider.APIBase != "" {
		opts = append(opts, openai.WithBaseURL(provider.APIBase))
	}

	return openai.New(opts...)
}

func (a *Agent) listFilesInPath(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		result = append(result, entry.Name())
	}
	return result, nil
}

func (a *Agent) readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
