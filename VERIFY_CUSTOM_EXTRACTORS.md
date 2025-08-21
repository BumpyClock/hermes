# Custom Extractor Verification Guide

This guide shows how to verify that custom extractors are being used when parsing URLs with the Go parser.

## Quick Verification

### 1. Check if extractor is being used

```bash
./bin/parser parse "https://www.npr.org/2025/08/20/nx-s1-5505987/trump-dc-takeover-timing-national-guard-police" | jq '.extractor_used'
```

**Expected output for NPR:** `"custom:www.npr.org"`
**Expected output for unknown domains:** `null` (generic extractor used)

### 2. Run comprehensive verification tool

```bash
go run tools/verify_parser_usage.go "https://www.npr.org/2025/08/20/nx-s1-5505987/trump-dc-takeover-timing-national-guard-police"
```

This will show:

- ✅ Custom extractor detection
- 📋 Configured selectors  
- 🔧 Which extractor was actually used
- 📊 Extraction quality assessment

### 3. Compare extraction results

**With Custom Extractor (NPR):**

```json
{
  "extractor_used": "custom:www.npr.org",
  "title": "How long can Trump's D.C. takeover last? Here's what to know",
  "author": "Rachel Treisman",
  "word_count": 1676
}
```

**With Generic Extractor (example.com):**

```json
{
  "extractor_used": null,
  "title": "Example Domain", 
  "author": "",
  "word_count": 24
}
```

## Available Custom Extractors

The parser includes 120+ custom extractors for major sites including:

### News Sites

- **NPR**: `www.npr.org` ✅
- **New York Times**: `www.nytimes.com` ✅  
- **CNN**: `www.cnn.com` ✅
- **Washington Post**: `www.washingtonpost.com` ✅
- **The Guardian**: `www.theguardian.com` ✅
- **Reuters**: `www.reuters.com` ✅

### Tech Sites  

- **GitHub**: `github.com` ✅
- **Hacker News**: `news.ycombinator.com` ✅
- **ArsTechnica**: `arstechnica.com` ✅
- **The Verge**: `www.theverge.com` ✅

### Social Platforms

- **Medium**: `medium.com` ✅
- **Reddit**: `www.reddit.com` ✅
- **Twitter**: `twitter.com` ✅

## How Custom Extractors Work

1. **Domain Matching**: Parser checks if a custom extractor exists for the domain
2. **Selector-Based Extraction**: Uses site-specific CSS selectors optimized for each site
3. **Fallback Support**: Falls back to generic extractors if custom extraction fails
4. **Quality Enhancement**: Custom extractors provide better title, author, content, and date extraction

## Example Verification

```bash
# Test NPR (has custom extractor)
./bin/parser parse "https://www.npr.org/2025/08/20/nx-s1-5505987/trump-dc-takeover-timing-national-guard-police" | jq '.extractor_used'
# Output: "custom:www.npr.org"

# Test unknown site (uses generic)  
./bin/parser parse "https://random-blog.example.com/post" | jq '.extractor_used'
# Output: null
```

## Adding New Custom Extractors

To add a custom extractor for a new site:

1. Create extractor file: `pkg/extractors/custom/www_newsite_com.go`
2. Define selectors for title, author, content, date
3. Register in `pkg/extractors/custom/index.go`
4. Test with verification tool

The parser will automatically use custom extractors when available, with generic fallback for unknown sites.
