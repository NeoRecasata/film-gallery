package slug

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s-]`)
	whitespace      = regexp.MustCompile(`\s+`)
	multiDash       = regexp.MustCompile(`-{2,}`)
)

func Generate(title, fallback string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphanumeric.ReplaceAllString(s, "")
	s = whitespace.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return fallback
	}
	return s
}
