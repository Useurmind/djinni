package ai

import "fmt"

type Client struct {
	provider *Provider
}

func NewClient(provider *Provider) (*Client, error) {
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
