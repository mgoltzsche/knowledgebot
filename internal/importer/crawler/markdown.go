package crawler

import (
	"regexp"
	"strings"
)

var markdownLinkRegex = regexp.MustCompile(`!?\[([^]]*)\]\([^)]+\)`)
var whitespaceRegex = regexp.MustCompile(` +`)

func stripMarkdownLinks(markdown string) string {
	markdown = markdownLinkRegex.ReplaceAllString(markdown, " $1 ")
	markdown = whitespaceRegex.ReplaceAllString(markdown, " ")
	markdown = strings.TrimSpace(markdown)

	return markdown
}
