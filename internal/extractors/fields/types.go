// ABOUTME: Extended field type definitions and extractors for categories, tags, related articles, and sentiment
// ABOUTME: Provides specialized extractors for advanced content metadata beyond basic text and URL fields

package fields

import (
	"fmt"
	"strings"
	"time"
)

// ExtendedFieldType represents different types of extended fields
type ExtendedFieldType string

const (
	FieldTypeCategory        ExtendedFieldType = "category"
	FieldTypeTags           ExtendedFieldType = "tags"
	FieldTypeRelatedArticles ExtendedFieldType = "related_articles"
	FieldTypeSentiment      ExtendedFieldType = "sentiment"
	FieldTypeReadingTime    ExtendedFieldType = "reading_time"
	FieldTypeLanguage       ExtendedFieldType = "language"
	FieldTypeKeywords       ExtendedFieldType = "keywords"
	FieldTypeEntities       ExtendedFieldType = "entities"
)

// CategoryField represents article categories
type CategoryField struct {
	Primary    string   `json:"primary"`
	Secondary  []string `json:"secondary,omitempty"`
	Confidence float64  `json:"confidence"`
}

// TagField represents article tags
type TagField struct {
	Name       string  `json:"name"`
	Weight     float64 `json:"weight"`
	Source     string  `json:"source"` // "extracted", "meta", "content"
	Normalized string  `json:"normalized"`
}

// RelatedArticle represents a related article
type RelatedArticle struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Excerpt     string    `json:"excerpt,omitempty"`
	PublishDate *time.Time `json:"publish_date,omitempty"`
	Similarity  float64   `json:"similarity"`
	Source      string    `json:"source"` // "internal", "external", "suggested"
}

// SentimentField represents content sentiment analysis
type SentimentField struct {
	Score     float64 `json:"score"`     // -1.0 to 1.0
	Label     string  `json:"label"`     // "positive", "negative", "neutral"
	Magnitude float64 `json:"magnitude"` // 0.0 to 1.0
	Confidence float64 `json:"confidence"`
}

// ReadingTimeField represents estimated reading time
type ReadingTimeField struct {
	Minutes    int     `json:"minutes"`
	Seconds    int     `json:"seconds"`
	WordCount  int     `json:"word_count"`
	ReadingWPM int     `json:"reading_wpm"` // Words per minute
	Confidence float64 `json:"confidence"`
}

// LanguageField represents detected language
type LanguageField struct {
	Code       string  `json:"code"`       // ISO 639-1 code (e.g., "en", "fr")
	Name       string  `json:"name"`       // Full language name
	Confidence float64 `json:"confidence"`
	Script     string  `json:"script,omitempty"` // Writing script (Latin, Cyrillic, etc.)
}

// KeywordField represents extracted keywords
type KeywordField struct {
	Term       string  `json:"term"`
	Frequency  int     `json:"frequency"`
	Weight     float64 `json:"weight"`
	Position   string  `json:"position"` // "title", "content", "meta"
	TFIDF      float64 `json:"tf_idf"`
}

// EntityField represents named entities
type EntityField struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`       // "PERSON", "ORGANIZATION", "LOCATION", etc.
	Confidence float64 `json:"confidence"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
	URL        string  `json:"url,omitempty"` // Knowledge base URL if available
}

// FieldExtractor interface for extended field extraction
type FieldExtractor interface {
	Extract(data interface{}) interface{}
	Type() ExtendedFieldType
	Name() string
	Confidence() float64
}

// BaseFieldExtractor provides common functionality
type BaseFieldExtractor struct {
	fieldType  ExtendedFieldType
	name       string
	confidence float64
}

// Type returns the field type
func (bfe *BaseFieldExtractor) Type() ExtendedFieldType {
	return bfe.fieldType
}

// Name returns the extractor name
func (bfe *BaseFieldExtractor) Name() string {
	return bfe.name
}

// Confidence returns the extraction confidence
func (bfe *BaseFieldExtractor) Confidence() float64 {
	return bfe.confidence
}

// CategoryExtractor extracts article categories
type CategoryExtractor struct {
	BaseFieldExtractor
	categoryMappings map[string]string
	keywordMappings  map[string][]string
}

// NewCategoryExtractor creates a new category extractor
func NewCategoryExtractor() *CategoryExtractor {
	return &CategoryExtractor{
		BaseFieldExtractor: BaseFieldExtractor{
			fieldType:  FieldTypeCategory,
			name:       "category_extractor",
			confidence: 0.8,
		},
		categoryMappings: map[string]string{
			"tech":       "Technology",
			"technology": "Technology",
			"science":    "Science",
			"news":       "News",
			"sports":     "Sports",
			"business":   "Business",
			"politics":   "Politics",
			"health":     "Health",
			"education":  "Education",
			"lifestyle":  "Lifestyle",
		},
		keywordMappings: map[string][]string{
			"Technology": {"programming", "software", "AI", "machine learning", "computer", "internet", "digital"},
			"Science":    {"research", "study", "experiment", "discovery", "scientific", "biology", "physics", "chemistry"},
			"News":       {"breaking", "report", "update", "announcement", "statement", "press release"},
			"Sports":     {"game", "match", "team", "player", "score", "tournament", "championship"},
			"Business":   {"company", "market", "finance", "economy", "stock", "investment", "corporate"},
		},
	}
}

