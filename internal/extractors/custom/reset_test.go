package custom

import "sync"

// ResetExtractorsForTest resets the package-level extractor caches for testing
// This function should only be used in tests to ensure clean state between test runs
//
// WARNING: NOT safe for concurrent use. Do not call from tests using t.Parallel().
// This function modifies package-level state without synchronization.
//
// TODO: If tests begin calling this helper in parallel, add a mutex or sync.Once
// reinitialization mechanism to ensure thread-safe reset operations.
func ResetExtractorsForTest() {
	allExtractors = nil
	domainToExtractor = nil
	extractorOnce = sync.Once{}
}
