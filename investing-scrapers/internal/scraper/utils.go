package scraper

import (
	"regexp"
	"strings"
)

var stripHTMLRe = regexp.MustCompile(`<[^>]+>`)

func subStr(html string) string {
	return stripHTMLRe.ReplaceAllString(html, "")
}

func normNumber(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return strings.TrimSpace(s)
}
