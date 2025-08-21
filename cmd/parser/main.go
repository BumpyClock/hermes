package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/BumpyClock/hermes/pkg/parser"
)

var (
	outputFormat string
	outputFile   string
	headers      string
	fetchAll     bool
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
	parseCmd.Flags().BoolVar(&fetchAll, "fetch-all", true, "Fetch all pages for multi-page articles")
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

	// Create parser once and reuse
	p := parser.New()
	
	var allResults []interface{}
	var totalParseTime time.Duration
	
	for i, url := range urls {
		if timing {
			fmt.Fprintf(os.Stderr, "Parsing URL %d/%d: %s\n", i+1, len(urls), url)
		}
		
		start := time.Now()
		result, err := p.Parse(url, &parser.ParserOptions{
			FetchAllPages: fetchAll,
			ContentType:   outputFormat,
			Headers:       customHeaders,
		})
		parseTime := time.Since(start)
		totalParseTime += parseTime

		if err != nil {
			if timing {
				fmt.Fprintf(os.Stderr, "Error parsing %s in %v: %v\n", url, parseTime, err)
			}
			continue
		}

		// Check if result contains an error
		if result.IsError() {
			if timing {
				fmt.Fprintf(os.Stderr, "Parser error for %s in %v: %s\n", url, parseTime, result.Message)
			}
			continue
		}

		if timing {
			fmt.Fprintf(os.Stderr, "Parsed %s in %v\n", url, parseTime)
		}

		// For multiple URLs, collect results in array
		if len(urls) > 1 {
			// Include timing and converted content for batch processing
			var convertedContent string
			switch outputFormat {
			case "json":
				convertedContent = result.Content // Already in requested format
			case "html":
				convertedContent = result.Content
			case "markdown":
				convertedContent = result.FormatMarkdown()
			case "text":
				convertedContent = result.Content
			default:
				convertedContent = result.Content
			}
			
			allResults = append(allResults, map[string]interface{}{
				"url":        url,
				"parseTime":  parseTime.String(),
				"result":     result,
				"convertedContent": convertedContent, // Add converted content for easy access
			})
		} else {
			// Single URL - output just the result
			allResults = append(allResults, result)
		}
	}

	if timing && len(urls) > 1 {
		avgTime := totalParseTime / time.Duration(len(urls))
		fmt.Fprintf(os.Stderr, "\nTiming Summary:\n")
		fmt.Fprintf(os.Stderr, "Total URLs: %d\n", len(urls))
		fmt.Fprintf(os.Stderr, "Total parse time: %v\n", totalParseTime)
		fmt.Fprintf(os.Stderr, "Average parse time: %v\n", avgTime)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no URLs were successfully parsed")
	}

	// Format output
	var output []byte
	var err error
	
	if len(urls) == 1 {
		// Single URL - output the result directly
		result := allResults[0]
		switch outputFormat {
		case "json":
			output, err = json.MarshalIndent(result, "", "  ")
		case "html":
			if r, ok := result.(*parser.Result); ok {
				output = []byte(r.Content)
			}
		case "markdown":
			if r, ok := result.(*parser.Result); ok {
				output = []byte(r.FormatMarkdown())
			}
		case "text":
			if r, ok := result.(*parser.Result); ok {
				output = []byte(r.Content)
			}
		default:
			return fmt.Errorf("unsupported format: %s", outputFormat)
		}
	} else {
		// Multiple URLs - always output as JSON array
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