package hermes

import "github.com/BumpyClock/hermes/internal/definitions"

// DefinitionError reports invalid local configuration, including its original cause.
type DefinitionError = definitions.Error

// DefinitionOperationError describes a failed article transform and preserves its cause.
type DefinitionOperationError = definitions.OperationError

// ErrDefinitionTransformLimit identifies an exceeded transform resource budget.
var ErrDefinitionTransformLimit = definitions.ErrTransformLimit

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

// UsedCapabilities returns the sorted capability and named algorithm IDs that
// the snapshot's definitions use, in the vocabulary of DefinitionCapabilities.
// A nil or zero snapshot uses none. Returned slices are independent and safe
// for callers to modify.
func (d *Definitions) UsedCapabilities() (capabilities, algorithms []string) {
	if d == nil {
		return nil, nil
	}
	return d.snapshot.UsedCapabilities()
}

// Site returns the site identifier of the definition that host selects.
// It applies the same host matching as parsing and reports false when parsing
// that host would use generic extraction only.
func (d *Definitions) Site(host string) (string, bool) {
	if d == nil || d.snapshot == nil {
		return "", false
	}
	rule := d.snapshot.Match(host)
	if rule == nil {
		return "", false
	}
	return rule.Domain, true
}

// DefinitionCapabilities describes the implemented local language and its limits.
// Returned slices are independent and safe for callers to modify.
func DefinitionCapabilities() DefinitionSupport {
	return DefinitionSupport{
		Schema:       definitions.SchemaVersion,
		Capabilities: append([]string{"metadata.text", "metadata.attribute", "metadata.text_capture", "content.groups", "content.remove", "content.preserve", "content.default_cleaner", "hosts.exact-www", "hosts.wildcard"}, definitions.TransformCapabilities()...),
		Algorithms:   definitions.AlgorithmCapabilities(),
		MaxFiles:     definitions.MaxFiles, MaxFileBytes: definitions.MaxFileBytes,
		MaxTotalBytes: definitions.MaxTotalBytes, MaxNodes: definitions.MaxNodes,
		MaxDepth: definitions.MaxDepth, MaxListItems: definitions.MaxListItems, MaxStringBytes: definitions.MaxStringBytes,
		MaxTransformSteps: definitions.MaxTransformSteps, MaxConditions: definitions.MaxConditions,
		MaxPatternBytes: definitions.MaxPatternBytes, MaxCaptureGroups: definitions.MaxCaptureGroups,
		MaxValueBytes: definitions.MaxValueBytes, MaxContentNodes: definitions.MaxContentNodes,
		MaxContentDepth: definitions.MaxContentDepth, MaxTransformWork: definitions.MaxTransformWork,
		MaxJSONDepth: definitions.MaxJSONDepth, MaxJSONTraversal: definitions.MaxJSONTraversal,
	}
}

// DefinitionSupport is version and numerical validation information for local tooling.
type DefinitionSupport struct {
	Schema                                                                                  int
	Capabilities                                                                            []string
	Algorithms                                                                              []string
	MaxFiles, MaxFileBytes, MaxTotalBytes, MaxNodes, MaxDepth, MaxListItems, MaxStringBytes int
	MaxTransformSteps, MaxConditions, MaxPatternBytes, MaxCaptureGroups, MaxValueBytes      int
	MaxContentNodes, MaxContentDepth, MaxTransformWork                                      int
	MaxJSONDepth, MaxJSONTraversal                                                          int
}

// WithDefinitions selects one immutable external definition snapshot.
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
