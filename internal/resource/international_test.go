package resource_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/resource"
)

func TestResource_InternationalContent_UTF8(t *testing.T) {
	r := resource.NewResource()

	// UTF-8 content with various international characters
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>International Test - Français</title>
</head>
<body>
	<h1>Café París 東京</h1>
	<p>This is a test with international characters:</p>
	<ul>
		<li>Spanish: ñáéíóú Ñ</li>
		<li>French: àâäéèêëïîôöùûüÿ</li>
		<li>German: äöüßÄÖÜ</li>
		<li>Japanese: こんにちは 東京 日本</li>
		<li>Russian: Привет мир</li>
		<li>Chinese: 你好世界</li>
		<li>Arabic: مرحبا بالعالم</li>
		<li>Emoji: 🌍🇫🇷🇩🇪🇯🇵🇷🇺🇨🇳</li>
	</ul>
</body>
</html>`

	doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, doc)

	// Verify title with international characters
	title := doc.Find("title").Text()
	assert.Equal(t, "International Test - Français", title)

	// Verify h1 with mixed scripts
	h1 := doc.Find("h1").Text()
	assert.Equal(t, "Café París 東京", h1)

	// Verify specific international characters are preserved
	content := doc.Find("body").Text()
	assert.Contains(t, content, "ñáéíóú")
	assert.Contains(t, content, "àâäéèêëïîôöùûüÿ")
	assert.Contains(t, content, "äöüßÄÖÜ")
	assert.Contains(t, content, "こんにちは")
	assert.Contains(t, content, "Привет мир")
	assert.Contains(t, content, "你好世界")
	assert.Contains(t, content, "مرحبا بالعالم")
	assert.Contains(t, content, "🌍🇫🇷🇩🇪🇯🇵🇷🇺🇨🇳")
}

func TestResource_InternationalContent_ISO88591(t *testing.T) {
	r := resource.NewResource()

	// HTML that declares ISO-8859-1 encoding
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="content-type" content="text/html; charset=iso-8859-1">
	<title>Test ISO-8859-1</title>
</head>
<body>
	<h1>Café en París</h1>
	<p>Characters: àáâãäåæçèéêë</p>
</body>
</html>`

	doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, doc)

	// Check that meta tag was normalized
	metaContent, exists := doc.Find("meta[http-equiv]").Attr("value")
	assert.True(t, exists)
	assert.Contains(t, metaContent, "iso-8859-1")

	title := doc.Find("title").Text()
	assert.Equal(t, "Test ISO-8859-1", title)
}

func TestResource_InternationalContent_Windows1251(t *testing.T) {
	r := resource.NewResource()

	// HTML with Windows-1251 (Cyrillic) encoding declaration
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="windows-1251">
	<title>Тест Windows-1251</title>
</head>
<body>
	<h1>Привет мир</h1>
	<p>Русские символы: абвгдеёжзийклмнопрстуфхцчшщъыьэюя</p>
	<p>АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ</p>
</body>
</html>`

	doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, doc)

	// Since we're passing as a string, it should work
	title := doc.Find("title").Text()
	assert.Contains(t, title, "1251") // At least part should be preserved
}

func TestResource_EncodingDetection_Various(t *testing.T) {
	tests := []struct {
		name        string
		charset     string
		content     string
		expectError bool
	}{
		{
			name:    "UTF-8",
			charset: "utf-8",
			content: "Hello 世界 🌍",
		},
		{
			name:    "ISO-8859-1",
			charset: "iso-8859-1",
			content: "Café París",
		},
		{
			name:    "Windows-1252",
			charset: "windows-1252",
			content: "Smart quotes test",
		},
		{
			name:    "Unknown charset",
			charset: "unknown-charset",
			content: "Fallback content",
		},
	}

	r := resource.NewResource()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="` + tt.charset + `">
	<title>Test</title>
</head>
<body>
	<p>` + tt.content + `</p>
</body>
</html>`

			doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, doc)

				// Should at least parse successfully
				title := doc.Find("title").Text()
				assert.Equal(t, "Test", title)
			}
		})
	}
}

func TestResource_LazyImages_International(t *testing.T) {
	r := resource.NewResource()

	// HTML with lazy images and international URLs
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Lazy Images Test</title>
</head>
<body>
	<img data-src="https://example.com/images/café.jpg" src="placeholder.gif" alt="Café image">
	<img data-lazy="https://example.jp/画像/写真.png" alt="Japanese image">
	<img data-original="https://example.ru/изображения/фото.jpg" alt="Russian image">
</body>
</html>`

	doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)

	// Check that lazy images with international URLs were processed
	images := doc.Find("img")
	assert.Equal(t, 3, images.Length())

	// First image should have data-src converted to src
	firstImg := images.First()
	src, exists := firstImg.Attr("src")
	assert.True(t, exists)
	assert.Equal(t, "https://example.com/images/café.jpg", src)
}

func TestResource_MetaTags_International(t *testing.T) {
	r := resource.NewResource()

	// HTML with international OpenGraph and meta tags
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta property="og:title" content="Café París - 東京レストラン">
	<meta property="og:description" content="Best café in París with Japanese influences">
	<meta name="description" content="Découvrez notre café unique à París">
	<meta property="article:author" content="Jean-François Müller">
</head>
<body>
	<h1>Content</h1>
</body>
</html>`

	doc, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)

	// Check that international content in meta tags is preserved
	ogTitle, exists := doc.Find("meta[name='og:title']").Attr("value")
	assert.True(t, exists)
	assert.Equal(t, "Café París - 東京レストラン", ogTitle)

	ogDesc, exists := doc.Find("meta[name='og:description']").Attr("value")
	assert.True(t, exists)
	assert.Contains(t, ogDesc, "París")

	author, exists := doc.Find("meta[name='article:author']").Attr("value")
	assert.True(t, exists)
	assert.Equal(t, "Jean-François Müller", author)
}

// Benchmark with international content.
func BenchmarkResource_InternationalContent(b *testing.B) {
	r := resource.NewResource()

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>International Benchmark - 国际基准测试</title>
</head>
<body>
	<h1>Performance Test with 国际字符 и Unicode símböls</h1>
	<p>Mixed content: English, 中文, Русский, Español, Français, Deutsch, العربية, 日本語</p>
	<div class="content">
		<p>Lorem ipsum with ñáéíóú characters</p>
		<p>Cyrillic: Привет мир! Как дела?</p>
		<p>Chinese: 你好世界！今天天气怎么样？</p>
		<p>Japanese: こんにちは世界！元気ですか？</p>
		<p>Arabic: مرحبا بالعالم! كيف حالك؟</p>
		<p>Emoji mix: 🌍🇫🇷🇩🇪🇯🇵🇷🇺🇨🇳🇪🇸🇸🇦</p>
	</div>
</body>
</html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Create(context.Background(), "http://example.com", htmlContent, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