// Extract extracts categories from various data sources
func (ce *CategoryExtractor) Extract(data interface{}) interface{} {
	categories := make([]string, 0)
	
	switch v := data.(type) {
	case []string:
		// Direct category list
		for _, cat := range v {
			if normalized := ce.normalizeCategory(cat); normalized != "" {
				categories = append(categories, normalized)
			}
		}
	case string:
		// Single category or content analysis
		if normalized := ce.normalizeCategory(v); normalized != "" {
			categories = append(categories, normalized)
		} else {
			// Analyze content for category keywords
			categories = ce.extractFromContent(v)
		}
	case map[string]interface{}:
		// Structured data with multiple sources
		if cats, ok := v["categories"].([]string); ok {
			for _, cat := range cats {
				if normalized := ce.normalizeCategory(cat); normalized != "" {
					categories = append(categories, normalized)
				}
			}
		}
		if content, ok := v["content"].(string); ok {
			categories = append(categories, ce.extractFromContent(content)...)
		}
	}
	
	if len(categories) == 0 {
		return CategoryField{Primary: "General", Confidence: 0.5}
	}
	
	// Return primary category and secondary categories
	primary := categories[0]
	secondary := categories[1:]
	
	return CategoryField{
		Primary:    primary,
		Secondary:  secondary,
		Confidence: ce.confidence,
	}
}

// normalizeCategory normalizes a category name
func (ce *CategoryExtractor) normalizeCategory(category string) string {
	lower := strings.ToLower(strings.TrimSpace(category))
	if normalized, exists := ce.categoryMappings[lower]; exists {
		return normalized
	}
	
	// Capitalize first letter of each word
	words := strings.Fields(lower)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	
	return strings.Join(words, " ")
}

// extractFromContent analyzes content to determine categories
func (ce *CategoryExtractor) extractFromContent(content string) []string {
	content = strings.ToLower(content)
	categoryScores := make(map[string]int)
	
	for category, keywords := range ce.keywordMappings {
		score := 0
		for _, keyword := range keywords {
			score += strings.Count(content, strings.ToLower(keyword))
		}
		if score > 0 {
			categoryScores[category] = score
		}
	}
	
	// Sort categories by score and return top matches
	var categories []string
	for category, score := range categoryScores {
		if score >= 2 { // Minimum threshold
			categories = append(categories, category)
		}
	}
	
	return categories
}

// TagsExtractor extracts and normalizes article tags
type TagsExtractor struct {
	BaseFieldExtractor
	stopWords map[string]bool
}

// NewTagsExtractor creates a new tags extractor
func NewTagsExtractor() *TagsExtractor {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "been": true, "by": true, "for": true, "from": true, "has": true,
		"he": true, "in": true, "is": true, "it": true, "its": true, "of": true,
		"on": true, "that": true, "the": true, "to": true, "was": true, "will": true,
		"with": true, "this": true, "these": true, "they": true, "we": true, "you": true,
	}
	
	return &TagsExtractor{
		BaseFieldExtractor: BaseFieldExtractor{
			fieldType:  FieldTypeTags,
			name:       "tags_extractor",
			confidence: 0.9,
		},
		stopWords: stopWords,
	}
}

// Extract extracts and normalizes tags
func (te *TagsExtractor) Extract(data interface{}) interface{} {
	var rawTags []string
	
	switch v := data.(type) {
	case []string:
		rawTags = v
	case string:
		// Split on common delimiters
		rawTags = te.splitTags(v)
	case map[string]interface{}:
		if tags, ok := v["tags"].([]string); ok {
			rawTags = tags
		}
	}
	
	var normalizedTags []string
	for _, tag := range rawTags {
		if normalized := te.normalizeTag(tag); normalized != "" {
			normalizedTags = append(normalizedTags, normalized)
		}
	}
	
	return normalizedTags
}

