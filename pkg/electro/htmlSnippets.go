package electro

import (
	"fmt"
	"strings"
)

const snippetNoticeStartTemplate = `
<div class="notices [[notice_type]]">
	<div class="label"><i class="fa [[icon]]"></i>[[title]]</div>
</div>
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
		return "", fmt.Errorf("Unrecognized noticeType: %s", noticeType)
	}

	snippet := snippetNoticeStartTemplate
	snippet = strings.ReplaceAll(snippet, "[[notice_type]]", noticeType)
	snippet = strings.ReplaceAll(snippet, "[[icon]]", icon)
	capitalized := strings.ToUpper(noticeType[:1]) + noticeType[1:]
	snippet = strings.ReplaceAll(snippet, "[[title]]", capitalized)

	return snippet, nil
}
