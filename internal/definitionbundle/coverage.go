package definitionbundle

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Coverage binds complete-release claims to separately pinned inventory and approval evidence.
// Neither bundle metadata nor a passing synthetic fixture constitutes cohort approval.
type Coverage struct {
	Protocol        int            `json:"protocol"`
	InventorySHA256 string         `json:"inventory_sha256"`
	EvidenceSHA256  string         `json:"evidence_sha256"`
	PayloadSHA256   string         `json:"payload_sha256"`
	Sites           []CoverageSite `json:"sites"`
}

// CoverageSite maps one frozen registry identity to a migrated definition and real cases.
type CoverageSite struct {
	ID         string   `json:"id"`
	Cohort     string   `json:"cohort"`
	Definition string   `json:"definition"`
	Cases      []string `json:"cases"`
}

// ReleaseGate is caller-supplied trusted, pinned input, never selected by the bundle.
type ReleaseGate struct {
	InventorySHA256 string            `json:"inventory_sha256"`
	EvidenceSHA256  string            `json:"evidence_sha256"`
	PayloadSHA256   string            `json:"payload_sha256"`
	Approved        bool              `json:"approved"`
	Members         map[string]string `json:"members"`
	Sites           []CoverageSite    `json:"sites"`
}

// CheckCoverage prevents a pilot from becoming a production certification.
// Complete bundles require exact membership/cohort equality with the approved gate.
func (m *Manifest) CheckCoverage(files map[string][]byte, suite *Suite, gate *ReleaseGate) error {
	if m.Scope == "pilot" {
		if gate != nil || files["coverage.json"] != nil {
			return fmt.Errorf("pilot is incomplete and cannot pass a production gate")
		}
		return nil
	}
	if gate == nil || !gate.Approved || !digestPattern.MatchString(gate.InventorySHA256) ||
		!digestPattern.MatchString(gate.EvidenceSHA256) || !digestPattern.MatchString(gate.PayloadSHA256) ||
		len(gate.Members) != 125 || len(gate.Sites) != 125 {
		return fmt.Errorf("complete bundle requires approved pinned inventory and cohort evidence for all 125 active identities")
	}
	var coverage Coverage
	if err := DecodeJSON(files["coverage.json"], &coverage); err != nil {
		return fmt.Errorf("complete coverage: %w", err)
	}
	if coverage.Protocol != Protocol || coverage.InventorySHA256 != gate.InventorySHA256 ||
		coverage.EvidenceSHA256 != gate.EvidenceSHA256 || coverage.PayloadSHA256 != gate.PayloadSHA256 ||
		len(coverage.Sites) != len(gate.Members) {
		return fmt.Errorf("complete coverage does not match pinned inventory/evidence")
	}
	records := make([]File, 0, len(m.Files))
	for _, file := range m.Files {
		if file.Path != "coverage.json" {
			records = append(records, file)
		}
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		return err
	}
	if Digest(append(encoded, '\n')) != gate.PayloadSHA256 {
		return fmt.Errorf("complete payload differs from approved cohort evidence")
	}
	for i, site := range coverage.Sites {
		approved := gate.Sites[i]
		if site.ID != approved.ID || site.Cohort != approved.Cohort || site.Definition != approved.Definition ||
			!slices.Equal(site.Cases, approved.Cases) || (i > 0 && site.ID <= coverage.Sites[i-1].ID) {
			return fmt.Errorf("complete coverage differs from approved evidence for %q", site.ID)
		}
	}
	cases := map[string]Case{}
	for _, c := range suite.Cases {
		cases[c.ID] = c
	}
	seen, definitions := map[string]bool{}, map[string]bool{}
	for _, site := range coverage.Sites {
		cohort, exists := gate.Members[site.ID]
		if !exists || seen[site.ID] || site.Cohort != cohort || site.Cohort == "" || len(site.Cases) == 0 ||
			definitions[site.Definition] || files[site.Definition] == nil || !slices.IsSorted(site.Cases) {
			return fmt.Errorf("complete coverage: invalid membership/cohort/definition for %q", site.ID)
		}
		for i, id := range site.Cases {
			c, exists := cases[id]
			if !exists || *c.Synthetic || c.Definition != site.Definition || (i > 0 && id == site.Cases[i-1]) {
				return fmt.Errorf("complete coverage: %s requires unique non-synthetic passing cases", site.ID)
			}
		}
		seen[site.ID], definitions[site.Definition] = true, true
	}
	for _, c := range suite.Cases {
		if !definitions[c.Definition] {
			return fmt.Errorf("definition %s is outside complete coverage", c.Definition)
		}
	}
	return nil
}
