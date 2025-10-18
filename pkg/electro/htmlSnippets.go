package electro

import (
	"fmt"
	"strings"
)

// NOTE: This snippet intentionally has no ending </div> because we insert that separately
// as a substitution for the notice end directive.
const snippetHtmlNoticeStartTemplate = `
<div class="notices [[notice_type]]">
	<div class="label"><i class="fa [[icon]]"></i>[[title]]</div>
`

const snippetHtmlNoticeEnd = "</div>"

var noticeIcons = map[string]string{
	"note":    "fa-exclamation-circle",
	"info":    "fa-info-circle",
	"tip":     "fa-wrench",
	"warning": "fa-exclamation-triangle",
}

func buildHtmlSnippetNoticeStart(noticeType string) (string, error) {
	var icon string
	var ok bool

	icon, ok = noticeIcons[noticeType]
	if !ok {
		return "", fmt.Errorf("unrecognized noticeType: %s", noticeType)
	}

	snippet := snippetHtmlNoticeStartTemplate
	snippet = strings.ReplaceAll(snippet, "[[notice_type]]", noticeType)
	snippet = strings.ReplaceAll(snippet, "[[icon]]", icon)
	capitalized := strings.ToUpper(noticeType[:1]) + noticeType[1:]
	snippet = strings.ReplaceAll(snippet, "[[title]]", capitalized)

	return snippet, nil
}
