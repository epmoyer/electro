package electro

import (
	"regexp"
	"strings"
)

func headingTextToId(text string) string {
	const NBSP = '\u00a0'
	originalText := text
	var id strings.Builder
	dashAppended := false

	// Replace &nbsp; and non-breaking space with regular space
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, string(NBSP), " ")

	// NOTE: This transformation is subtle. We need headingTextToID() to output identical
	//       id strings for both "raw" heading text (parsed from markdown text directly) and
	//       for heading text which has already been pre-processed by the markdown converter.
	//       The markdown converter changes '&' to '&amp;', so in order to make both ids
	//       identical we change it BACK to '&', which will then get DROPPED by the
	//       ID conversion.
	text = strings.ReplaceAll(text, "&amp;", "&")

	// Replace decimal with dashes so that heading numbers like "3.12" vs "31.2" remain unique.
	text = strings.ReplaceAll(text, ".", "-")

	// Convert to lowercase and iterate through characters
	text = strings.ToLower(text)

	for _, char := range text {
		if char == ' ' {
			if !dashAppended {
				id.WriteRune('-')
				dashAppended = true
			}
		} else if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			id.WriteRune(char)
			dashAppended = false
		}
	}

	// Combine multiple dashes into single dash
	result := id.String()
	dashRe := regexp.MustCompile(`-+`)
	result = dashRe.ReplaceAllString(result, "-")

	qlog.Debugf("headingTextToId() %q -> %q", originalText, result)

	return result
}
