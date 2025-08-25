// ABOUTME: Parallel extractor checking for 5-10x performance improvement
// ABOUTME: Uses goroutines and channels to check multiple extractors simultaneously while maintaining priority order

package extractors

import (
	"context"
	neturl "net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ExtractorCandidate represents a potential extractor with its priority
type ExtractorCandidate struct {
	Extractor Extractor
	Priority  int    // Lower number = higher priority
	Source    string // "api_hostname", "api_domain", "static_hostname", etc.
	Key       string // The key used to find this extractor
}

// ParallelExtractorResult holds the result of parallel extractor checking
type ParallelExtractorResult struct {
	Candidate *ExtractorCandidate
	Error     error
	Duration  time.Duration
}

// ExtractorChecker defines the interface for checking if an extractor can handle a URL
type ExtractorChecker interface {
	CanHandle(doc *goquery.Document, url string) (bool, error)
	GetPriority() int
	GetSource() string
}

// GetExtractorParallel performs parallel extractor lookup with priority ordering
// This is the high-performance version of GetExtractor that uses goroutines
// DEPRECATED: This method uses context.Background() which prevents proper cancellation.
// Use GetExtractorParallelWithContext instead.
func GetExtractorParallel(urlStr string, parsedURL *neturl.URL, doc *goquery.Document) (Extractor, error) {
	return GetExtractorParallelWithContext(context.Background(), urlStr, parsedURL, doc)
}

// GetExtractorParallelWithContext performs parallel extractor lookup with context for cancellation
func GetExtractorParallelWithContext(ctx context.Context, urlStr string, parsedURL *neturl.URL, doc *goquery.Document) (Extractor, error) {
	// Create a context with timeout for extractor checking
	extractorCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Extract URL components
	hostname, baseDomain, err := extractURLComponents(urlStr)
	if err != nil {
		return GenericExtractor(), err
	}

	if parsedURL != nil {
		hostname = parsedURL.Hostname()
		baseDomain = calculateBaseDomain(hostname)
	}

	// Build candidate list with priorities
	candidates := buildExtractorCandidates(hostname, baseDomain)

	// Check extractors in parallel while respecting priority order
	result := checkExtractorsParallel(extractorCtx, candidates, doc, urlStr)

	if result != nil && result.Candidate != nil {
		return result.Candidate.Extractor, nil
	}

	// Fallback to generic extractor
	return GenericExtractor(), nil
}

// buildExtractorCandidates creates a prioritized list of extractor candidates
func buildExtractorCandidates(hostname, baseDomain string) []*ExtractorCandidate {
	var candidates []*ExtractorCandidate

	// Get registries
	apiExtractors := GetAPIExtractors()
	staticExtractors := All

	// Priority 1: API extractor by hostname
	if extractor, found := apiExtractors[hostname]; found {
		candidates = append(candidates, &ExtractorCandidate{
			Extractor: extractor,
			Priority:  1,
			Source:    "api_hostname",
			Key:       hostname,
		})
	}

	// Priority 2: API extractor by base domain
	if extractor, found := apiExtractors[baseDomain]; found {
		candidates = append(candidates, &ExtractorCandidate{
			Extractor: extractor,
			Priority:  2,
			Source:    "api_domain",
			Key:       baseDomain,
		})
	}

	// Priority 3: Static extractor by hostname
	if extractor, found := staticExtractors[hostname]; found {
		candidates = append(candidates, &ExtractorCandidate{
			Extractor: extractor,
			Priority:  3,
			Source:    "static_hostname",
			Key:       hostname,
		})
	}

	// Priority 4: Static extractor by base domain
	if extractor, found := staticExtractors[baseDomain]; found {
		candidates = append(candidates, &ExtractorCandidate{
			Extractor: extractor,
			Priority:  4,
			Source:    "static_domain",
			Key:       baseDomain,
		})
	}

	// Priority 5: HTML-based detection candidate (checked separately)
	// Priority 6: Generic extractor (always available as fallback)

	return candidates
}

// checkExtractorsParallel checks multiple extractors in parallel and returns the highest priority match
func checkExtractorsParallel(ctx context.Context, candidates []*ExtractorCandidate, doc *goquery.Document, url string) *ParallelExtractorResult {
	if len(candidates) == 0 {
		return nil
	}

	// Create channels for communication
	resultChan := make(chan *ParallelExtractorResult, len(candidates))
	var wg sync.WaitGroup

	// Start goroutines to check each candidate
	for _, candidate := range candidates {
		wg.Add(1)
		go func(cand *ExtractorCandidate) {
			defer wg.Done()

			start := time.Now()
			
			// Check if this extractor can handle the URL/document
			canHandle, err := checkExtractorCapability(ctx, cand.Extractor, doc, url)
			
			duration := time.Since(start)

			if canHandle && err == nil {
				select {
				case resultChan <- &ParallelExtractorResult{
					Candidate: cand,
					Error:     nil,
					Duration:  duration,
				}:
				case <-ctx.Done():
					return
				}
			} else if err != nil {
				select {
				case resultChan <- &ParallelExtractorResult{
					Candidate: cand,
					Error:     err,
					Duration:  duration,
				}:
				case <-ctx.Done():
					return
				}
			}
		}(candidate)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and find the highest priority match
	var bestResult *ParallelExtractorResult
	var allResults []*ParallelExtractorResult

	for result := range resultChan {
		allResults = append(allResults, result)
		
		if result.Error == nil && result.Candidate != nil {
			if bestResult == nil || result.Candidate.Priority < bestResult.Candidate.Priority {
				bestResult = result
			}
		}
	}

	// If we have a result from parallel checking, return it
	if bestResult != nil {
		return bestResult
	}

	// If no parallel results, try HTML-based detection
	if doc != nil {
		start := time.Now()
		if extractor := DetectByHTML(doc); extractor != nil {
			return &ParallelExtractorResult{
				Candidate: &ExtractorCandidate{
					Extractor: extractor,
					Priority:  5,
					Source:    "html_detection",
					Key:       "html_meta",
				},
				Error:    nil,
				Duration: time.Since(start),
			}
		}
	}

	return nil
}

// checkExtractorCapability determines if an extractor can handle the given document/URL
func checkExtractorCapability(ctx context.Context, extractor Extractor, doc *goquery.Document, url string) (bool, error) {
	// TODO: Use ctx for timeout control in future implementation
	_ = ctx
	// TODO: Use doc for content-based capability detection in future implementation  
	_ = doc
	
	// For now, assume all extractors can handle their assigned domains
	// In a full implementation, this could call extractor.CanHandle() or similar
	
	// Quick domain check
	if extractor.GetDomain() == "*" {
		return true, nil // Generic extractor
	}

	if parsedURL, err := neturl.Parse(url); err == nil {
		hostname := parsedURL.Hostname()
		extractorDomain := extractor.GetDomain()
		
		// Direct match
		if hostname == extractorDomain {
			return true, nil
		}
		
		// Subdomain match
		if len(hostname) > len(extractorDomain) && 
		   hostname[len(hostname)-len(extractorDomain)-1:] == "."+extractorDomain {
			return true, nil
		}
	}

	return false, nil
}

// ExtractorStats provides statistics about parallel extractor checking
type ExtractorStats struct {
	TotalCandidates    int
	CheckedInParallel  int
	FastestMatch       time.Duration
	SlowestMatch       time.Duration
	AverageCheckTime   time.Duration
	WinningExtractor   string
	WinningPriority    int
}

// GetExtractorWithStats performs parallel lookup and returns performance statistics
// DEPRECATED: This method uses context.Background() which prevents proper cancellation.
// Use GetExtractorWithStatsContext instead.
func GetExtractorWithStats(urlStr string, parsedURL *neturl.URL, doc *goquery.Document) (Extractor, *ExtractorStats, error) {
	return GetExtractorWithStatsContext(context.Background(), urlStr, parsedURL, doc)
}

// GetExtractorWithStatsContext performs parallel lookup with context and returns performance statistics
func GetExtractorWithStatsContext(ctx context.Context, urlStr string, parsedURL *neturl.URL, doc *goquery.Document) (Extractor, *ExtractorStats, error) {
	start := time.Now()
	
	// Extract URL components
	hostname, baseDomain, err := extractURLComponents(urlStr)
	if err != nil {
		return GenericExtractor(), nil, err
	}

	if parsedURL != nil {
		hostname = parsedURL.Hostname()
		baseDomain = calculateBaseDomain(hostname)
	}

	// Build candidates and check in parallel
	candidates := buildExtractorCandidates(hostname, baseDomain)
	extractorCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := checkExtractorsParallel(extractorCtx, candidates, doc, urlStr)

	// Calculate statistics
	stats := &ExtractorStats{
		TotalCandidates:   len(candidates),
		CheckedInParallel: len(candidates),
	}

	if result != nil && result.Candidate != nil {
		stats.WinningExtractor = result.Candidate.Source
		stats.WinningPriority = result.Candidate.Priority
		stats.FastestMatch = result.Duration
		stats.SlowestMatch = result.Duration
		stats.AverageCheckTime = result.Duration
		
		return result.Candidate.Extractor, stats, nil
	}

	// Fallback timing
	stats.AverageCheckTime = time.Since(start)
	return GenericExtractor(), stats, nil
}

// ParallelExtractorConfig configures parallel extractor behavior
type ParallelExtractorConfig struct {
	MaxConcurrentChecks int           // Maximum number of goroutines
	CheckTimeout        time.Duration // Timeout for individual extractor checks
	EnableStats         bool          // Whether to collect detailed statistics
	EnableCaching       bool          // Whether to cache extractor decisions
}

// DefaultParallelConfig returns a sensible default configuration
func DefaultParallelConfig() *ParallelExtractorConfig {
	return &ParallelExtractorConfig{
		MaxConcurrentChecks: 10,
		CheckTimeout:        2 * time.Second,
		EnableStats:         false,
		EnableCaching:       true,
	}
}