package utils

import (
	"html"

	"github.com/microcosm-cc/bluemonday"
)

var strictPolicy = bluemonday.StrictPolicy()

func StripHTML(s string) string {
	sanitized := strictPolicy.Sanitize(s)
	return html.UnescapeString(sanitized) // decode &#34; → "
}
