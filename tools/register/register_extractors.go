// ABOUTME: Extractor registration utility to populate GlobalRegistryManager
// ABOUTME: Registers all available custom extractors for conversion operations

package main

import (
	"fmt"
	"log"

	"github.com/BumpyClock/hermes/internal/extractors/custom"
)

func main() {
	fmt.Println("🔧 Registering all custom extractors...")
	
	// Register all extractors
	extractors := []*custom.CustomExtractor{
		// International extractors
		custom.GetWwwLemondeFrExtractor(),
		custom.GetWwwYomiuriCoJpExtractor(),
		custom.GetQdailyExtractor(),
		custom.JapanZdnetComExtractor,
		custom.WeeklyAsciiJpExtractor,
		custom.GetWwwMoongiftJpExtractor(),
		custom.GetWwwLifehackerJpExtractor(),
		custom.GetnewsJpExtractor,
		custom.GetJvndbJvnJpExtractor(),
		custom.GetSectIijAdJpExtractor(),
		custom.GetBuzzapJpExtractor(),
		custom.MaTtiasBeExtractor,
		custom.GetWwwGrueneDeExtractor(),
		custom.GetWwwAbendblattDeExtractor(),
		custom.GetWwwProspectmagazineCoUkExtractor(),
		custom.GetWwwCbcCaExtractor(),
		custom.GetIciRadioCanadaCaExtractor(),
		custom.GetTimesofindiaIndiatimesComExtractor(),
		custom.GetWwwNdtvComExtractor(),
		custom.GetWwwFortinetComExtractor(),
		custom.GetWwwPublickey1JpExtractor(),
		custom.GetWwwGizmodoJpExtractor(),
		custom.BookwalkerJpExtractor,
		custom.GetTakagihiromitsuJpExtractor(),
		custom.GetWwwIpaGoJpExtractor(),
		custom.GetScanNetsecurityNeJpExtractor(),
		custom.GetWwwJnsaOrgExtractor(),
	}
	
	successCount := 0
	for _, extractor := range extractors {
		if extractor != nil {
			err := custom.GlobalRegistryManager.Register(extractor)
			if err != nil {
				log.Printf("Failed to register %s: %v", extractor.Domain, err)
			} else {
				fmt.Printf("✅ Registered: %s\n", extractor.Domain)
				successCount++
			}
		}
	}
	
	fmt.Printf("\n🎯 Registration completed: %d extractors registered\n", successCount)
}