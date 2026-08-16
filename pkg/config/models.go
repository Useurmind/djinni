package config

import (
	"fmt"
	"strings"
)

// FindModelInProvider finds a model by ID within a specific provider.
// Returns the model if found, nil otherwise.
func FindModelInProvider(provider *ModelProvider, modelID string) *Model {
	for i := range provider.Models {
		if strings.EqualFold(provider.Models[i].ID, modelID) {
			return &provider.Models[i]
		}
	}
	return nil
}

// FindModelGlobal finds a model by ID across all providers.
// If providerName is specified, searches only that provider.
// Returns error if model is not found or if duplicate models exist across providers (when providerName is empty).
func FindModelGlobal(cfg *GlobalConfig, modelID string, providerName string) (*ModelProvider, *Model, error) {
	if len(cfg.ModelProviders) == 0 {
		return nil, nil, fmt.Errorf("no model providers configured")
	}

	if providerName != "" {
		return findModelInNamedProvider(cfg, modelID, providerName)
	}

	return findModelInAllProviders(cfg, modelID)
}

func findModelInNamedProvider(cfg *GlobalConfig, modelID string, providerName string) (*ModelProvider, *Model, error) {
	for i := range cfg.ModelProviders {
		if strings.EqualFold(cfg.ModelProviders[i].Name, providerName) {
			provider := &cfg.ModelProviders[i]
			model := FindModelInProvider(provider, modelID)
			if model == nil {
				return nil, nil, fmt.Errorf("model '%s' not found in provider '%s'", modelID, providerName)
			}
			return provider, model, nil
		}
	}
	return nil, nil, fmt.Errorf("provider '%s' not found", providerName)
}

func findModelInAllProviders(cfg *GlobalConfig, modelID string) (*ModelProvider, *Model, error) {
	var matchingProviders []*ModelProvider
	var matchingModels []*Model

	for i := range cfg.ModelProviders {
		provider := &cfg.ModelProviders[i]
		model := FindModelInProvider(provider, modelID)
		if model != nil {
			matchingProviders = append(matchingProviders, provider)
			matchingModels = append(matchingModels, model)
		}
	}

	if len(matchingProviders) == 0 {
		return nil, nil, fmt.Errorf("model '%s' not found in any provider", modelID)
	}

	if len(matchingProviders) > 1 {
		providerNames := make([]string, len(matchingProviders))
		for i := range matchingProviders {
			providerNames[i] = matchingProviders[i].Name
		}
		return nil, nil, fmt.Errorf("model '%s' found in multiple providers: %s", modelID, strings.Join(providerNames, ", "))
	}

	return matchingProviders[0], matchingModels[0], nil
}
