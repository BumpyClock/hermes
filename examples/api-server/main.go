// Package main demonstrates how to build an HTTP API server using Hermes.
//
// This example shows how to:
// - Create a REST API for content extraction
// - Handle different content formats (JSON, HTML, Markdown, Text)
// - Implement proper error handling and HTTP status codes
// - Add request validation and rate limiting basics
// - Structure responses consistently
// - Handle concurrent requests efficiently
//
// Run with: go run examples/api-server/main.go
// Test with: curl "http://localhost:8080/parse?url=https://example.com&format=json"
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BumpyClock/hermes"
)

// Server holds the Hermes client and server configuration
type Server struct {
	hermes *hermes.Client
	port   string
}

// ParseRequest represents the request payload for POST requests
type ParseRequest struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"`
}

// ParseResponse represents the API response structure
type ParseResponse struct {
	Success   bool                   `json:"success"`
	Data      *hermes.Result         `json:"data,omitempty"`
	Error     *ErrorDetail           `json:"error,omitempty"`
	Metadata  *ResponseMetadata      `json:"metadata,omitempty"`
}

// ErrorDetail provides structured error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// ResponseMetadata includes timing and processing information
type ResponseMetadata struct {
	ProcessingTime string `json:"processing_time"`
	Timestamp      string `json:"timestamp"`
	Version        string `json:"version"`
}

func main() {
	fmt.Println("Hermes API Server Example")
	fmt.Println("=========================")

	// Create optimized Hermes client for server use
	hermesClient := hermes.New(
		hermes.WithTimeout(30*time.Second),
		hermes.WithUserAgent("HermesAPIServer/1.0"),
		hermes.WithContentType("html"), // Default format
		// hermes.WithAllowPrivateNetworks(false), // SSRF protection enabled by default
	)

	// Create server
	server := &Server{
		hermes: hermesClient,
		port:   "8080",
	}

	// Setup routes
	http.HandleFunc("/", server.handleHome)
	http.HandleFunc("/parse", server.handleParse)
	http.HandleFunc("/health", server.handleHealth)

	// Start server
	addr := ":" + server.port
	fmt.Printf("🚀 Server starting on http://localhost%s\n", addr)
	fmt.Println("\nAPI Endpoints:")
	fmt.Printf("  GET  /                           - API documentation\n")
	fmt.Printf("  GET  /parse?url=<url>&format=<f> - Parse URL (GET)\n")
	fmt.Printf("  POST /parse                      - Parse URL (POST JSON)\n")
	fmt.Printf("  GET  /health                     - Health check\n")
	fmt.Println("\nExample requests:")
	fmt.Printf("  curl \"http://localhost%s/parse?url=https://example.com&format=json\"\n", addr)
	fmt.Printf("  curl -X POST http://localhost%s/parse -H 'Content-Type: application/json' -d '{\"url\":\"https://example.com\",\"format\":\"markdown\"}'\n", addr)
	fmt.Println()

	log.Fatal(http.ListenAndServe(addr, nil))
}

