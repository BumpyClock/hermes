package definitions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type descendantAttributeSource struct {
	selection selectorTarget
	input     attributeInput
}

func (s descendantAttributeSource) read(e *execution, node *html.Node) (string, bool, error) {
	nodes, err := selectedDescendants(e, node, s.selection)
	if err != nil {
		return "", false, err
	}
	if len(nodes) == 0 {
		if s.input.required {
			return "", false, fmt.Errorf("required descendant attribute %q is absent", s.input.name)
		}
		return "", false, nil
	}
	return s.input.read(e, nodes[0])
}

type jsonPathSegment struct {
	field   string
	index   int
	isIndex bool
}

type jsonValueSource struct {
	input attributeInput
	path  []jsonPathSegment
}

func (s jsonValueSource) read(e *execution, node *html.Node) (string, bool, error) {
	raw, exists, err := s.input.read(e, node)
	if err != nil || !exists {
		return "", false, err
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return s.invalid(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("contains multiple values")
		}
		return s.invalid(err)
	}
	if depth := jsonDepth(value); depth > MaxJSONDepth {
		return s.invalid(fmt.Errorf("%w: JSON exceeds depth %d", ErrTransformLimit, MaxJSONDepth))
	}
	for _, segment := range s.path {
		if err := e.spend(1); err != nil {
			return "", false, err
		}
		if segment.isIndex {
			array, ok := value.([]any)
			if !ok {
				return s.invalid(fmt.Errorf("JSON path index %d requires an array", segment.index))
			}
			if segment.index >= len(array) {
				return s.invalid(fmt.Errorf("JSON path index %d is out of bounds", segment.index))
			}
			value = array[segment.index]
			continue
		}
		object, ok := value.(map[string]any)
		if !ok {
			return s.invalid(fmt.Errorf("JSON path field %q requires an object", segment.field))
		}
		var found bool
		value, found = object[segment.field]
		if !found {
			return s.invalid(fmt.Errorf("JSON path field %q is absent", segment.field))
		}
	}
	text, ok := value.(string)
	if !ok {
		return s.invalid(fmt.Errorf("JSON path result must be a string"))
	}
	if err := e.value(text); err != nil {
		return "", false, err
	}
	return text, true, nil
}

func (s jsonValueSource) invalid(err error) (string, bool, error) {
	if errors.Is(err, ErrTransformLimit) {
		return "", false, err
	}
	if s.input.required {
		return "", false, fmt.Errorf("required JSON value: %w", err)
	}
	return "", false, nil
}

func jsonDepth(value any) int {
	switch current := value.(type) {
	case map[string]any:
		depth := 1
		for _, child := range current {
			if childDepth := 1 + jsonDepth(child); childDepth > depth {
				depth = childDepth
			}
		}
		return depth
	case []any:
		depth := 1
		for _, child := range current {
			if childDepth := 1 + jsonDepth(child); childDepth > depth {
				depth = childDepth
			}
		}
		return depth
	default:
		return 0
	}
}

type namedAlgorithm struct{ name string }

func (o namedAlgorithm) apply(e *execution, node *html.Node) error {
	switch o.name {
	case "abendblatt.deobfuscate":
		return deobfuscateAbendblatt(e, node)
	default:
		return fmt.Errorf("unsupported named algorithm %q", o.name)
	}
}

func deobfuscateAbendblatt(e *execution, node *html.Node) error {
	if !hasClass(node, "obfuscated") {
		return nil
	}
	var text strings.Builder
	if err := e.walk(node, func(current *html.Node, _ int) error {
		if current.Type != html.TextNode {
			return nil
		}
		if text.Len()+len(current.Data) > MaxValueBytes {
			return fmt.Errorf("%w: decoded text exceeds %d bytes", ErrTransformLimit, MaxValueBytes)
		}
		if err := e.value(current.Data); err != nil {
			return err
		}
		text.WriteString(current.Data)
		return nil
	}); err != nil {
		return err
	}
	var output strings.Builder
	for _, character := range text.String() {
		switch character {
		case 177:
			output.WriteByte('%')
		case 178:
			output.WriteByte('!')
		case 180:
			output.WriteByte(';')
		case 181:
			output.WriteByte('=')
		case ' ', '\n':
			output.WriteRune(character)
		default:
			if character > 33 {
				output.WriteRune(character - 1)
			}
		}
	}
	if err := e.value(output.String()); err != nil {
		return err
	}
	for node.FirstChild != nil {
		node.RemoveChild(node.FirstChild)
	}
	node.AppendChild(&html.Node{Type: html.TextNode, Data: output.String()})
	setClass(node, "obfuscated", false)
	setClass(node, "deobfuscated", true)
	return nil
}

func hasClass(node *html.Node, class string) bool {
	for _, attribute := range node.Attr {
		if attribute.Namespace != "" || attribute.Key != "class" {
			continue
		}
		for _, value := range strings.Fields(attribute.Val) {
			if value == class {
				return true
			}
		}
	}
	return false
}

func setClass(node *html.Node, class string, include bool) {
	for i := range node.Attr {
		if node.Attr[i].Namespace != "" || node.Attr[i].Key != "class" {
			continue
		}
		classes := make([]string, 0, len(strings.Fields(node.Attr[i].Val))+1)
		for _, value := range strings.Fields(node.Attr[i].Val) {
			if value != class {
				classes = append(classes, value)
			}
		}
		if include {
			classes = append(classes, class)
		}
		node.Attr[i].Val = strings.Join(classes, " ")
		return
	}
	if include {
		node.Attr = append(node.Attr, html.Attribute{Key: "class", Val: class})
	}
}
