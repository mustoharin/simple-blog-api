package sanitize

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	strictPolicy = bluemonday.StrictPolicy()
	ugcPolicy    = bluemonday.UGCPolicy()
)

// Trim removes leading and trailing whitespace from s.
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// SanitizeStrict strips all HTML tags from s using bluemonday's StrictPolicy.
// Use for all inputs except post content.
func SanitizeStrict(s string) string {
	return strictPolicy.Sanitize(s)
}

// SanitizeUGC sanitizes s using bluemonday's UGCPolicy, which allows safe
// formatting tags. Use only for post content (markdown/HTML body).
func SanitizeUGC(s string) string {
	return ugcPolicy.Sanitize(s)
}
