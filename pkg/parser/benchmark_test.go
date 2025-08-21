package parser_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

// BenchmarkParseHTML tests parsing performance with real HTML fixtures
func BenchmarkParseHTML(b *testing.B) {
	// Load a sample fixture
	fixtureFile := "../../internal/fixtures/www.nytimes.com.html"
	html, err := ioutil.ReadFile(fixtureFile)
	if err != nil {
		b.Skip("Fixture file not available:", err)
	}

	p := parser.New()
	htmlStr := string(html)
	url := "https://www.nytimes.com/test-article"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := p.ParseHTML(htmlStr, url, &parser.ParserOptions{})
		if err != nil {
			b.Fatal(err)
		}
		if result.IsError() {
			b.Fatal(result.Message)
		}
	}
}

// BenchmarkParseHTMLMemory measures memory allocations
func BenchmarkParseHTMLMemory(b *testing.B) {
	fixtureFile := "../../internal/fixtures/www.nytimes.com.html"
	html, err := ioutil.ReadFile(fixtureFile)
	if err != nil {
		b.Skip("Fixture file not available:", err)
	}

	p := parser.New()
	htmlStr := string(html)
	url := "https://www.nytimes.com/test-article"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result, err := p.ParseHTML(htmlStr, url, &parser.ParserOptions{})
		if err != nil {
			b.Fatal(err)
		}
		if result.IsError() {
			b.Fatal(result.Message)
		}
	}
}

// BenchmarkParseMultipleFixtures tests with various site fixtures
func BenchmarkParseMultipleFixtures(b *testing.B) {
	fixtures := []string{
		"www.nytimes.com.html",
		"www.washingtonpost.com.html", 
		"www.cnn.com.html",
		"medium.com.html",
		"arstechnica.com.html",
	}

	p := parser.New()
	
	for _, fixture := range fixtures {
		b.Run(fixture, func(b *testing.B) {
			fixtureFile := filepath.Join("../../internal/fixtures", fixture)
			html, err := ioutil.ReadFile(fixtureFile)
			if err != nil {
				b.Skip("Fixture not available:", fixture)
				return
			}

			htmlStr := string(html)
			url := fmt.Sprintf("https://%s/test-article", fixture[:len(fixture)-5]) // remove .html

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := p.ParseHTML(htmlStr, url, &parser.ParserOptions{})
				if err != nil {
					b.Fatal(err)
				}
				if result.IsError() {
					b.Fatal(result.Message)
				}
			}
		})
	}
}

// BenchmarkDifferentContentTypes tests output format performance
func BenchmarkDifferentContentTypes(b *testing.B) {
	fixtureFile := "../../internal/fixtures/www.nytimes.com.html"
	html, err := ioutil.ReadFile(fixtureFile)
	if err != nil {
		b.Skip("Fixture file not available:", err)
	}

	htmlStr := string(html)
	url := "https://www.nytimes.com/test-article"

	contentTypes := []string{"html", "markdown", "text"}
	
	for _, contentType := range contentTypes {
		b.Run(contentType, func(b *testing.B) {
			p := parser.New()
			opts := parser.ParserOptions{
				ContentType: contentType,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := p.ParseHTML(htmlStr, url, &opts)
				if err != nil {
					b.Fatal(err)
				}
				if result.IsError() {
					b.Fatal(result.Message)
				}
			}
		})
	}
}