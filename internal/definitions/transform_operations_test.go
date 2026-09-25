package definitions

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func idElement(tag, id string, children ...*html.Node) *html.Node {
	n := &html.Node{Type: html.ElementNode, Data: tag, DataAtom: atom.Lookup([]byte(tag)), Attr: []html.Attribute{{Key: "id", Val: id}}}
	for _, child := range children {
		n.AppendChild(child)
	}
	return n
}

func nodeIDs(nodes []*html.Node) string {
	values := make([]string, len(nodes))
	for i, node := range nodes {
		values[i] = node.Attr[0].Val
	}
	return strings.Join(values, ",")
}

func TestOutermostNodesKeepsOutermostSelectionsInDocumentOrder(t *testing.T) {
	root := idElement("article", "root",
		idElement("div", "a", idElement("p", "b", idElement("span", "c"))),
		idElement("p", "d", idElement("span", "e")),
		idElement("span", "f"),
	)
	e := &execution{ctx: context.Background(), remaining: MaxTransformWork}
	selected, err := selectedDescendants(e, root, selectorTarget{selector: cascadia.MustCompile("span, div, p"), mode: "all"})
	if err != nil {
		t.Fatal(err)
	}
	if got := nodeIDs(selected); got != "a,b,c,d,e,f" {
		t.Fatalf("selection is not in document order: %s", got)
	}
	if got := nodeIDs(outermostNodes(selected)); got != "a,d,f" {
		t.Fatalf("outermost nodes = %s, want a,d,f", got)
	}
}

func TestDescendantConditionStopsAtFirstMatch(t *testing.T) {
	root := idElement("article", "root", idElement("img", "match"))
	for range 100 {
		root.AppendChild(idElement("p", "tail"))
	}
	condition := descendantCondition{selector: cascadia.MustCompile("img"), present: true}
	// The root and first child fit the budget; the remaining siblings do not.
	e := &execution{ctx: context.Background(), remaining: 2}
	ok, err := condition.matches(e, root)
	if err != nil || !ok {
		t.Fatalf("matching descendant = %v, %v", ok, err)
	}
	if e.remaining != 0 {
		t.Fatalf("remaining work = %d, want 0", e.remaining)
	}
	absent := descendantCondition{selector: cascadia.MustCompile("img"), present: false}
	e.remaining = 2
	ok, err = absent.matches(e, root)
	if err != nil || ok {
		t.Fatalf("absence test on matching descendant = %v, %v", ok, err)
	}
}

func TestURLBuildDoesNotMutateSharedBase(t *testing.T) {
	base, err := httpURL("https://example.com/api/?v=1")
	if err != nil {
		t.Fatal(err)
	}
	build := urlBuild{
		attribute: "src",
		base:      base,
		path:      []scalarSource{attributeInput{name: "id", required: true}},
		query:     []queryValue{{name: "id", source: attributeInput{name: "id", required: true}}},
	}
	e := &execution{ctx: context.Background(), remaining: MaxTransformWork}
	for _, id := range []string{"a b", "c"} {
		n := idElement("img", id)
		if err := build.apply(e, n); err != nil {
			t.Fatal(err)
		}
		want := "https://example.com/api/" + url.PathEscape(id) + "?id=" + url.QueryEscape(id) + "&v=1"
		if got := n.Attr[1].Val; got != want {
			t.Fatalf("built URL = %q, want %q", got, want)
		}
	}
	if base.String() != "https://example.com/api/?v=1" {
		t.Fatalf("shared base mutated: %s", base)
	}
}

func TestURLResolveReportsInvalidArticleBaseOnEveryMatch(t *testing.T) {
	resolve := urlResolve{input: attributeInput{name: "src"}, to: "src"}
	e := &execution{ctx: context.Background(), base: "file:///article", remaining: MaxTransformWork}
	for range 2 {
		n := &html.Node{Type: html.ElementNode, Data: "img", Attr: []html.Attribute{{Key: "src", Val: "a.jpg"}}}
		if err := resolve.apply(e, n); err == nil || !strings.Contains(err.Error(), "invalid article base") {
			t.Fatalf("invalid article base accepted: %v", err)
		}
	}
}

