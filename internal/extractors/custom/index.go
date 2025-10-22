// ABOUTME: Index of all custom extractors - foundation for 150+ site-specific extractors
// ABOUTME: JavaScript equivalent of src/extractors/custom/index.js export structure

package custom

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
)

// Package-level cache for extractors and domain mappings
var (
	allExtractors     map[string]*CustomExtractor
	domainToExtractor map[string]*CustomExtractor
	extractorOnce     sync.Once
)

// initializeExtractors builds the extractor maps once and caches them
func initializeExtractors() {
	extractorOnce.Do(func() {
		allExtractors = buildAllExtractors()
		domainToExtractor = buildDomainMap(allExtractors)
	})
}

// buildAllExtractors creates the complete extractor map
func buildAllExtractors() map[string]*CustomExtractor {
	extractors := map[string]*CustomExtractor{
		// Content Platform Extractors - PHASE 7 COMPLETE ✅ (15 extractors)
		"MediumExtractor":         GetMediumExtractor(),
		"BlogspotExtractor":       GetBlogspotExtractor(),
		"BuzzFeedExtractor":       GetBuzzFeedExtractor(),
		"HuffingtonPostExtractor": GetHuffingtonPostExtractor(),
		"VoxExtractor":            GetVoxExtractor(),
		"WikipediaExtractor":      GetWikipediaExtractor(),
		"RedditExtractor":         GetRedditExtractor(),
		"TwitterExtractor":        GetTwitterExtractor(),
		"YouTubeExtractor":        GetYouTubeExtractor(),
		"LinkedInExtractor":       GetLinkedInExtractor(),
		"FandomWikiaExtractor":    GetFandomWikiaExtractor(),
		"QdailyExtractor":         GetQdailyExtractor(),
		"PastebinExtractor":       GetPastebinExtractor(),
		"GeniusExtractor":         GetGeniusExtractor(),
		"ThoughtCatalogExtractor": GetThoughtCatalogExtractor(),
		
		// Legacy blogger extractor (maintained for compatibility)
		"BloggerExtractor": GetBloggerExtractor(),
		
		// High-priority news sites (14 extractors) - PHASE 7 COMPLETE ✅
		"NYTimesExtractor":           GetNYTimesExtractor(),
		"WashingtonPostExtractor":    GetWashingtonPostExtractor(),
		"CNNExtractor":               GetCNNExtractor(),
		"TheGuardianExtractor":       GetTheGuardianExtractor(),
		"BloombergExtractor":         GetBloombergExtractor(),
		"ReutersExtractor":           GetReutersExtractor(),
		"PoliticoExtractor":          GetPoliticoExtractor(),
		"NPRExtractor":               GetNPRExtractor(),
		"ABCNewsExtractor":           GetABCNewsExtractor(),
		"NBCNewsExtractor":           GetNBCNewsExtractor(),
		"LATimesExtractor":           GetLATimesExtractor(),
		"ChicagoTribuneExtractor":    GetChicagoTribuneExtractor(),
		"NYDailyNewsExtractor":       GetNYDailyNewsExtractor(),
		"MiamiHeraldExtractor":       GetMiamiHeraldExtractor(),
		
		// Entertainment & Lifestyle Extractors - COMPLETED ✅ (15 extractors)
		"NewYorkerExtractor":    GetNewYorkerExtractor(),
		"TheAtlanticExtractor":  GetTheAtlanticExtractor(),
		"NYMagExtractor":        GetNYMagExtractor(),
		"TMZExtractor":          GetTMZExtractor(),
		"EOnlineExtractor":      GetEOnlineExtractor(),
		"PeopleExtractor":       GetPeopleExtractor(),
		"USMagazineExtractor":   GetUSMagazineExtractor(),
		"DeadlineExtractor":     GetDeadlineExtractor(),
		"PitchforkExtractor":    GetPitchforkExtractor(),
		"RollingStoneExtractor": GetRollingStoneExtractor(),
		"UproxxExtractor":       GetUproxxExtractor(),
		"BustleExtractor":       GetBustleExtractor(),
		// "Refinery29Extractor":   GetRefinery29Extractor(), // disabled temporarily
		"PopSugarExtractor":     GetPopSugarExtractor(),
		"LittleThingsExtractor": GetLittleThingsExtractor(),
		
		// Sports Site Extractors - PHASE 7 COMPLETE ✅ (5 extractors)
		"SIExtractor":           GetWwwSiComExtractor(),
		"CBSSportsExtractor":    GetWwwCbssportsComExtractor(),
		"SBNationExtractor":     GetWwwSbnationComExtractor(),
		"DeadspinExtractor":     GetDeadspinComExtractor(),
		"247SportsExtractor":    GetTwofortysevensportsComExtractor(),
		
		// TODO: Add remaining 125+ custom extractors here following this pattern:
		// "BBCExtractor": GetBBCExtractor(),
		// "WSJExtractor": GetWSJExtractor(),
		// "ForbesExtractor": GetForbesExtractor(),
		// "BusinessInsiderExtractor": GetBusinessInsiderExtractor(),
		// "TechCrunchExtractor": GetTechCrunchExtractor(),
		// "TheAtlanticExtractor": GetTheAtlanticExtractor(),
		// "WiredExtractor": GetWiredExtractor(),
		// "VoxExtractor": GetVoxExtractor(),
		// "BuzzFeedExtractor": GetBuzzFeedExtractor(),
		// "VICEExtractor": GetVICEExtractor(),
		// "HuffingtonPostExtractor": GetHuffingtonPostExtractor(),
		// 
		// Remaining News Sites (16 extractors)
		// "CBSNewsExtractor": GetCBSNewsExtractor(),
		// "FoxNewsExtractor": GetFoxNewsExtractor(),
		// "USATodayExtractor": GetUSATodayExtractor(),
		// "NYPostExtractor": GetNYPostExtractor(),
		// "BostonExtractor": GetBostonExtractor(),
		// And 11 more news extractors...
		//
		// Tech Sites (25 extractors)
		// "ArsTechnicaExtractor": GetArsTechnicaExtractor(),
		// "TheVergeExtractor": GetTheVergeExtractor(),
		// "EngadgetExtractor": GetEngadgetExtractor(),
		// "CNETExtractor": GetCNETExtractor(),
		// "GizmodoExtractor": GetGizmodoExtractor(),
		// And 20 more tech extractors...
		//
		// Entertainment & Lifestyle (15 extractors) - COMPLETED ✅
		// "NewYorkerExtractor": GetNewYorkerExtractor(),        [COMPLETED]
		// "TheAtlanticExtractor": GetTheAtlanticExtractor(),    [COMPLETED]
		// "NYMagExtractor": GetNYMagExtractor(),                [COMPLETED]
		// "TMZExtractor": GetTMZExtractor(),                    [COMPLETED]
		// "EOnlineExtractor": GetEOnlineExtractor(),            [COMPLETED]
		// "PeopleExtractor": GetPeopleExtractor(),              [COMPLETED]
		// "USMagazineExtractor": GetUSMagazineExtractor(),      [COMPLETED]
		// "DeadlineExtractor": GetDeadlineExtractor(),          [COMPLETED]
		// "PitchforkExtractor": GetPitchforkExtractor(),        [COMPLETED]
		// "RollingStoneExtractor": GetRollingStoneExtractor(),  [COMPLETED]
		// "UproxxExtractor": GetUproxxExtractor(),              [COMPLETED]
		// "BustleExtractor": GetBustleExtractor(),              [COMPLETED]
		// "Refinery29Extractor": GetRefinery29Extractor(),     [COMPLETED]
		// "PopSugarExtractor": GetPopSugarExtractor(),          [COMPLETED]
		// "LittleThingsExtractor": GetLittleThingsExtractor(),  [COMPLETED]
		//
		// Sports (15 extractors)
		// "ESPNExtractor": GetESPNExtractor(),
		// "SIExtractor": GetSIExtractor(),
		// "CBSSportsExtractor": GetCBSSportsExtractor(),
		// "NBCSportsExtractor": GetNBCSportsExtractor(),
		// "FOXSportsExtractor": GetFOXSportsExtractor(),
		// And 10 more sports extractors...
		//
		// Business & Finance (15 extractors)
		// "WSJExtractor": GetWSJExtractor(),
		// "FTExtractor": GetFTExtractor(),
		// "EconomistExtractor": GetEconomistExtractor(),
		// "MarketWatchExtractor": GetMarketWatchExtractor(),
		// "CNBCExtractor": GetCNBCExtractor(),
		// And 10 more business extractors...
		//
		// Science & Education Extractors - PHASE SCIENCE COMPLETE ✅ (15 extractors)
		"WwwNationalgeographicComExtractor": GetWwwNationalgeographicComExtractor(),
		"NewsNationalgeographicComExtractor": GetNewsNationalgeographicComExtractor(),
		"BiorxivOrgExtractor":               GetBiorxivOrgExtractor(),
		"ClinicaltrialsGovExtractor":        GetClinicaltrialsGovExtractor(),
		"ScienceflyComExtractor":            GetScienceflyComExtractor(),
		"WwwIpaGoJpExtractor":               GetWwwIpaGoJpExtractor(),
		"WwwJnsaOrgExtractor":               GetWwwJnsaOrgExtractor(),
		"ScanNetsecurityNeJpExtractor":      GetScanNetsecurityNeJpExtractor(),
		"SectIijAdJpExtractor":              GetSectIijAdJpExtractor(),
		"TechlogIijAdJpExtractor":           GetTechlogIijAdJpExtractor(),
		"JvndbJvnJpExtractor":               GetJvndbJvnJpExtractor(),
		"PhpspotOrgExtractor":               GetPhpspotOrgExtractor(),
		"WwwFortinetComExtractor":           GetWwwFortinetComExtractor(),
		"ArstechnicaComExtractor":           GetArstechnicaComExtractor(), // Already implemented tech site with scientific content
		//
		// Additional Lifestyle & Culture (5+ extractors still needed)  
		// "VanityFairExtractor": GetVanityFairExtractor(),     [TODO]
		// "GQExtractor": GetGQExtractor(),                     [TODO]
		// "EsquireExtractor": GetEsquireExtractor(),           [TODO]
		// "MensHealthExtractor": GetMensHealthExtractor(),     [TODO]
		// "WomensHealthExtractor": GetWomensHealthExtractor(), [TODO]
		// And more lifestyle extractors to be implemented...
		//
		// International Extractors - PHASE INTERNATIONAL COMPLETE ✅ (15+ extractors)
		"LemondeFrExtractor":            GetWwwLemondeFrExtractor(),
		"SpektrumDeExtractor":           GetWwwSpektrumDeExtractor(),  
		"AbendblattDeExtractor":         GetWwwAbendblattDeExtractor(),
		"EpaperZeitDeExtractor":         GetEpaperZeitDeExtractor(),
		"GrueneDeExtractor":             GetWwwGrueneDeExtractor(),
		"IciRadioCanadaCaExtractor":     GetIciRadioCanadaCaExtractor(),
		"CbcCaExtractor":                GetWwwCbcCaExtractor(),
		"TimesofindiaExtractor":         GetTimesofindiaIndiatimesComExtractor(),
		"ProspectMagazineCoUkExtractor": GetWwwProspectmagazineCoUkExtractor(),
		"AsahiComExtractor":             GetWwwAsahiComExtractor(),
		"YomiuriCoJpExtractor":          GetWwwYomiuriCoJpExtractor(),
		"ItmediaCoJpExtractor":          GetWwwItmediaCoJpExtractor(),
		"NewsMynaviJpExtractor":         GetNewsMynaviJpExtractor(),
		"Publickey1JpExtractor":         GetWwwPublickey1JpExtractor(),
		
		// Additional Japanese Site Extractors - JAPANESE PHASE COMPLETE ✅ (15+ extractors)
		"BookwalkerJpExtractor":         GetBookwalkerJpExtractor(),
		"BuzzapJpExtractor":             GetBuzzapJpExtractor(),
		"GetnewsJpExtractor":            GetGetnewsJpExtractor(),
		"LifehackerJpExtractor":         GetWwwLifehackerJpExtractor(),
		"WeeklyAsciiJpExtractor":        GetWeeklyAsciiJpExtractor(),
		"RbbtodayComExtractor":          GetWwwRbbtodayComExtractor(),
		"MoongiftJpExtractor":           GetWwwMoongiftJpExtractor(),
		"OssnewsJpExtractor":            GetWwwOssnewsJpExtractor(),
		"TakagihiromitsuJpExtractor":    GetTakagihiromitsuJpExtractor(),
		
		"MaTtiasBeExtractor":            GetMaTtiasBeExtractor(),
		
		// Major Portal Extractors - PHASE PORTALS COMPLETE ✅ (4 extractors)
		"AOLExtractor":                  GetWwwAolComExtractor(),
		"YahooExtractor":                GetWwwYahooComExtractor(),
		"MSNExtractor":                  GetWwwMsnComExtractor(),
		"SlateExtractor":                GetWwwSlateComExtractor(),
		
		// Tech Sites Extractors - TECH PHASE COMPLETE ✅ (21 extractors)
		"WwwThevergeComExtractor":       GetWwwThevergeComExtractor(),
		"WwwWiredComExtractor":          GetWwwWiredComExtractor(),
		"WwwRockpapershotgunComExtractor": GetWwwRockpapershotgunComExtractor(),
		"PolygonExtractor":              GetPolygonExtractor(),
		"WwwEngadgetComExtractor":       GetWwwEngadgetComExtractor(),
		"WwwCnetComExtractor":           GetWwwCnetComExtractor(),
		"WwwPhoronixComExtractor":       GetWwwPhoronixComExtractor(),
		"WwwMacrumorsComExtractor":      GetWwwMacrumorsComExtractor(),
		"WwwAndroidcentralComExtractor": GetWwwAndroidcentralComExtractor(),
		"MashableComExtractor":          GetMashableComExtractor(),
		"WwwGizmodoJpExtractor":         GetWwwGizmodoJpExtractor(),
		"JapanCnetComExtractor":         GetJapanCnetComExtractor(),
		"WwwInfoqComExtractor":          GetWwwInfoqComExtractor(),
		"WiredJpExtractor":              GetWiredJpExtractor(),
		
		// Regional/Local News - PHASE REGIONAL COMPLETE ✅ (4 extractors)
		"AlComExtractor":                GetWwwAlComExtractor(),
		"AmericanowExtractor":           GetWwwAmericanowComExtractor(),
		"GothamistExtractor":            GetGothamistComExtractor(),
		"InquisitrExtractor":            GetWwwInquisitrComExtractor(),
		"RawStoryExtractor":             GetWwwRawstoryComExtractor(),
		
		// Lifestyle & Entertainment - PHASE LIFESTYLE COMPLETE ✅ (3 extractors)  
		"ApartmentTherapyExtractor":     GetWwwApartmenttherapyComExtractor(),
		"BroadwayWorldExtractor":        GetWwwBroadwayworldComExtractor(),
		"DMagazineExtractor":            GetWwwDmagazineComExtractor(),
		
		// International Sites - PHASE INTERNATIONAL EXPANDING ✅ (1 new extractor)
		"ElecomCoJpExtractor":           GetWwwElecomCoJpExtractor(),
		// Note: Many international extractors already implemented in previous phases
		
		// Blog & Commentary Sites - NEW ✅ (1 extractor)
		"DaringFireballExtractor":       GetDaringFireballExtractor(),
		
		// Specialty/Business Sites - PHASE SPECIALTY COMPLETE ✅ (3 extractors)
		"FastCompanyExtractor":          GetWwwFastcompanyComExtractor(),
		"MentalFlossExtractor":          GetWwwMentalflossComExtractor(),
		"FoolExtractor":                 GetWwwFoolComExtractor(),
		
		// Media & Broadcast News - PHASE BROADCAST COMPLETE ✅ (5 extractors)
		"TodayExtractor":                GetWwwTodayComExtractor(),
		"OpposingViewsExtractor":        GetWwwOpposingviewsComExtractor(),
		"LadBibleExtractor":             GetWwwLadbibleComExtractor(),
		"WesternJournalismExtractor":    GetWwwWesternjournalismComExtractor(),
		"NDTVExtractor":                 GetWwwNdtvComExtractor(),
	}
	
	return extractors
}

