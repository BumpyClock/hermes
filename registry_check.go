package main

import (
	"fmt"
	"github.com/postlight/parser-go/pkg/extractors"
	"github.com/postlight/parser-go/pkg/extractors/custom"
)

func main() {
	// Initialize the registry
	err := extractors.RegisterAllCustomExtractors()
	if err != nil {
		fmt.Printf("Error initializing extractors: %v\n", err)
		return
	}
	
	// Get counts
	primary, total := custom.GlobalRegistryManager.Count()
	fmt.Printf("Primary extractors: %d\n", primary)
	fmt.Printf("Total domain mappings: %d\n", total)
	
	// List all domains
	domains := custom.GlobalRegistryManager.ListDomains()
	fmt.Printf("All registered domains: %d\n", len(domains))
	
	for i, domain := range domains {
		if i < 10 { // Show first 10
			fmt.Printf("  %s\n", domain)
		}
	}
	
	if len(domains) > 10 {
		fmt.Printf("  ... and %d more\n", len(domains) - 10)
	}
}