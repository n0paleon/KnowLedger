package utils

import "github.com/microcosm-cc/bluemonday"

var strictPolicy = bluemonday.StrictPolicy()

func StripHTML(s string) string {
	return strictPolicy.Sanitize(s)
}
