package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/BumpyClock/hermes"
)

var (
	outputFormat string
	outputFile   string
	headers      string
	timeout      time.Duration
	concurrency  int
	timing       bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "parser",
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
	parseCmd.Flags().StringVar(&headers, "headers", "", "Custom headers as JSON")
	parseCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Timeout per URL")
	parseCmd.Flags().IntVar(&concurrency, "concurrency", 10, "Maximum concurrent requests")
	parseCmd.Flags().BoolVar(&timing, "timing", false, "Show timing information for each URL")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hermes v0.1.0")
			fmt.Println("Go version: 1.24.6")
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

	// Parse custom headers if provided
	customHeaders := make(map[string]string)
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &customHeaders); err != nil {
			return fmt.Errorf("invalid headers JSON: %w", err)
		}
	}

	// Create hermes client with options
	clientOptions := []hermes.Option{
		hermes.WithTimeout(timeout),
	}
	
	// Set content type based on output format for the parser
	// This determines how the content is extracted, not just how it's formatted
	switch outputFormat {
	case "html":
		clientOptions = append(clientOptions, hermes.WithContentType("html"))
	case "markdown":
		clientOptions = append(clientOptions, hermes.WithContentType("markdown"))
	case "text":
		clientOptions = append(clientOptions, hermes.WithContentType("text"))
	default:
		clientOptions = append(clientOptions, hermes.WithContentType("html"))
	}
	
	// Add custom headers if provided
	if len(customHeaders) > 0 {
		// TODO: Add header support to hermes.Option - for now we'll skip this
		// Will add hermes.WithHeaders() in future enhancement
	}
	
	client := hermes.New(clientOptions...)

	// Use batch processing for concurrent parsing
	results, err := batchParse(client, urls)
	if err != nil {
		return err
	}

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
		avgTime := totalParseTime / time.Duration(len(successfulResults))
		fmt.Fprintf(os.Stderr, "\nTiming Summary:\n")
		fmt.Fprintf(os.Stderr, "Total URLs processed: %d\n", len(urls))
		fmt.Fprintf(os.Stderr, "Successful parses: %d\n", len(successfulResults))
		fmt.Fprintf(os.Stderr, "Total parse time: %v\n", totalParseTime)
		if len(successfulResults) > 0 {
			fmt.Fprintf(os.Stderr, "Average parse time: %v\n", avgTime)
		}
	}

	if len(successfulResults) == 0 {
		return fmt.Errorf("no URLs were successfully parsed")
	}

	// Format output
	return formatOutput(successfulResults, len(urls) == 1)
}

// ParseResult holds the result of parsing a single URL
type ParseResult struct {
	URL       string
	Result    *hermes.Result
	ParseTime time.Duration
	Error     error
}

// batchParse processes multiple URLs concurrently using semaphore pattern
func batchParse(client *hermes.Client, urls []string) ([]ParseResult, error) {
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
	return results, nil
}

// formatOutput formats the successful results according to the output format
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
		var allResults []interface{}
		for _, result := range results {
			// Content is already in the requested format from the parser
			convertedContent := result.Result.Content

			allResults = append(allResults, map[string]interface{}{
				"url":             result.URL,
				"parseTime":       result.ParseTime.String(),
				"result":          result.Result,
				"convertedContent": convertedContent,
			})
		}
		output, err = json.MarshalIndent(allResults, "", "  ")
	}

	if err != nil {
		return err
	}

	// Write output
	if outputFile != "" {
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}