package ai

import "context"

type Agent struct {
	Provider   *Provider
	ModelID    string
	Tools      []Tool
	WorkingDir string
	ReadPaths  []string
}

type Provider struct {
	Name    string
	APIBase string
	APIKey  string
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
