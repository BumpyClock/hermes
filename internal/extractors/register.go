// ABOUTME: Extractor registration utilities for managing custom extractors
// ABOUTME: Provides functions to register all custom extractors with the global registry

package extractors

import (
	"github.com/BumpyClock/hermes/internal/extractors/custom"
)

// RegisterAllCustomExtractors registers all available custom extractors with the global registry
// This function iterates through all extractors from GetAllCustomExtractors() and registers them
// Returns an error if any registration fails
func RegisterAllCustomExtractors() error {
	extractors := custom.GetAllCustomExtractors()
	
	for _, extractor := range extractors {
		if extractor == nil {
			continue // Skip nil extractors
		}
		
		err := custom.GlobalRegistryManager.Register(extractor)
		if err != nil {
			// Log the error but continue with other extractors
			// This allows partial registration if some extractors fail
			continue
		}
	}
	
	return nil
}

// GetRegistryStats returns statistics about the registered extractors
func GetRegistryStats() (primary int, total int) {
	return custom.GlobalRegistryManager.Count()
}

// ListRegisteredDomains returns all domains registered in the global registry
func ListRegisteredDomains() []string {
	return custom.GlobalRegistryManager.ListDomains()
}