// normalizeTag normalizes a single tag
func (te *TagsExtractor) normalizeTag(tag string) string {
	// Trim and convert to lowercase
	tag = strings.TrimSpace(strings.ToLower(tag))
	
	// Skip if empty or too short
	if len(tag) < 2 {
		return ""
	}
	
	// Skip stop words
	if te.stopWords[tag] {
		return ""
	}
	
	// Convert spaces and underscores to hyphens
	tag = strings.ReplaceAll(tag, " ", "-")
	tag = strings.ReplaceAll(tag, "_", "-")
	
	// Remove special characters except hyphens
	var result strings.Builder
	for _, char := range tag {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// splitTags splits a tag string on common delimiters
func (te *TagsExtractor) splitTags(tagString string) []string {
	// Common tag delimiters
	delimiters := []string{",", ";", "|", "#"}
	
	tags := []string{tagString}
	for _, delimiter := range delimiters {
		var newTags []string
		for _, tag := range tags {
			splitTags := strings.Split(tag, delimiter)
			newTags = append(newTags, splitTags...)
		}
		tags = newTags
	}
	
	// Clean up tags
	var cleanTags []string
	for _, tag := range tags {
		if cleaned := strings.TrimSpace(tag); cleaned != "" {
			cleanTags = append(cleanTags, cleaned)
		}
	}
	
	return cleanTags
}

// RelatedArticlesExtractor extracts related articles
type RelatedArticlesExtractor struct {
	BaseFieldExtractor
}

// NewRelatedArticlesExtractor creates a new related articles extractor
func NewRelatedArticlesExtractor() *RelatedArticlesExtractor {
	return &RelatedArticlesExtractor{
		BaseFieldExtractor: BaseFieldExtractor{
			fieldType:  FieldTypeRelatedArticles,
			name:       "related_articles_extractor",
			confidence: 0.7,
		},
	}
}

// Extract extracts related articles from structured data
func (rae *RelatedArticlesExtractor) Extract(data interface{}) interface{} {
	var articles []RelatedArticle
	
	switch v := data.(type) {
	case []map[string]interface{}:
		for _, item := range v {
			if article := rae.parseArticle(item); article != nil {
				articles = append(articles, *article)
			}
		}
	case map[string]interface{}:
		if relatedData, ok := v["related"]; ok {
			if relatedList, ok := relatedData.([]map[string]interface{}); ok {
				for _, item := range relatedList {
					if article := rae.parseArticle(item); article != nil {
						articles = append(articles, *article)
					}
				}
			}
		}
	}
	
	return articles
}

// parseArticle parses a single article from structured data
func (rae *RelatedArticlesExtractor) parseArticle(data map[string]interface{}) *RelatedArticle {
	article := &RelatedArticle{
		Similarity: 0.5, // Default similarity
		Source:     "external",
	}
	
	if title, ok := data["title"].(string); ok {
		article.Title = title
	} else {
		return nil // Title is required
	}
	
	if url, ok := data["url"].(string); ok {
		article.URL = url
	} else {
		return nil // URL is required
	}
	
	if excerpt, ok := data["excerpt"].(string); ok {
		article.Excerpt = excerpt
	}
	
	if similarity, ok := data["similarity"].(float64); ok {
		article.Similarity = similarity
	}
	
	if source, ok := data["source"].(string); ok {
		article.Source = source
	}
	
	// Parse publish date if available
	if dateStr, ok := data["publish_date"].(string); ok {
		if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
			article.PublishDate = &date
		}
	}
	
	return article
}

// String returns a string representation of ExtendedFieldType
func (eft ExtendedFieldType) String() string {
	return string(eft)
}

// IsValidFieldType checks if a field type is valid
func IsValidFieldType(fieldType string) bool {
	validTypes := []ExtendedFieldType{
		FieldTypeCategory,
		FieldTypeTags,
		FieldTypeRelatedArticles,
		FieldTypeSentiment,
		FieldTypeReadingTime,
		FieldTypeLanguage,
		FieldTypeKeywords,
		FieldTypeEntities,
	}
	
	for _, validType := range validTypes {
		if fieldType == string(validType) {
			return true
		}
	}
	
	return false
}

// GetFieldTypeMetadata returns metadata for a field type
func GetFieldTypeMetadata(fieldType ExtendedFieldType) map[string]interface{} {
	metadata := map[ExtendedFieldType]map[string]interface{}{
		FieldTypeCategory: {
			"description": "Article categories and topics",
			"output_type": "CategoryField",
			"examples":    []string{"Technology", "Science", "News"},
		},
		FieldTypeTags: {
			"description": "Article tags and keywords",
			"output_type": "[]string",
			"examples":    []string{"web-development", "go-programming", "api-design"},
		},
		FieldTypeRelatedArticles: {
			"description": "Related articles and cross-references",
			"output_type": "[]RelatedArticle",
			"examples":    []string{"Similar articles", "Cross-references", "Suggested reading"},
		},
		FieldTypeSentiment: {
			"description": "Content sentiment analysis",
			"output_type": "SentimentField",
			"examples":    []string{"positive", "negative", "neutral"},
		},
		FieldTypeReadingTime: {
			"description": "Estimated reading time",
			"output_type": "ReadingTimeField",
			"examples":    []string{"5 minutes", "300 words", "Average reading speed"},
		},
		FieldTypeLanguage: {
			"description": "Detected content language",
			"output_type": "LanguageField",
			"examples":    []string{"en", "fr", "es", "de"},
		},
		FieldTypeKeywords: {
			"description": "Extracted keywords and key phrases",
			"output_type": "[]KeywordField",
			"examples":    []string{"machine learning", "artificial intelligence", "deep learning"},
		},
		FieldTypeEntities: {
			"description": "Named entities (people, places, organizations)",
			"output_type": "[]EntityField",
			"examples":    []string{"Apple Inc.", "San Francisco", "Tim Cook"},
		},
	}
	
	if meta, exists := metadata[fieldType]; exists {
		return meta
	}
	
	return map[string]interface{}{
		"description": fmt.Sprintf("Unknown field type: %s", fieldType),
		"output_type": "interface{}",
		"examples":    []string{},
	}
}