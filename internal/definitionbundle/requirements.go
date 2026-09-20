package definitionbundle

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/BumpyClock/hermes/internal/definitions"
)

// AuditRequirements runs only after the real loader accepts the definitions.
// This is a capability-use inventory, not an alternative YAML interpreter.
func (m *Manifest) AuditRequirements(files map[string][]byte, suite *Suite, directory string) error {
	matcher, err := definitions.LoadDirectory(directory)
	if err != nil {
		return err
	}
	used := map[string]bool{"content.groups": true, "content.default_cleaner": true}
	algorithms := map[string]bool{}
	sites := map[string]string{}
	for _, file := range m.Files {
		if !strings.HasPrefix(file.Path, "definitions/") {
			continue
		}
		var root yaml.Node
		if err := yaml.Unmarshal(files[file.Path], &root); err != nil {
			return err
		}
		doc := mapping(root.Content[0])
		sites[file.Path] = doc["site"].Value
		for _, host := range doc["hosts"].Content {
			if strings.HasPrefix(host.Value, "*.") {
				used["hosts.wildcard"] = true
			} else {
				used["hosts.exact-www"] = true
			}
		}
		if metadata := doc["metadata"]; metadata != nil {
			for _, field := range mapping(metadata) {
				for _, alternative := range field.Content {
					for name := range mapping(alternative) {
						switch name {
						case "text", "attribute", "text_capture":
							used["metadata."+name] = true
						default:
							return fmt.Errorf("%s: bundle capability auditor does not support metadata %q", file.Path, name)
						}
					}
				}
			}
		}
		for key := range mapping(doc["content"]) {
			switch key {
			case "groups", "default_cleaner":
			case "remove":
				used["content.remove"] = true
			case "preserve":
				used["content.preserve"] = true
			case "transforms":
				for _, step := range mapping(doc["content"])["transforms"].Content {
					for key, value := range mapping(step) {
						switch key {
						case "target", "selector":
						case "when":
							for _, condition := range value.Content {
								for kind := range mapping(condition) {
									used["condition."+kind] = true
								}
							}
						default:
							used["transform."+key] = true
							if key == "algorithm.apply" {
								algorithms[mapping(value)["name"].Value] = true
							}
							auditOperationFeatures(key, value, used)
						}
					}
				}
			default:
				return fmt.Errorf("%s: bundle capability auditor does not support content %q", file.Path, key)
			}
		}
	}
	capabilities := make([]string, 0, len(used))
	for capability := range used {
		capabilities = append(capabilities, capability)
	}
	slices.Sort(capabilities)
	for _, capability := range capabilities {
		if !slices.Contains(m.Engine.Operations, capability) {
			return fmt.Errorf("missing required operation declaration %q", capability)
		}
	}
	names := make([]string, 0, len(algorithms))
	for name := range algorithms {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if !slices.Contains(m.Engine.Algorithms, name) {
			return fmt.Errorf("missing required named algorithm declaration %q", name)
		}
	}
	for _, c := range suite.Cases {
		u, _ := url.Parse(c.URL)
		selected := matcher.Match(u.Hostname())
		if selected == nil || selected.Domain != sites[c.Definition] {
			return fmt.Errorf("case %s: URL does not select declared definition %q", c.ID, c.Definition)
		}
	}
	return nil
}

func auditOperationFeatures(name string, parameters *yaml.Node, used map[string]bool) {
	values := mapping(parameters)
	switch name {
	case "url.resolve":
		if base := values["base"]; base != nil {
			used["transform.url.resolve.base"] = true
			auditValueFeatures(base, used)
		}
		if values["to"] != nil {
			used["transform.url.resolve.to"] = true
		}
	case "url.build":
		if path := values["path"]; path != nil {
			for _, value := range path.Content {
				auditValueFeatures(value, used)
			}
		}
		if query := values["query"]; query != nil {
			for _, entry := range query.Content {
				auditValueFeatures(mapping(entry)["value"], used)
			}
		}
	case "attribute.set_from":
		auditValueFeatures(values["value"], used)
	case "noscript.recover":
		if mapping(values["source"])["self"] != nil {
			used["transform.noscript.recover.self"] = true
		}
	case "element.create":
		if mapping(values["target"])["self"] != nil {
			used["element.create.self_target"] = true
		}
		for _, value := range mapping(mapping(values["node"])["attributes"]) {
			if value.Kind == yaml.MappingNode {
				used["element.create.typed_attributes"] = true
				auditValueFeatures(value, used)
			}
		}
	}
}

func auditValueFeatures(value *yaml.Node, used map[string]bool) {
	for kind := range mapping(value) {
		switch kind {
		case "json", "descendant_attribute":
			used["value."+kind] = true
		}
	}
}

func mapping(node *yaml.Node) map[string]*yaml.Node {
	result := map[string]*yaml.Node{}
	if node != nil {
		for i := 0; i+1 < len(node.Content); i += 2 {
			result[node.Content[i].Value] = node.Content[i+1]
		}
	}
	return result
}
