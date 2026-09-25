package definitionbundle

import (
	"fmt"
	"net/url"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"

	"github.com/BumpyClock/hermes/internal/definitions"
)

// AuditRequirements checks a snapshot that the real loader already accepted.
// The loader records capability use while validating, so this is an inventory
// check against the manifest, not an alternative YAML interpreter.
// The snapshot must be loaded from the bundle's DefinitionFiles, either in
// memory with definitions.LoadFiles or from the directory that WriteDefinitions
// produced.
func (m *Manifest) AuditRequirements(suite *Suite, snapshot *definitions.Snapshot) error {
	if err := m.CheckDeclaredCapabilities(snapshot.UsedCapabilities()); err != nil {
		return err
	}
	// DefinitionFiles flattens definitions/<name>.yaml to <name>.yaml.
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

// CheckDeclaredCapabilities requires the manifest to declare every capability
// and named algorithm that the loaded definitions use.
func (m *Manifest) CheckDeclaredCapabilities(capabilities, algorithms []string) error {
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
	return nil
}

// DefinitionSite reads the site identifier of one definition file.
// Call it only for files that the real loader already accepted.
func DefinitionSite(data []byte) (string, error) {
	var header struct {
		Site string `yaml:"site"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return "", err
	}
	if header.Site == "" {
		return "", fmt.Errorf("definition has no site")
	}
	return header.Site, nil
}
