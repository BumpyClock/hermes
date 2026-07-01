package generic

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func eachJSONLD(selection *goquery.Selection, visit func(map[string]interface{}) bool) {
	selection.Find(`script[type="application/ld+json"]`).EachWithBreak(func(_ int, script *goquery.Selection) bool {
		jsonText := strings.TrimSpace(script.Text())
		if jsonText == "" {
			return true
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
			return true
		}

		return visit(data)
	})
}