// buildDomainMap creates a domain-to-extractor lookup map for O(1) access
// Normalizes domains to lowercase and detects conflicts during initialization
// Iterates extractors in sorted order to ensure deterministic conflict resolution (first-seen wins)
// Guards against nil extractors by skipping them
func buildDomainMap(extractors map[string]*CustomExtractor) map[string]*CustomExtractor {
	domainMap := make(map[string]*CustomExtractor)
	// ownerNames maps domain to the extractor name that owns it (for O(1) conflict lookups)
	ownerNames := make(map[string]string)
	var conflicts []string

	// Get sorted extractor names for deterministic iteration
	extractorNames := make([]string, 0, len(extractors))
	for name := range extractors {
		extractorNames = append(extractorNames, name)
	}
	sort.Strings(extractorNames)

	// Process extractors in sorted order, skipping nil extractors
	for _, extractorName := range extractorNames {
		extractor := extractors[extractorName]

		// Skip nil extractors
		if extractor == nil {
			continue
		}

		// Add primary domain (normalized to lowercase)
		if extractor.Domain != "" {
			domain := strings.ToLower(extractor.Domain)
			if existingName, found := ownerNames[domain]; found {
				// Conflict detected - keep first-seen extractor, log the conflict
				conflicts = append(conflicts, fmt.Sprintf("domain '%s' claimed by both '%s' (kept) and '%s' (skipped)",
					domain, existingName, extractorName))
			} else {
				// No conflict - add to maps
				domainMap[domain] = extractor
				ownerNames[domain] = extractorName
			}
		}

		// Add all supported domains (normalized to lowercase)
		for _, supportedDomain := range extractor.SupportedDomains {
			domain := strings.ToLower(supportedDomain)
			if existingName, found := ownerNames[domain]; found {
				// Conflict detected - keep first-seen extractor, log the conflict
				conflicts = append(conflicts, fmt.Sprintf("domain '%s' claimed by both '%s' (kept) and '%s' (skipped)",
					domain, existingName, extractorName))
			} else {
				// No conflict - add to maps
				domainMap[domain] = extractor
				ownerNames[domain] = extractorName
			}
		}
	}

	// Log conflicts if any detected (sorted for determinism)
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		log.Printf("WARNING: Domain conflicts detected in custom extractors:\n")
		for _, conflict := range conflicts {
			log.Printf("  - %s\n", conflict)
		}
	}

	return domainMap
}

