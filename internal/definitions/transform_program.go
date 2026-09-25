package definitions

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	MaxTransformSteps = 128
	MaxConditions     = 16
	MaxPatternBytes   = 512
	MaxCaptureGroups  = 16
	MaxValueBytes     = 64 * 1024
	MaxContentNodes   = 32768
	MaxContentDepth   = 256
	MaxTransformWork  = 2 * 1024 * 1024
	MaxJSONDepth      = 16
	MaxJSONTraversal  = 32
)

// ErrTransformLimit identifies an exceeded content, value, or execution budget.
var ErrTransformLimit = errors.New("definition transform resource limit exceeded")

// OperationError preserves the definition, one-based step, and underlying failure.
type OperationError struct {
	Source, Site, Operation string
	Step, Line, Column      int
	Err                     error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("definition %s:%d:%d site %q transform step %d (%s): %v",
		e.Source, e.Line, e.Column, e.Site, e.Step, e.Operation, e.Err)
}

func (e *OperationError) Unwrap() error { return e.Err }

type operation interface {
	apply(*execution, *html.Node) error
}

type condition interface {
	matches(*execution, *html.Node) (bool, error)
}

type transformStep struct {
	root       bool
	selector   cascadia.Selector
	conditions []condition
	operation  operation
	name       string
	line, col  int
}

type program struct {
	source, site string
	steps        []transformStep
}

type execution struct {
	ctx       context.Context
	base      string
	remaining int

	// Lazily parsed URL bases shared by url.resolve matches. Callers must not
	// mutate them.
	articleBase, defaultBase       *url.URL
	articleBaseErr, defaultBaseErr error
	articleParsed, defaultParsed   bool
}

func (e *execution) sourceBase() (*url.URL, error) {
	if !e.articleParsed {
		e.articleBase, e.articleBaseErr = httpURL(e.base)
		e.articleParsed = true
	}
	return e.articleBase, e.articleBaseErr
}

// defaultResolutionBase caches the resolution base used when url.resolve has
// no explicit base: the article base resolved against itself.
func (e *execution) defaultResolutionBase() (*url.URL, error) {
	if !e.defaultParsed {
		e.defaultBase, e.defaultBaseErr = e.resolutionBase(e.base)
		e.defaultParsed = true
	}
	return e.defaultBase, e.defaultBaseErr
}

func (e *execution) resolutionBase(raw string) (*url.URL, error) {
	sourceBase, err := e.sourceBase()
	if err != nil {
		return nil, fmt.Errorf("invalid article base: %w", err)
	}
	reference, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	base, err := httpURL(sourceBase.ResolveReference(reference).String())
	if err != nil {
		return nil, fmt.Errorf("invalid URL resolution base: %w", err)
	}
	return base, nil
}

func (e *execution) spend(n int) error {
	if err := e.ctx.Err(); err != nil {
		return err
	}
	if n > e.remaining {
		return fmt.Errorf("%w: work exceeds %d units", ErrTransformLimit, MaxTransformWork)
	}
	e.remaining -= n
	return nil
}

func (e *execution) value(s string) error {
	if err := checkValueSize(s); err != nil {
		return err
	}
	return e.spend(len(s))
}

func checkValueSize(s string) error {
	if len(s) > MaxValueBytes {
		return fmt.Errorf("%w: value exceeds %d bytes", ErrTransformLimit, MaxValueBytes)
	}
	return nil
}

type nodeMatch struct {
	node  *html.Node
	depth int
}

// walk visits a bounded tree in document order and checks cancellation between nodes.
func (e *execution) walk(root *html.Node, visit func(*html.Node, int) error) error {
	stack := []nodeMatch{{root, 0}}
	count := 0
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		count++
		if count > MaxContentNodes || current.depth > MaxContentDepth {
			return fmt.Errorf("%w: content exceeds %d nodes or depth %d", ErrTransformLimit, MaxContentNodes, MaxContentDepth)
		}
		if err := e.spend(1); err != nil {
			return err
		}
		if err := visit(current.node, current.depth); err != nil {
			return err
		}
		for child := current.node.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, nodeMatch{child, current.depth + 1})
			if len(stack) > MaxContentNodes {
				return fmt.Errorf("%w: too many content nodes", ErrTransformLimit)
			}
		}
	}
	return nil
}

