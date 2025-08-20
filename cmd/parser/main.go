package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/BumpyClock/parser-go/pkg/parser"
)

var (
	outputFormat string
	outputFile   string
	headers      string
	fetchAll     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "parser",
		Short: "Postlight Parser - Web content extraction tool",
		Long:  "Extract clean, structured content from any web page",
	}

	parseCmd := &cobra.Command{
		Use:   "parse [url]",
		Short: "Parse a URL and extract content",
		Args:  cobra.ExactArgs(1),
		RunE:  runParse,
	}

	parseCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json|html|markdown|text)")
	parseCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	parseCmd.Flags().StringVar(&headers, "headers", "", "Custom headers as JSON")
	parseCmd.Flags().BoolVar(&fetchAll, "fetch-all", true, "Fetch all pages for multi-page articles")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Postlight Parser Go v0.1.0")
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
	url := args[0]

	// Parse custom headers if provided
	customHeaders := make(map[string]string)
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &customHeaders); err != nil {
			return fmt.Errorf("invalid headers JSON: %w", err)
		}
	}

	// Create parser and extract content
	p := parser.New()
	result, err := p.Parse(url, parser.ParserOptions{
		FetchAllPages: fetchAll,
		ContentType:   outputFormat,
		Headers:       customHeaders,
	})

	if err != nil {
		return err
	}

	// Check if result contains an error
	if result.IsError() {
		return fmt.Errorf("%s", result.Message)
	}

	// Format output
	var output []byte
	switch outputFormat {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "html":
		output = []byte(result.Content)
	case "markdown", "text":
		output = []byte(result.Content)
	default:
		return fmt.Errorf("unsupported format: %s", outputFormat)
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