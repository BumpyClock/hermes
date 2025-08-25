// ABOUTME: Character encoding detection from HTML content and headers
// ABOUTME: Faithful port of JavaScript getEncoding function with iconv-like validation

package text

import (
	"strings"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// GetEncoding extracts and validates character encoding from a string.
// This function is a faithful port of the JavaScript getEncoding function.
// It checks a string for encoding using ENCODING_RE pattern and validates
// the charset exists before returning it, otherwise returns DEFAULT_ENCODING.
func GetEncoding(str string) string {
	encoding := DEFAULT_ENCODING
	
	// Use ENCODING_RE to extract charset from string
	// Pattern: (?i)charset=['"]?([\w-]+)['"]?
	// Group 1: charset name
	matches := ENCODING_RE.FindStringSubmatch(str)
	if matches != nil && len(matches) > 1 {
		// Extract the charset (group 1 in our pattern)
		str = matches[1]  // Update str to the captured charset like in JS: [, str] = matches
	}
	
	// Check if the encoding exists (equivalent to iconv.encodingExists)
	// This checks either the extracted charset or the original string if no match
	if encodingExists(str) {
		encoding = str
	}
	
	return encoding
}

// encodingExists checks if a charset name is supported
// This replicates the behavior of iconv-lite's encodingExists function
func encodingExists(charset string) bool {
	if charset == "" {
		return false
	}
	
	return getEncodingByCharsetName(charset) != nil
}

// getEncodingByCharsetName returns encoding by charset name
// This includes comprehensive charset mapping similar to iconv-lite
func getEncodingByCharsetName(charset string) encoding.Encoding {
	charset = strings.ToLower(strings.TrimSpace(charset))
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
	case "utf-32", "utf32":
		return unicode.UTF8 // UTF-32 not directly available, fallback to UTF-8
		
	// ASCII
	case "ascii", "us-ascii":
		return unicode.UTF8 // ASCII is subset of UTF-8
		
	// Western European (ISO 8859 series)
	case "iso-8859-1", "latin1", "l1":
		return charmap.ISO8859_1
	case "iso-8859-2", "latin2", "l2":
		return charmap.ISO8859_2
	case "iso-8859-3", "latin3", "l3":
		return charmap.ISO8859_3
	case "iso-8859-4", "latin4", "l4":
		return charmap.ISO8859_4
	case "iso-8859-5":
		return charmap.ISO8859_5
	case "iso-8859-6":
		return charmap.ISO8859_6
	case "iso-8859-7":
		return charmap.ISO8859_7
	case "iso-8859-8":
		return charmap.ISO8859_8
	case "iso-8859-9", "latin5", "l5":
		return charmap.ISO8859_9
	case "iso-8859-10", "latin6", "l6":
		return charmap.ISO8859_10
	case "iso-8859-13", "latin7", "l7":
		return charmap.ISO8859_13
	case "iso-8859-14", "latin8", "l8":
		return charmap.ISO8859_14
	case "iso-8859-15", "latin9", "l9":
		return charmap.ISO8859_15
	case "iso-8859-16", "latin10", "l10":
		return charmap.ISO8859_16
		
	// Windows encodings (Code Pages)
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
		
	// IBM Code Pages
	case "cp437", "ibm437":
		return charmap.CodePage437
	case "cp850", "ibm850":
		return charmap.CodePage850
	case "cp852", "ibm852":
		return charmap.CodePage852
	case "cp855", "ibm855":
		return charmap.CodePage855
	case "cp858", "ibm858":
		return charmap.CodePage858
	case "cp860", "ibm860":
		return charmap.CodePage860
	case "cp862", "ibm862":
		return charmap.CodePage862
	case "cp863", "ibm863":
		return charmap.CodePage863
	case "cp865", "ibm865":
		return charmap.CodePage865
	case "cp866", "ibm866":
		return charmap.CodePage866
		
	// Japanese
	case "shift-jis", "shift_jis", "sjis", "shiftjis":
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
		
	// Mac encodings
	case "macintosh", "mac-roman":
		return charmap.Macintosh
		
	default:
		return nil
	}
}