func BenchmarkURLResolveArticleBase(b *testing.B) {
	resolve := urlResolve{input: attributeInput{name: "src"}, to: "data-src"}
	nodes := make([]*html.Node, 1000)
	for i := range nodes {
		nodes[i] = &html.Node{Type: html.ElementNode, Data: "img", Attr: []html.Attribute{{Key: "src", Val: "../images/" + strconv.Itoa(i) + ".jpg"}}}
	}
	for b.Loop() {
		e := &execution{ctx: context.Background(), base: "https://example.com/news/2026/article.html", remaining: MaxTransformWork}
		for _, n := range nodes {
			if err := resolve.apply(e, n); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// Flat siblings keep the same shape after retain, so each iteration reuses the tree.
func BenchmarkElementRetainFlatSiblings(b *testing.B) {
	for _, count := range []int{2000, 8000, 16000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			root := &html.Node{Type: html.ElementNode, Data: "article", DataAtom: atom.Article}
			for range count {
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "p", DataAtom: atom.P})
			}
			retain := elementRetain{selection: selectorTarget{selector: cascadia.MustCompile("p"), mode: "all"}}
			e := &execution{ctx: context.Background()}
			for b.Loop() {
				e.remaining = MaxTransformWork
				if err := retain.apply(e, root); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestElementMoveRejectsCycles(t *testing.T) {
	for _, tc := range []struct {
		name, source, target, position string
		wantCycle                      bool
	}{
		{"target is source", "#a", "#a", "append", true},
		{"target inside source", "#a", "#b", "append", true},
		{"replace target containing source", "#b", "#a", "replace", true},
		{"append into target containing source", "#b", "#a", "append", false},
		{"replace disjoint target", "#b", "#d", "replace", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := idElement("article", "root", idElement("div", "a", idElement("p", "b")), idElement("p", "d"))
			move := elementMove{
				source:   selectorTarget{selector: cascadia.MustCompile(tc.source), mode: "all"},
				target:   selectorTarget{selector: cascadia.MustCompile(tc.target), mode: "all"},
				position: tc.position,
			}
			e := &execution{ctx: context.Background(), remaining: MaxTransformWork}
			err := move.apply(e, root)
			if tc.wantCycle && (err == nil || !strings.Contains(err.Error(), "cannot be the source")) {
				t.Fatalf("move error = %v, want cycle rejection", err)
			}
			if !tc.wantCycle && err != nil {
				t.Fatalf("move error = %v, want success", err)
			}
		})
	}
}

// Every sources×targets pair is checked for cycles before the generated-node
// limit rejects the move, so each iteration fails on the unchanged tree.
func BenchmarkElementMoveManySourcesAndTargets(b *testing.B) {
	for _, count := range []int{1000, 4000, 16000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			root := &html.Node{Type: html.ElementNode, Data: "article", DataAtom: atom.Article}
			for range count {
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "p", DataAtom: atom.P, Attr: []html.Attribute{{Key: "class", Val: "a"}}})
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "p", DataAtom: atom.P, Attr: []html.Attribute{{Key: "class", Val: "b"}}})
			}
			move := elementMove{
				source:   selectorTarget{selector: cascadia.MustCompile("p.a"), mode: "all"},
				target:   selectorTarget{selector: cascadia.MustCompile("p.b"), mode: "all"},
				position: "append",
			}
			e := &execution{ctx: context.Background()}
			for b.Loop() {
				e.remaining = MaxTransformWork
				if err := move.apply(e, root); !errors.Is(err, ErrTransformLimit) {
					b.Fatalf("move error = %v, want transform limit", err)
				}
			}
		})
	}
}
