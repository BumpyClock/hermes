package definitionbundle

import (
	"fmt"
	"net/url"
	"path/filepath"
	"slices"

	"github.com/BumpyClock/hermes/internal/definitions"
)

// AuditRequirements checks a snapshot that the real loader already accepted.
// The loader records capability use while validating, so this is an inventory
// check against the manifest, not an alternative YAML interpreter.
// The snapshot must be loaded from the directory that WriteDefinitions produced for the bundle.
func (m *Manifest) AuditRequirements(suite *Suite, snapshot *definitions.Snapshot) error {
	capabilities, algorithms := snapshot.UsedCapabilities()
	for _, capability := range capabilities {
		if !slices.Contains(m.Engine.Operations, capability) {
			return fmt.Errorf("missing required operation declaration %q", capability)
		}
	}
	for _, name := range algorithms {
		if !slices.Contains(m.Engine.Algorithms, name) {
			return fmt.Errorf("missing required named algorithm declaration %q", name)
		}
	}
	// WriteDefinitions flattens definitions/<name>.yaml to <name>.yaml.
	sites := snapshot.Sites()
	for _, c := range suite.Cases {
		u, _ := url.Parse(c.URL)
		selected := snapshot.Match(u.Hostname())
		if selected == nil || selected.Domain != sites[filepath.Base(c.Definition)] {
			return fmt.Errorf("case %s: URL does not select declared definition %q", c.ID, c.Definition)
		}
	}
	return nil
}