func (p *program) CopyAndExecute(ctx context.Context, elements *goquery.Selection, base string) (*goquery.Selection, error) {
	e := &execution{ctx: ctx, base: base, remaining: MaxTransformWork}
	count := 1
	for _, node := range elements.Nodes {
		err := e.walk(node, func(n *html.Node, depth int) error {
			count++
			if count > MaxContentNodes || depth+1 > MaxContentDepth {
				return fmt.Errorf("%w: content exceeds %d nodes or depth %d", ErrTransformLimit, MaxContentNodes, MaxContentDepth)
			}
			if err := e.spend(len(n.Data)); err != nil {
				return err
			}
			for _, attr := range n.Attr {
				if err := e.spend(len(attr.Key) + len(attr.Val) + 1); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, p.failure(0, err)
		}
	}
	root := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	wrapper := goquery.NewDocumentFromNode(root).Selection
	for _, node := range elements.Nodes {
		if err := e.spend(1); err != nil {
			return nil, p.failure(0, err)
		}
		wrapper.AppendSelection(goquery.NewDocumentFromNode(node).Clone())
	}
	roots := map[*html.Node]bool{}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		roots[child] = true
	}
	for index, step := range p.steps {
		err := p.executeStep(e, root, roots, step)
		if err != nil {
			return nil, p.failure(index, err)
		}
		if requiresTreeValidation(step.operation) {
			if err := e.validateTree(root); err != nil {
				return nil, p.failure(index, err)
			}
		}
	}
	return wrapper, nil
}

// requiresTreeValidation is false only for operations that cannot add, move, or
// deepen nodes. Their bounded reads and writes are charged at the operation.
func requiresTreeValidation(op operation) bool {
	switch op.(type) {
	case attributeCopy, attributeSet, attributeSetFrom, attributeRemove, stringReplace,
		regexCapture, urlResolve, urlBuild, elementRename:
		return false
	default:
		return true
	}
}

func (e *execution) validateTree(root *html.Node) error {
	return e.walk(root, func(node *html.Node, _ int) error {
		if err := e.spend(len(node.Data)); err != nil {
			return err
		}
		for _, attribute := range node.Attr {
			if err := e.spend(len(attribute.Key) + len(attribute.Val) + 1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *program) failure(index int, err error) error {
	step := p.steps[index]
	return &OperationError{Source: p.source, Site: p.site, Step: index + 1,
		Operation: step.name, Line: step.line, Column: step.col, Err: err}
}

func (p *program) executeStep(e *execution, root *html.Node, roots map[*html.Node]bool, step transformStep) error {
	matches := []nodeMatch{}
	err := e.walk(root, func(n *html.Node, depth int) error {
		if n == root || n.Type != html.ElementNode {
			return nil
		}
		if (step.root && roots[n]) || (!step.root && !roots[n] && step.selector.Match(n)) {
			matches = append(matches, nodeMatch{n, depth})
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].depth > matches[j].depth })
	for _, match := range matches {
		if err := e.spend(1); err != nil {
			return err
		}
		if !attachedTo(match.node, root) {
			continue
		}
		run := true
		for _, predicate := range step.conditions {
			ok, err := predicate.matches(e, match.node)
			if err != nil {
				return err
			}
			if !ok {
				run = false
				break
			}
		}
		if run {
			if err := step.operation.apply(e, match.node); err != nil {
				return err
			}
		}
	}
	return nil
}

func attachedTo(n, root *html.Node) bool {
	for n != nil {
		if n == root {
			return true
		}
		n = n.Parent
	}
	return false
}

func httpURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
		return nil, fmt.Errorf("expected absolute HTTP(S) URL without credentials")
	}
	return u, nil
}