// handleHome serves API documentation
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Hermes API Server</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { color: #2196F3; font-weight: bold; }
        code { background: #eee; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🚀 Hermes API Server</h1>
    <p>Web content extraction API powered by Hermes library.</p>
    
    <h2>Endpoints</h2>
    
    <div class="endpoint">
        <h3><span class="method">GET</span> /parse</h3>
        <p>Extract content from a URL using query parameters.</p>
        <p><strong>Parameters:</strong></p>
        <ul>
            <li><code>url</code> - The URL to parse (required)</li>
            <li><code>format</code> - Output format: json, html, markdown, text (optional, default: json)</li>
        </ul>
        <p><strong>Example:</strong><br>
        <code>GET /parse?url=https://example.com&format=markdown</code></p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method">POST</span> /parse</h3>
        <p>Extract content from a URL using JSON payload.</p>
        <p><strong>Request Body:</strong></p>
        <pre><code>{
  "url": "https://example.com",
  "format": "json"
}</code></pre>
    </div>
    
    <div class="endpoint">
        <h3><span class="method">GET</span> /health</h3>
        <p>Health check endpoint.</p>
    </div>
    
    <h2>Supported Formats</h2>
    <ul>
        <li><code>json</code> - Structured JSON response with all extracted fields</li>
        <li><code>html</code> - Clean HTML content</li>
        <li><code>markdown</code> - Markdown formatted content</li>
        <li><code>text</code> - Plain text content</li>
    </ul>
</body>
</html>`
	
	fmt.Fprint(w, html)
}

// handleParse handles content extraction requests (both GET and POST)
func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Extract parameters based on request method
	var targetURL, format string
	var err error
	
	switch r.Method {
	case http.MethodGet:
		targetURL, format, err = s.parseGETParams(r)
	case http.MethodPost:
		targetURL, format, err = s.parsePOSTParams(r)
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET and POST methods are supported", "", start)
		return
	}
	
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error(), targetURL, start)
		return
	}
	
	// Validate URL
	if targetURL == "" {
		s.sendError(w, http.StatusBadRequest, "missing_url", "URL parameter is required", "", start)
		return
	}
	
	if !s.isValidURL(targetURL) {
		s.sendError(w, http.StatusBadRequest, "invalid_url", "Invalid URL format", targetURL, start)
		return
	}
	
	// Default format
	if format == "" {
		format = "json"
	}
	
	// Validate format
	if !s.isValidFormat(format) {
		s.sendError(w, http.StatusBadRequest, "invalid_format", "Format must be one of: json, html, markdown, text", targetURL, start)
		return
	}
	
	// Parse the URL
	s.parseURL(w, r, targetURL, format, start)
}

// parseGETParams extracts parameters from GET request
func (s *Server) parseGETParams(r *http.Request) (string, string, error) {
	targetURL := r.URL.Query().Get("url")
	format := r.URL.Query().Get("format")
	return targetURL, format, nil
}

// parsePOSTParams extracts parameters from POST request JSON body
func (s *Server) parsePOSTParams(r *http.Request) (string, string, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return "", "", fmt.Errorf("Content-Type must be application/json")
	}
	
	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", "", fmt.Errorf("invalid JSON payload")
	}
	
	return req.URL, req.Format, nil
}

// parseURL performs the actual content extraction
func (s *Server) parseURL(w http.ResponseWriter, r *http.Request, targetURL, format string, start time.Time) {
	// Create client with appropriate content type for extraction
	var contentType string
	if format == "json" {
		contentType = "html" // Use HTML for JSON responses
	} else {
		contentType = format
	}
	
	client := hermes.New(
		hermes.WithTimeout(25*time.Second),
		hermes.WithUserAgent("HermesAPIServer/1.0"),
		hermes.WithContentType(contentType),
	)
	
	// Create request context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	
	// Parse the URL
	result, err := client.Parse(ctx, targetURL)
	
	if err != nil {
		var code, message string
		
                if parseErr, ok := err.(*hermes.ParseError); ok {
                        // Use the ErrorCode's String method to avoid
                        // converting the numeric code to a single rune.
                        code = parseErr.Code.String()
                        message = parseErr.Err.Error()
                } else {
                        code = "parse_error"
                        message = err.Error()
                }
		
		s.sendError(w, http.StatusBadGateway, code, message, targetURL, start)
		return
	}
	
	// Send successful response
	s.sendSuccess(w, result, format, targetURL, start)
}

// sendSuccess sends a successful response in the requested format
func (s *Server) sendSuccess(w http.ResponseWriter, result *hermes.Result, format, url string, start time.Time) {
	duration := time.Since(start)
	
	// For non-JSON formats, return content directly
	if format != "json" {
		var contentType string
		switch format {
		case "html":
			contentType = "text/html"
		case "markdown":
			contentType = "text/markdown"
		case "text":
			contentType = "text/plain"
		}
		
		w.Header().Set("Content-Type", contentType+"; charset=utf-8")
		w.Header().Set("X-Processing-Time", duration.String())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, result.Content)
		return
	}
	
	// JSON response
	response := ParseResponse{
		Success: true,
		Data:    result,
		Metadata: &ResponseMetadata{
			ProcessingTime: duration.String(),
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			Version:        "1.0",
		},
	}
	
	s.sendJSON(w, http.StatusOK, response)
}

// sendError sends an error response
func (s *Server) sendError(w http.ResponseWriter, status int, code, message, url string, start time.Time) {
	duration := time.Since(start)
	
	response := ParseResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			URL:     url,
		},
		Metadata: &ResponseMetadata{
			ProcessingTime: duration.String(),
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			Version:        "1.0",
		},
	}
	
	s.sendJSON(w, status, response)
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0",
	}
	
	s.sendJSON(w, http.StatusOK, health)
}

// isValidURL validates URL format
func (s *Server) isValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	
	return u.Scheme == "http" || u.Scheme == "https"
}

// isValidFormat validates output format
func (s *Server) isValidFormat(format string) bool {
	validFormats := []string{"json", "html", "markdown", "text"}
	format = strings.ToLower(format)
	
	for _, valid := range validFormats {
		if format == valid {
			return true
		}
	}
	
	return false
}