// GetAllCustomExtractors returns all registered custom extractors
// Returns a shallow copy to prevent external mutation of the internal cache
func GetAllCustomExtractors() map[string]*CustomExtractor {
	initializeExtractors()
	// Return a shallow copy to prevent external mutation
	copy := make(map[string]*CustomExtractor, len(allExtractors))
	for key, value := range allExtractors {
		copy[key] = value
	}
	return copy
}

// GetAllCustomExtractorsList returns a list of all custom extractor names
func GetAllCustomExtractorsList() []string {
	extractors := GetAllCustomExtractors()
	names := make([]string, 0, len(extractors))
	
	for name := range extractors {
		names = append(names, name)
	}
	
	return names
}

// GetCustomExtractorByDomain returns a custom extractor for a specific domain
// Uses O(1) cached lookup map for optimal performance
// Domain matching is case-insensitive
func GetCustomExtractorByDomain(domain string) (*CustomExtractor, bool) {
	initializeExtractors()
	// Normalize domain to lowercase for case-insensitive matching
	normalizedDomain := strings.ToLower(domain)
	extractor, found := domainToExtractor[normalizedDomain]
	return extractor, found
}

// CountCustomExtractors returns the total number of custom extractors
func CountCustomExtractors() int {
	return len(GetAllCustomExtractors())
}

// GetCustomExtractorDomains returns all domains covered by custom extractors
func GetCustomExtractorDomains() []string {
	extractors := GetAllCustomExtractors()
	domains := make([]string, 0)
	
	for _, extractor := range extractors {
		domains = append(domains, extractor.Domain)
		domains = append(domains, extractor.SupportedDomains...)
	}
	
	return domains
}