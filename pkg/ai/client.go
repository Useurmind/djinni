package ai

import (
	"fmt"

	"github.com/useurmind/djinni/pkg/config"
)

type Client struct {
	provider *config.ModelProvider
}

func NewClient(provider *config.ModelProvider) (*Client, error) {
	if provider.APIKey == "" {
		return nil, fmt.Errorf("API key is required for model provider '%s'", provider.Name)
	}

	if provider.APIBase == "" {
		return nil, fmt.Errorf("API base URL is required for model provider '%s'", provider.Name)
	}

	return &Client{
		provider: provider,
	}, nil
}

func (c *Client) GetProviderName() string {
	return c.provider.Name
}
