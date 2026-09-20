package definitions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestTransformExactContentLimits(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "site.yaml", valid+`  transforms:
    - target: root
      element.rename: {tag: article}
`)
	snapshot, err := LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	program := snapshot.Match("example.com").Content.OrderedTransforms
	for _, dimension := range []string{"nodes", "depth", "work"} {
		for _, delta := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("%s/%+d", dimension, delta), func(t *testing.T) {
				root := &html.Node{Type: html.ElementNode, Data: "article", DataAtom: atom.Article}
				switch dimension {
				case "nodes":
					// The selected root and the executor's container count too.
					for i := 0; i < MaxContentNodes+delta-2; i++ {
						root.AppendChild(&html.Node{Type: html.ElementNode, Data: "span", DataAtom: atom.Span})
					}
				case "depth":
					node := root
					for i := 1; i < MaxContentDepth+delta; i++ {
						child := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
						node.AppendChild(child)
						node = child
					}
				case "work":
					// Initial nodes/data, clone, step walk/match, and rename bytes.
					overhead := (1 + len("article")) + 1 + 1 + 3 + 1 + len("article")
					root.AppendChild(&html.Node{Type: html.TextNode, Data: strings.Repeat("x", MaxTransformWork+delta-overhead)})
				}
				source := goquery.NewDocumentFromNode(root).Selection
				output, err := program.CopyAndExecute(context.Background(), source, "https://example.com/article")
				if delta <= 0 {
					if err != nil || output == nil {
						t.Fatalf("valid %s boundary rejected: %v", dimension, err)
					}
					if output.Find("article").Length() != 1 {
						t.Fatal("valid boundary did not execute the transform")
					}
				} else {
					var operationError *OperationError
					if output != nil || !errors.Is(err, ErrTransformLimit) || !errors.As(err, &operationError) || operationError.Step != 1 {
						t.Fatalf("excess %s did not return contextual limit failure: %v", dimension, err)
					}
				}
				if root.Parent != nil || root.Data != "article" {
					t.Fatal("executor mutated the source")
				}
			})
		}
	}
}
