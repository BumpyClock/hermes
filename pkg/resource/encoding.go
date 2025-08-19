package resource

import (
	"bufio"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// DetectAndDecodeText detects encoding and converts to UTF-8
func DetectAndDecodeText(data []byte, contentType string) (string, error) {
	// First try to get encoding from content type
	if enc := getEncodingFromContentType(contentType); enc != nil {
		decoded, err := enc.NewDecoder().Bytes(data)
		if err == nil {
			return string(decoded), nil
		}
	}

	// Try to detect encoding automatically
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil || result.Confidence < 80 {
		// Fallback: assume UTF-8
		return string(data), nil
	}

	// Get encoder for detected charset
	enc := getEncodingByName(result.Charset)
	if enc == nil {
		// Fallback: assume UTF-8
		return string(data), nil
	}

	// Decode to UTF-8
	decoded, err := enc.NewDecoder().Bytes(data)
	if err != nil {
		// Fallback: assume UTF-8
		return string(data), nil
	}

	return string(decoded), nil
}

// getEncodingFromContentType extracts encoding from Content-Type header
func getEncodingFromContentType(contentType string) encoding.Encoding {
	if contentType == "" {
		return nil
	}

	// Look for charset parameter
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "charset=") {
			charset := strings.TrimPrefix(strings.ToLower(part), "charset=")
			charset = strings.Trim(charset, "\"'")
			return getEncodingByName(charset)
		}
	}

	return nil
}

// getEncodingFromHTML tries to extract encoding from HTML meta tags
func getEncodingFromHTML(data []byte) encoding.Encoding {
	// Look for charset in first 1KB of HTML
	searchData := data
	if len(searchData) > 1024 {
		searchData = data[:1024]
	}

	content := strings.ToLower(string(searchData))

	// Look for <meta charset="...">
	if idx := strings.Index(content, "charset="); idx != -1 {
		start := idx + 8
		end := start
		for end < len(content) && content[end] != '"' && content[end] != '\'' && content[end] != '>' && content[end] != ' ' {
			end++
		}
		if end > start {
			charset := content[start:end]
			return getEncodingByName(charset)
		}
	}

	return nil
}

// getEncodingByName returns encoding by charset name
func getEncodingByName(charset string) encoding.Encoding {
	charset = strings.ToLower(charset)
	charset = strings.ReplaceAll(charset, "_", "-")

	switch charset {
	// UTF encodings
	case "utf-8", "utf8":
		return unicode.UTF8
	case "utf-16", "utf16":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)

	// Western European
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1
	case "iso-8859-2", "latin2":
		return charmap.ISO8859_2
	case "iso-8859-3", "latin3":
		return charmap.ISO8859_3
	case "iso-8859-4", "latin4":
		return charmap.ISO8859_4
	case "iso-8859-5":
		return charmap.ISO8859_5
	case "iso-8859-6":
		return charmap.ISO8859_6
	case "iso-8859-7":
		return charmap.ISO8859_7
	case "iso-8859-8":
		return charmap.ISO8859_8
	case "iso-8859-9", "latin5":
		return charmap.ISO8859_9
	case "iso-8859-10", "latin6":
		return charmap.ISO8859_10
	case "iso-8859-13", "latin7":
		return charmap.ISO8859_13
	case "iso-8859-14", "latin8":
		return charmap.ISO8859_14
	case "iso-8859-15", "latin9":
		return charmap.ISO8859_15
	case "iso-8859-16", "latin10":
		return charmap.ISO8859_16

	// Windows encodings
	case "windows-1250", "cp1250":
		return charmap.Windows1250
	case "windows-1251", "cp1251":
		return charmap.Windows1251
	case "windows-1252", "cp1252":
		return charmap.Windows1252
	case "windows-1253", "cp1253":
		return charmap.Windows1253
	case "windows-1254", "cp1254":
		return charmap.Windows1254
	case "windows-1255", "cp1255":
		return charmap.Windows1255
	case "windows-1256", "cp1256":
		return charmap.Windows1256
	case "windows-1257", "cp1257":
		return charmap.Windows1257
	case "windows-1258", "cp1258":
		return charmap.Windows1258

	// Japanese
	case "shift-jis", "shift_jis", "sjis":
		return japanese.ShiftJIS
	case "euc-jp", "eucjp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP

	// Korean
	case "euc-kr", "euckr":
		return korean.EUCKR

	// Chinese
	case "gb2312", "gb-2312":
		return simplifiedchinese.GB18030 // GB2312 is subset of GB18030
	case "gbk":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5

	// Russian/Cyrillic
	case "koi8-r":
		return charmap.KOI8R
	case "koi8-u":
		return charmap.KOI8U

	default:
		return nil
	}
}

// IsTextContent checks if content type indicates text content
func IsTextContent(contentType string) bool {
	if contentType == "" {
		return false
	}

	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "application/xhtml") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "text/xml")
}

// GetEncodingFromMeta extracts encoding from HTML meta tags
// This matches the JavaScript getEncoding function behavior
func GetEncodingFromMeta(htmlContent string) encoding.Encoding {
	// First try HTML meta tag parsing
	if enc := getEncodingFromHTML([]byte(htmlContent)); enc != nil {
		return enc
	}
	
	// Fallback to default
	return unicode.UTF8
}

// GetEncodingByCharset returns encoding by charset name (public wrapper)
func GetEncodingByCharset(charset string) encoding.Encoding {
	return getEncodingByName(charset)
}

// NormalizeHTML performs basic HTML normalization
func NormalizeHTML(html string) string {
	// Convert line endings
	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\r", "\n")

	// Basic whitespace cleanup
	scanner := bufio.NewScanner(strings.NewReader(html))
	var result strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}