package definitions

import (
	"maps"
	"slices"
)

// UsedCapabilities returns the sorted engine capability and named algorithm
// identifiers that validation recorded for the loaded rules.
// Returned slices are independent and safe for callers to modify.
func (s *Snapshot) UsedCapabilities() (operations, algorithms []string) {
	used, named := map[string]bool{}, map[string]bool{}
	if s != nil {
		for _, r := range s.rules {
			maps.Copy(used, r.capabilities)
			maps.Copy(named, r.algorithms)
		}
	}
	return slices.Sorted(maps.Keys(used)), slices.Sorted(maps.Keys(named))
}

// Sites maps each loaded definition file's base name to its site identifier.
// The returned map is independent and safe for callers to modify.
func (s *Snapshot) Sites() map[string]string {
	sites := map[string]string{}
	if s != nil {
		for _, r := range s.rules {
			sites[r.source] = r.extractor.Domain
		}
	}
	return sites
}
