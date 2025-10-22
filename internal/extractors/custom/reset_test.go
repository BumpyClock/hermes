//go:build test
// +build test

package custom

import "sync"

// ResetExtractorsForTest resets the package-level extractor caches for testing
// This function should only be used in tests to ensure clean state between test runs
func ResetExtractorsForTest() {
	allExtractors = nil
	domainToExtractor = nil
	extractorOnce = sync.Once{}
}
