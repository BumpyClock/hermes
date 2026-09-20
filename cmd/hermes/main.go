package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/BumpyClock/hermes"
)

var (
	outputFormat              string
	outputFile                string
	timeout                   time.Duration
	concurrency               int
	timing                    bool
	definitionsDirectory      string
	managedDefinitionsVersion string
	managedDefinitionsAuto    bool
	definitionsCacheDirectory string
)

var loadManagedDefinitions = hermes.LoadManagedDefinitions

func main() {
	rootCmd := &cobra.Command{
		Use:   "hermes",
		Short: "Hermes - High-performance web content extraction tool",
		Long:  "Hermes extracts clean, structured content from any web page with lightning speed",
	}

	parseCmd := &cobra.Command{
		Use:   "parse [url...]",
		Short: "Parse one or more URLs and extract content",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runParse,
	}

	parseCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json|html|markdown|text)")
	parseCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	parseCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Timeout per URL")
	parseCmd.Flags().IntVar(&concurrency, "concurrency", 10, "Maximum concurrent requests")
	parseCmd.Flags().BoolVar(&timing, "timing", false, "Show timing information for each URL")
	parseCmd.Flags().StringVar(&definitionsDirectory, "definitions", "", "Local YAML definitions directory (loaded once before parsing)")
	parseCmd.Flags().StringVar(&managedDefinitionsVersion, "managed-definitions", "", "Exact managed hermes-definitions release tag (loaded once before parsing)")
	parseCmd.Flags().BoolVar(&managedDefinitionsAuto, "managed-definitions-auto", false, "Follow the newest compatible managed definitions release at startup")
	parseCmd.Flags().StringVar(&definitionsCacheDirectory, "definitions-cache", "", "Updater-owned managed definitions cache directory")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hermes v1.1.1")
			fmt.Printf("Go version: %s\n", runtime.Version())
		},
	}

	rootCmd.AddCommand(parseCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runParse(cmd *cobra.Command, args []string) error {
	urls := args

	contentType := "html"
	switch outputFormat {
	case "markdown", "text":
		contentType = outputFormat
	}

	options := []hermes.Option{
		hermes.WithTimeout(timeout),
		hermes.WithContentType(contentType),
	}
	localDefinitions := definitionsDirectory != "" || flagChanged(cmd, "definitions")
	managedDefinitions := managedDefinitionsVersion != "" || managedDefinitionsAuto || flagChanged(cmd, "managed-definitions") || flagChanged(cmd, "managed-definitions-auto")
	cacheDirectory := definitionsCacheDirectory != "" || flagChanged(cmd, "definitions-cache")
	if localDefinitions && managedDefinitions {
		return fmt.Errorf("--definitions and --managed-definitions cannot be used together")
	}
	if managedDefinitionsVersion != "" && managedDefinitionsAuto {
		return fmt.Errorf("--managed-definitions and --managed-definitions-auto cannot be used together")
	}
	if cacheDirectory && !managedDefinitions {
		return fmt.Errorf("--definitions-cache requires --managed-definitions")
	}
	if localDefinitions {
		snapshot, err := hermes.LoadDefinitions(definitionsDirectory)
		if err != nil {
			return err
		}
		options = append(options, hermes.WithDefinitions(snapshot))
	}
	if managedDefinitions {
		if managedDefinitionsVersion == "" && !managedDefinitionsAuto {
			return fmt.Errorf("--managed-definitions requires an exact release tag")
		}
		loadContext := context.Background()
		if cmd != nil && cmd.Context() != nil {
			loadContext = cmd.Context()
		}
		managed, err := loadManagedDefinitions(loadContext, hermes.ManagedDefinitionsOptions{
			Version: managedDefinitionsVersion, Automatic: managedDefinitionsAuto, CacheDirectory: definitionsCacheDirectory,
		})
		if err != nil {
			return err
		}
		if managed.Warning != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", managed.Warning)
		}
		options = append(options, hermes.WithDefinitions(managed.Snapshot))
	}
	client := hermes.New(options...)

	// Use batch processing for concurrent parsing
	results := batchParse(client, urls)

	// Filter out failed results for output
	var successfulResults []ParseResult
	var totalParseTime time.Duration

	for _, result := range results {
		if result.Error != nil {
			if timing {
				fmt.Fprintf(os.Stderr, "Error parsing %s in %v: %v\n", result.URL, result.ParseTime, result.Error)
			}
			continue
		}

		totalParseTime += result.ParseTime
		successfulResults = append(successfulResults, result)

		if timing {
			fmt.Fprintf(os.Stderr, "Parsed %s in %v\n", result.URL, result.ParseTime)
		}
	}

	if timing && len(urls) > 1 {
		fmt.Fprintf(os.Stderr, "\nTiming Summary:\n")
		fmt.Fprintf(os.Stderr, "Total URLs processed: %d\n", len(urls))
		fmt.Fprintf(os.Stderr, "Successful parses: %d\n", len(successfulResults))
		fmt.Fprintf(os.Stderr, "Total parse time: %v\n", totalParseTime)
		if len(successfulResults) > 0 {
			avgTime := totalParseTime / time.Duration(len(successfulResults))
			fmt.Fprintf(os.Stderr, "Average parse time: %v\n", avgTime)
		}
	}

	if len(successfulResults) == 0 {
		return fmt.Errorf("no URLs were successfully parsed")
	}

	// Format output
	return formatOutput(successfulResults, len(urls) == 1)
}

func flagChanged(cmd *cobra.Command, name string) bool {
	return cmd != nil && cmd.Flags().Lookup(name) != nil && cmd.Flags().Changed(name)
}

// ParseResult holds the result of parsing a single URL.
type ParseResult struct {
	URL       string
	Result    *hermes.Result
	ParseTime time.Duration
	Error     error
}

// batchParse processes multiple URLs concurrently using semaphore pattern.
func batchParse(client hermes.Parser, urls []string) []ParseResult {
	results := make([]ParseResult, len(urls))
	sem := make(chan struct{}, concurrency) // Semaphore for concurrency control
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(index int, u string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			start := time.Now()
			result, err := client.Parse(ctx, u)
			parseTime := time.Since(start)

			results[index] = ParseResult{
				URL:       u,
				Result:    result,
				ParseTime: parseTime,
				Error:     err,
			}
		}(i, url)
	}

	wg.Wait()
	return results
}

type batchOutput struct {
	ParseTime string         `json:"parseTime"`
	Result    *hermes.Result `json:"result"`
	URL       string         `json:"url"`
}

// formatOutput formats the successful results according to the output format.
func formatOutput(results []ParseResult, singleURL bool) error {
	var output []byte
	var err error

	if singleURL {
		// Single URL - output the result directly in requested format
		result := results[0].Result
		switch outputFormat {
		case "json":
			output, err = json.MarshalIndent(result, "", "  ")
		case "html", "markdown", "text":
			// Content is already in the requested format from the parser
			output = []byte(result.Content)
		default:
			return fmt.Errorf("unsupported format: %s", outputFormat)
		}
	} else {
		// Multiple URLs - create JSON array with metadata
		var allResults []batchOutput
		for _, result := range results {
			allResults = append(allResults, batchOutput{
				ParseTime: result.ParseTime.String(),
				Result:    result.Result,
				URL:       result.URL,
			})
		}
		output, err = json.MarshalIndent(allResults, "", "  ")
	}

	if err != nil {
		return err
	}

	// Write output
	if outputFile != "" {
		//nolint:gosec // CLI output files are user-facing artifacts, so standard readable permissions are intentional.
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}
