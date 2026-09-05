package hermes_test

import (
	"context"
	"testing"

	"github.com/BumpyClock/hermes"
)

func TestParseHTMLPreservesSiteMetadata(t *testing.T) {
	client := hermes.New(hermes.WithAllowPrivateNetworks(true))
	html := `<!doctype html>
<html>
<head>
  <meta property="og:site_name" content="Example Site">
  <meta property="og:title" content="Example Site Title">
  <meta property="og:image" content="https://cdn.example.com/social.png">
</head>
<body><article><h1>Article title</h1><p>Article content.</p></article></body>
</html>`

	result, err := client.ParseHTML(context.Background(), html, "http://127.0.0.1/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}
	if result.SiteName != "Example Site" {
		t.Errorf("SiteName = %q, want %q", result.SiteName, "Example Site")
	}
	if result.SiteTitle != "Example Site Title" {
		t.Errorf("SiteTitle = %q, want %q", result.SiteTitle, "Example Site Title")
	}
	if result.SiteImage != "https://cdn.example.com/social.png" {
		t.Errorf("SiteImage = %q, want %q", result.SiteImage, "https://cdn.example.com/social.png")
	}
}
