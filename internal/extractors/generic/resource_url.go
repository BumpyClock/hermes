package generic

import (
	"net/url"
	"strings"
)

func normalizeResourceURL(resourceURL, pageURL string) string {
	resourceURL = strings.TrimSpace(resourceURL)
	if strings.HasPrefix(resourceURL, "http://") || strings.HasPrefix(resourceURL, "https://") {
		return resourceURL
	}
	if strings.HasPrefix(resourceURL, "//") {
		return "https:" + resourceURL
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return resourceURL
	}
	relativeURL, err := url.Parse(resourceURL)
	if err != nil {
		return resourceURL
	}
	return baseURL.ResolveReference(relativeURL).String()
}
