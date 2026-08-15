package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/git"
	"github.com/useurmind/djinni/pkg/log"
)

func (a *Agent) Execute() (string, error) {
	ctx := context.Background()

	gitOutput, err := git.GetChangedFilesWithDiffs(a.WorkingDir)
	if err != nil {
		return "", fmt.Errorf("failed to get changed files: %w", err)
	}

	if gitOutput == "No changes detected." {
		return "", fmt.Errorf("no changes detected in repository")
	}

	prompt := strings.TrimSpace(`
You are an expert software engineer reviewing code changes. Analyze the code changes provided and generate a commit message following best practices.

Format requirements:
1. First line: Short, imperative mood summary (50 chars or less)
2. Empty line
3. Detailed description of the changes
4. Use bullet points for multiple changes if needed

Changes to analyze:
` + gitOutput + `

Generate the commit message:`)

	log.Info("Executing prompt to generate commit message:\n%s", prompt)

	input := map[string]interface{}{
		"input": prompt,
	}

	result, err := a.executePrompt(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate code: %w", err)
	}

	return result, nil
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
