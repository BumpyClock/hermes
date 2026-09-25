package definitions

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type existsCondition struct {
	attribute string
	present   bool
}

func (c existsCondition) matches(e *execution, n *html.Node) (bool, error) {
	if err := e.spend(len(n.Attr)); err != nil {
		return false, err
	}
	for _, attr := range n.Attr {
		if attr.Namespace == "" && attr.Key == c.attribute {
			return c.present, nil
		}
	}
	return !c.present, nil
}

type textCondition struct {
	input             attributeInput
	comparison, value string
}

func (c textCondition) matches(e *execution, n *html.Node) (bool, error) {
	value := ""
	if c.input.name != "" {
		var exists bool
		var err error
		value, exists, err = c.input.read(e, n)
		if err != nil || !exists {
			return false, err
		}
	} else {
		var text strings.Builder
		err := e.walk(n, func(child *html.Node, _ int) error {
			if child.Type == html.TextNode {
				if text.Len()+len(child.Data) > MaxValueBytes {
					return fmt.Errorf("%w: text exceeds %d bytes", ErrTransformLimit, MaxValueBytes)
				}
				if err := e.value(child.Data); err != nil {
					return err
				}
				text.WriteString(child.Data)
			}
			return nil
		})
		if err != nil {
			return false, err
		}
		value = text.String()
	}
	switch c.comparison {
	case "equals":
		return value == c.value, nil
	case "contains":
		return strings.Contains(value, c.value), nil
	case "prefix":
		return strings.HasPrefix(value, c.value), nil
	case "suffix":
		return strings.HasSuffix(value, c.value), nil
	default:
		return false, fmt.Errorf("unsupported text comparison %q", c.comparison)
	}
}

type numberCondition struct {
	input      attributeInput
	comparison string
	value      float64
}

func (c numberCondition) matches(e *execution, n *html.Node) (bool, error) {
	raw, exists, err := c.input.read(e, n)
	if err != nil || !exists {
		return false, err
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return false, fmt.Errorf("numeric attribute %q: %w", c.input.name, err)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return false, fmt.Errorf("numeric attribute %q must be finite", c.input.name)
	}
	switch c.comparison {
	case "eq":
		return number == c.value, nil
	case "ne":
		return number != c.value, nil
	case "lt":
		return number < c.value, nil
	case "le":
		return number <= c.value, nil
	case "gt":
		return number > c.value, nil
	case "ge":
		return number >= c.value, nil
	default:
		return false, fmt.Errorf("unsupported numeric comparison %q", c.comparison)
	}
}

type descendantCondition struct {
	selector cascadia.Selector
	present  bool
}

// errStopWalk ends a walk early once its visitor has its answer.
var errStopWalk = errors.New("stop walk")

func (c descendantCondition) matches(e *execution, n *html.Node) (bool, error) {
	found := false
	err := e.walk(n, func(child *html.Node, _ int) error {
		if child != n && child.Type == html.ElementNode && c.selector.Match(child) {
			found = true
			return errStopWalk
		}
		return nil
	})
	if errors.Is(err, errStopWalk) {
		err = nil
	}
	return found == c.present, err
}
