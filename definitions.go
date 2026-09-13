package hermes

import "github.com/BumpyClock/hermes/internal/definitions"

// DefinitionError reports invalid local configuration, including its original cause.
type DefinitionError = definitions.Error

// Definitions is an immutable snapshot loaded before client construction.
// Its zero value is an empty, generic-only snapshot.
type Definitions struct{ snapshot *definitions.Snapshot }

// LoadDefinitions validates all immediate YAML files in a directory atomically.
// It performs no network access and never returns a partial snapshot.
func LoadDefinitions(directory string) (*Definitions, error) {
	s, err := definitions.LoadDirectory(directory)
	if err != nil {
		return nil, err
	}
	return &Definitions{snapshot: s}, nil
}

// DefinitionCapabilities describes the implemented local language and its limits.
// Returned slices are independent and safe for callers to modify.
func DefinitionCapabilities() DefinitionSupport {
	return DefinitionSupport{
		Schema:       definitions.SchemaVersion,
		Capabilities: []string{"metadata.text", "metadata.attribute", "content.groups", "content.remove", "content.default_cleaner", "hosts.exact-www", "hosts.wildcard"},
		MaxFiles:     definitions.MaxFiles, MaxFileBytes: definitions.MaxFileBytes,
		MaxTotalBytes: definitions.MaxTotalBytes, MaxNodes: definitions.MaxNodes,
		MaxDepth: definitions.MaxDepth, MaxListItems: definitions.MaxListItems, MaxStringBytes: definitions.MaxStringBytes,
	}
}

// DefinitionSupport is version and numerical validation information for local tooling.
type DefinitionSupport struct {
	Schema                                                                                  int
	Capabilities                                                                            []string
	MaxFiles, MaxFileBytes, MaxTotalBytes, MaxNodes, MaxDepth, MaxListItems, MaxStringBytes int
}

// WithDefinitions isolates a client from the legacy compiled site registry.
// A nil or zero snapshot explicitly selects generic-only parsing.
func WithDefinitions(snapshot *Definitions) Option {
	return func(c *Client) {
		c.definitionsConfigured = true
		if snapshot == nil {
			c.definitions = nil
		} else {
			c.definitions = snapshot.snapshot
		}
	}
}
