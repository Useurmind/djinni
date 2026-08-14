package ai

import (
	"context"

	"github.com/useurmind/djinni/pkg/config"
)

type Agent struct {
	Provider   *config.ModelProvider
	ModelID    string
	Tools      []Tool
	WorkingDir string
	ReadPaths  []string
}

type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}

type ToolResult struct {
	Name   string
	Input  string
	Output string
	Error  error
}

type CommitMessage struct {
	Title       string
	Description string
}
