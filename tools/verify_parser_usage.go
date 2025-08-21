package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BumpyClock/hermes/pkg/extractors/custom"
	"github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_parser_usage.go <url>")
		fmt.Println("This tool verifies custom extractor usage by parsing a URL")
		os.Exit(1)
	}

	targetURL := os.Args[1]
	
	// Parse URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 PARSING URL: %s\n", targetURL)
	fmt.Printf("📍 DOMAIN: %s\n", parsedURL.Host)
	fmt.Printf("\n")

	// Check if custom extractor exists
	fmt.Println("=== CUSTOM EXTRACTOR CHECK ===")
	allExtractors := custom.GetAllCustomExtractors()
	var foundExtractor *custom.CustomExtractor
	
	for _, extractor := range allExtractors {
		if extractor.Domain == parsedURL.Host {
			foundExtractor = extractor
			break
		}
	}
	
	if foundExtractor != nil {
		fmt.Printf("✅ Custom extractor found for domain: %s\n", parsedURL.Host)
		fmt.Printf("   📋 Title selectors: %v\n", foundExtractor.Title.Selectors)
		fmt.Printf("   👤 Author selectors: %v\n", foundExtractor.Author.Selectors)
		fmt.Printf("   📄 Content selectors: %v\n", foundExtractor.Content.Selectors)
	} else {
		fmt.Printf("❌ No custom extractor found for domain: %s\n", parsedURL.Host)
		fmt.Printf("   ℹ️  Will use generic extractors\n")
	}

	// Test the parser
	fmt.Printf("\n=== PARSER EXECUTION ===")
	p := parser.New()
	result, err := p.Parse(targetURL, &parser.ParserOptions{
		ContentType: "html",
		Fallback:    true,
	})

	if err != nil {
		fmt.Printf("❌ Parser error: %v\n", err)
		os.Exit(1)
	}

	if result.IsError() {
		fmt.Printf("❌ Parser returned error: %s\n", result.Message)
		os.Exit(1)
	}

	// Display results
	fmt.Printf("✅ Parser succeeded\n")
	fmt.Printf("\n=== EXTRACTION RESULTS ===")
	fmt.Printf("🔧 Extractor used: %s\n", result.ExtractorUsed)
	fmt.Printf("📰 Title: %s\n", truncateString(result.Title, 80))
	fmt.Printf("👤 Author: %s\n", result.Author)
	
	if result.DatePublished != nil {
		fmt.Printf("📅 Date: %s\n", result.DatePublished.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("📅 Date: (not found)\n")
	}
	
	fmt.Printf("📊 Word count: %d\n", result.WordCount)
	fmt.Printf("📝 Excerpt: %s\n", truncateString(result.Excerpt, 120))
	
	// Show content preview
	contentPreview := strings.ReplaceAll(result.Content, "\n", " ")
	contentPreview = strings.ReplaceAll(contentPreview, "\t", " ")
	for strings.Contains(contentPreview, "  ") {
		contentPreview = strings.ReplaceAll(contentPreview, "  ", " ")
	}
	fmt.Printf("📄 Content preview: %s\n", truncateString(contentPreview, 150))

	// Custom extractor quality check
	fmt.Printf("\n=== QUALITY ASSESSMENT ===")
	if foundExtractor != nil && result.ExtractorUsed == "custom:"+parsedURL.Host {
		fmt.Printf("✅ Custom extractor was used successfully\n")
		
		if result.Title != "" {
			fmt.Printf("✅ Title extracted via custom selectors\n")
		} else {
			fmt.Printf("⚠️  Title not found (check custom selectors)\n")
		}
		
		if result.Author != "" {
			fmt.Printf("✅ Author extracted via custom selectors\n")
		} else {
			fmt.Printf("⚠️  Author not found (check custom selectors)\n")
		}
		
		if result.Content != "" && result.WordCount > 100 {
			fmt.Printf("✅ Content extracted successfully (%d words)\n", result.WordCount)
		} else {
			fmt.Printf("⚠️  Content extraction may need improvement\n")
		}
	} else if foundExtractor != nil {
		fmt.Printf("⚠️  Custom extractor exists but was not used\n")
		fmt.Printf("   Expected: custom:%s\n", parsedURL.Host)
		fmt.Printf("   Actual: %s\n", result.ExtractorUsed)
	} else {
		fmt.Printf("ℹ️  Using generic extraction (no custom extractor available)\n")
	}

	// JSON output option
	if len(os.Args) > 2 && os.Args[2] == "--json" {
		fmt.Printf("\n=== JSON OUTPUT ===\n")
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonData))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}