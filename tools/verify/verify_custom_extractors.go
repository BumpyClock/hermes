package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BumpyClock/hermes/internal/extractors/custom"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_custom_extractors.go <url>")
		os.Exit(1)
	}

	targetURL := os.Args[1]
	
	// Parse URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Target URL: %s\n", targetURL)
	fmt.Printf("Domain: %s\n", parsedURL.Host)
	fmt.Printf("\n")

	// Check if custom extractor exists for this domain
	fmt.Println("=== CUSTOM EXTRACTOR LOOKUP ===")
	
	// Check via custom package - first by domain lookup
	allExtractors := custom.GetAllCustomExtractors()
	var foundExtractor *custom.CustomExtractor
	
	// Look for exact domain match
	for _, extractor := range allExtractors {
		if extractor.Domain == parsedURL.Host {
			foundExtractor = extractor
			break
		}
	}
	
	if foundExtractor != nil {
		fmt.Printf("✅ Found custom extractor for domain: %s\n", parsedURL.Host)
		fmt.Printf("   Extractor domain: %s\n", foundExtractor.Domain)
		if foundExtractor.Title != nil {
			fmt.Printf("   Title selectors: %v\n", foundExtractor.Title.Selectors)
		}
		if foundExtractor.Author != nil {
			fmt.Printf("   Author selectors: %v\n", foundExtractor.Author.Selectors)
		}
		if foundExtractor.Content != nil {
			fmt.Printf("   Content selectors: %v\n", foundExtractor.Content.Selectors)
		}
	} else {
		fmt.Printf("❌ No custom extractor found for domain: %s\n", parsedURL.Host)
		
		// Try some variations
		fmt.Printf("Trying domain variations...\n")
		withoutWWW := strings.TrimPrefix(parsedURL.Host, "www.")
		withWWW := "www." + strings.TrimPrefix(parsedURL.Host, "www.")
		
		for _, extractor := range allExtractors {
			if extractor.Domain == withoutWWW {
				fmt.Printf("✅ Found custom extractor for base domain: %s\n", withoutWWW)
				foundExtractor = extractor
				break
			}
			if extractor.Domain == withWWW {
				fmt.Printf("✅ Found custom extractor for www domain: %s\n", withWWW)
				foundExtractor = extractor
				break
			}
		}
	}

	fmt.Printf("\n=== REGISTRY STATUS ===")
	fmt.Printf("Total custom extractors registered: %d\n", len(allExtractors))
	
	// List a few examples
	fmt.Printf("Sample registered domains:\n")
	count := 0
	for domain := range allExtractors {
		if count < 10 {
			fmt.Printf("  - %s\n", domain)
		}
		count++
		if count >= 10 {
			break
		}
	}
	if len(allExtractors) > 10 {
		fmt.Printf("  ... and %d more\n", len(allExtractors)-10)
	}
}