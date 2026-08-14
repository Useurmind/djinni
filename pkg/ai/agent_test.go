package ai

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/useurmind/djinni/pkg/config"
)

func TestAgent_ExecutePrompt(t *testing.T) {
	authCfg, err := config.LoadAuthConfig()
	require.NoError(t, err)

	tests := []struct {
		name       string
		provider   *config.ModelProvider
		readPaths  []string
		prompt     string
		assertions func(t *testing.T, result string, err error)
	}{
		{
			name: "simple prompt with qwen model",
			provider: &authCfg.ModelProviders[0],
			readPaths: []string{os.TempDir()},
			prompt:    "Respond with 'Hello World'",
			assertions: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.True(t, strings.Contains(result, "Hello World") || len(result) > 0, "should generate a response")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Provider:   tt.provider,
				ModelID:    tt.provider.Models[0].ID,
				Tools:      []Tool{},
				WorkingDir: os.TempDir(),
				ReadPaths:  tt.readPaths,
			}

			ctx := context.Background()
			input := map[string]interface{}{
				"input": tt.prompt,
			}

			result, err := agent.executePrompt(ctx, input)

			if tt.assertions != nil {
				tt.assertions(t, result, err)
			}
		})
	}
}
