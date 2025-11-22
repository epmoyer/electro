package electro

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

type mdRendererT struct {
	Markdown                    string
	Filename                    string
	PathOutputDir               string
	Substitutions               map[string]string
	DoStripFrontmatter          bool
	DoNumberHeadings            bool
	NumberHeadingsAtLevel       int
	DoWrangleInterdocumentLinks bool
	TocLines                    []string
}

func NewMdRenderer(markdown string, filename string, pathOutputDir string) *mdRendererT {
	return &mdRendererT{
		Markdown:                    markdown,
		Filename:                    filename,
		PathOutputDir:               pathOutputDir,
		Substitutions:               make(map[string]string),
		DoStripFrontmatter:          true,
		DoNumberHeadings:            false,
		NumberHeadingsAtLevel:       2,
		DoWrangleInterdocumentLinks: false,
		TocLines:                    []string{},
	}
}

func (r *mdRendererT) Render() (string, error) {
	md := r.Markdown

	// -------------------------
	// Pre-parser
	// -------------------------
	md, err := r.PreParseMarkdown(md)
	if err != nil {
		return "", fmt.Errorf("error pre-parsing markdown content: %w", err)
	}
	pathMdPreParsed := filepath.Join(r.PathOutputDir, "md_pre_parsed", r.Filename)
	err = writeStringToFileEnsureDir(pathMdPreParsed, md)
	if err != nil {
		return "", fmt.Errorf("error writing pre-parsed markdown to %q: %w", pathMdPreParsed, err)
	}

	// -------------------------
	// Render markdown to HTML
	// -------------------------
	var bufHtmlBytes bytes.Buffer
	mdConverter := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
			),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
		),
	)
	err = mdConverter.Convert([]byte(md), &bufHtmlBytes)
	if err != nil {
		return "", fmt.Errorf("error converting markdown to HTML: %w", err)
	}
	html := bufHtmlBytes.String()

	// -------------------------
	// Post-parser
	// -------------------------
	html, err = r.PostParseHtml(html)
	if err != nil {
		return "", fmt.Errorf("error post-processing HTML: %w", err)
	}

	return html, nil
}

func (r *mdRendererT) PreParseMarkdown(md string) (string, error) {
	var err error

	// -------------------------
	// Strip frontmatter
	// -------------------------
	if r.DoStripFrontmatter {
		md, err = r.stripFrontmatter(md)
		if err != nil {
			return "", fmt.Errorf("error stripping frontmatter: %w", err)
		}
	}

	// -------------------------
	// Tighten bullet lists
	// -------------------------
	md = r.MdTightenlBulletLists(md)

	// -------------------------
	// Number headings
	// -------------------------
	if r.DoNumberHeadings {
		md, err = r.MdAddHeadingNumbers(md)
		if err != nil {
			return "", fmt.Errorf("error numbering headings: %w", err)
		}
	}

	// -------------------------
	// Parse notices
	// -------------------------
	md, err = r.MdParseNotices(md)
	if err != nil {
		return "", fmt.Errorf("error parsing notices: %w", err)
	}

	// -------------------------
	// Parse checklists
	// -------------------------
	md = r.MdParseChecklists(md)

	// -------------------------
	// Parse interdocument links
	// -------------------------
	if r.DoWrangleInterdocumentLinks {
		md = r.MdWrangleInterDocumentLinks(md)
	}

	// -------------------------
	// Generate Table of Contents
	// -------------------------
	md = r.MdGenerateTableOfContents(md)

	return md, nil
}

func (r *mdRendererT) MdGenerateTableOfContents(md string) string {
	tocIndent := strings.Repeat("&nbsp;", 4)
	lines := strings.Split(md, "\n")
	toc_lines := []string{"", "# Table of Contents", "<div class=\"toc-body\">"}
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			continue
		}
		pieces := strings.SplitN(line, " ", 2)
		if len(pieces) < 2 {
			// Malformed heading, skip
			continue
		}
		level := strings.Count(pieces[0], "#")
		headingText := strings.TrimSpace(pieces[1])
		headingText = strings.ReplaceAll(headingText, "&nbsp;", " ")
		headingNumberText := "0"
		indent := strings.Repeat(tocIndent, level-1)
		pageId := "unassigned"
		// toc_line := fmt.Sprintf("%s- [%s](#%s)", indent, headingText, headingTextToId(headingText))
		toc_line := fmt.Sprintf(
			"%s<div class=\"toc-number toc{%d}\">%s</div>"+
				"<a class=\"toc-link\" href=\"?pageId={%s}&headingId={%s}\">"+
				headingText+
				"</a><br>",
			indent,
			level,
			headingNumberText,
			pageId,
			headingTextToId(headingText))
		toc_lines = append(toc_lines, toc_line)
	}
	toc_lines = append(toc_lines, "</div>")
	md = strings.ReplaceAll(md, "{{% table_of_contents %}}", strings.Join(toc_lines, "\n"))
	return md
}

func (r *mdRendererT) stripFrontmatter(md string) (string, error) {
	lines := strings.Split(md, "\n")

	// Frontmatter must start on the first line
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return md, nil // No frontmatter
	}

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			// Found end of frontmatter, return everything after
			if i+1 < len(lines) {
				return strings.Join(lines[i+1:], "\n"), nil
			}
			return "", nil // File was only frontmatter
		}
	}

	return md, fmt.Errorf("frontmatter start found but no closing '---'")
}

func (r *mdRendererT) MdTightenlBulletLists(md string) string {
	// Remove blank lines between bullet list items
	lines := strings.Split(md, "\n")
	bulletRe := regexp.MustCompile(`^(\s*[-*+]\s+)`)

	var tightened []string
	pendingBlank := false
	previousNonBlankWasBullet := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Blank line
			pendingBlank = true
			continue
		}

		// Non-blank line
		isBullet := bulletRe.MatchString(line)
		if pendingBlank && !(isBullet && previousNonBlankWasBullet) {
			tightened = append(tightened, "")
		}
		pendingBlank = false
		previousNonBlankWasBullet = isBullet
		tightened = append(tightened, line)
	}

	return strings.Join(tightened, "\n")
}

func (r *mdRendererT) MdAddHeadingNumbers(md string) (string, error) {
	pragmaNumberHeadingsRe := regexp.MustCompile(`@pragma\{number_headings:(?P<setting>\S+)\}`)
	pragmaNumberHeadingsEnabled := true
	headingManager := newHeadingManager(r.NumberHeadingsAtLevel)
	headingIdToHeadingIdWithLineNumber := make(map[string]string)
	renumberedLines := []string{}
	inFencedBlock := false
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		// Detect pragmas
		// NOTE: We allow heading numbering to be "manually" toggled within the document so that
		// the user disable numbering for certain sections. Typically this is used for
		// opening sections like "Document Info", "Revision History", "Table of Contents", etc.
		// Example:
		//   @pragma{number_headings:off}
		//   # Document Info
		//   ...
		//   @pragma{number_headings:on}
		//   # Introduction
		if matches := pragmaNumberHeadingsRe.FindStringSubmatch(line); matches != nil {
			setting := matches[pragmaNumberHeadingsRe.SubexpIndex("setting")]
			if setting == "off" {
				pragmaNumberHeadingsEnabled = false
			} else if setting == "on" {
				pragmaNumberHeadingsEnabled = true
			}
			// We do not include the pragma line in the output.
			// By convention, pragmas are always on their own line.	Any other text
			// appearing on the pragma line will not be rendered.
			continue
		}

		// Detect fenced code blocks
		if strings.HasPrefix(line, "```") {
			inFencedBlock = !inFencedBlock
		}
		if inFencedBlock || !strings.HasPrefix(line, "#") {
			renumberedLines = append(renumberedLines, line)
			continue
		}
		// This is a heading line, and is not in a fenced code block

		// Split into heading markdown prefix (###...) and heading text
		pieces := strings.SplitN(line, " ", 2)
		level := len(pieces[0])
		headingText := pieces[1]

		if level < r.NumberHeadingsAtLevel || !pragmaNumberHeadingsEnabled {
			// Do not number this heading
			renumberedLines = append(renumberedLines, line)
			continue
		}

		headingNumberText := headingManager.GetNextHeadingNumber(level)
		idWithoutHeadingNumber := headingTextToId(headingText)
		idWithHeadingNumber := headingTextToId(headingNumberText + " " + headingText)
		headingIdToHeadingIdWithLineNumber[idWithoutHeadingNumber] = idWithHeadingNumber

		line = fmt.Sprintf(
			"%s %s&nbsp;&nbsp;&nbsp;&nbsp;%s",
			pieces[0],
			headingNumberText,
			headingText)
		renumberedLines = append(renumberedLines, line)
	}

	// -------------------------
	// Replace heading links to match re-numbered headings
	// -------------------------
	qlog.Debug("---- replacing links ----")
	outLines := []string{}
	for _, line := range renumberedLines {
		reHeadingLink := regexp.MustCompile(`\[.*?\]\(#.*?\)`)
		mdLinks := reHeadingLink.FindAllString(line, -1)
		if mdLinks == nil {
			outLines = append(outLines, line)
			continue
		}
		qlog.Debugf("LINE: %q", line)
		for _, mdLink := range mdLinks {
			reReference := regexp.MustCompile(`\[.*?\]\(#(?P<reference>.*?)\)`)
			match := reReference.FindStringSubmatch(mdLink)
			reference := match[reReference.SubexpIndex("reference")]
			idWiththHeadingNumber, ok := headingIdToHeadingIdWithLineNumber[reference]
			if !ok {
				qlog.Debug("    🔴 Mapping not found")
				continue
			}
			newReference := fmt.Sprintf("#%s", idWiththHeadingNumber)
			qlog.Debugf("    REPLACEMENT: %q -> %q", mdLink, newReference)
			newMdLink := strings.ReplaceAll(mdLink, "#"+reference, newReference)
			qlog.Debugf("    NEW MD LINK: %q", newMdLink)
			line = strings.ReplaceAll(line, mdLink, newMdLink)
		}
		qlog.Debugf("    NEW LINE: %q", line)
		outLines = append(outLines, line)
	}

	return strings.Join(outLines, "\n"), nil
}

func (r *mdRendererT) MdParseNotices(md string) (string, error) {
	// Parse custom notice blocks

	// FIXME: Find a way to ignore notices inside fenced code blocks.
	// This only really happens when we write notes about our own markdown syntax

	reNoticeStart := regexp.MustCompile(`{{% notice (\S*) %}}`)
	noticeTypes := reNoticeStart.FindAllStringSubmatch(md, -1)
	for _, match := range noticeTypes {
		// fmt.Printf("*** Notice: %#v\n", match)
		// This is the full directice, e.g. "{{% notice info %}}"
		noticeDirective := match[0]
		// This is the type of notice, e.g. "info"
		noticeType := match[1]
		htmlNoticeStart, err := buildHtmlSnippetNoticeStart(noticeType)
		if err != nil {
			return "", fmt.Errorf("error building notice snippet for type %q: %w", noticeType, err)
		}
		sub := r.CreateSubstitution(htmlNoticeStart)
		// NOTE: We need to force an extra newline after the substitution to ensure
		// that the markdown parser treats the first contiguous lines after the html notice start
		// substitution as markdown.
		// e.g. Without it inline code blocks in the first line after a notice start wewe
		// not rendered if the source contained no blank line between the notice start and
		// the notice content.
		md = strings.ReplaceAll(md, noticeDirective, sub+"\n")
	}

	noticeEndDirective := "{{% /notice %}}"
	if strings.Contains(md, noticeEndDirective) {
		sub := r.CreateSubstitution(snippetHtmlNoticeEnd)
		md = strings.ReplaceAll(md, noticeEndDirective, sub)
	}

	return md, nil
}

func (r *mdRendererT) MdParseChecklists(md string) string {
	outLines := []string{}
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ] ") {
			line = strings.Replace(line, "[ ] ", "🔲&nbsp;&nbsp;", 1)
		} else if strings.HasPrefix(trimmed, "- [x] ") {
			line = strings.Replace(line, "[x] ", "✅&nbsp;&nbsp;", 1)
		} else if strings.HasPrefix(trimmed, "- [X] ") {
			line = strings.Replace(line, "[X] ", "✅&nbsp;&nbsp;", 1)
		}
		outLines = append(outLines, line)
	}
	return strings.Join(outLines, "\n")
}

func (r *mdRendererT) MdWrangleInterDocumentLinks(md string) string {
	qlog.Debug("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-")
	qlog.Trace()
	linesOut := []string{}
	mdDocLinksRe := regexp.MustCompile(`\[.*?\]\(.*?\.md\)`)
	mdDocLinkPgRe := regexp.MustCompile(`\[.*?\]\((?P<page_id>.*?).md\)`)
	mdHeadingLinksRe := regexp.MustCompile(`\[.*?\]\(.*?\.md#.*?\)`)
	mdHeadingLinksPgRe := regexp.MustCompile(`\[.*?\]\((?P<page_id>.*?).md#(?P<heading_id>.*?)\)`)
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		// qlog.Debugf("MD line: %s", line)s
		lineOriginal := line

		// -----------------
		// Links to .md documents
		// -----------------
		mdDocLinks := mdDocLinksRe.FindAllString(line, -1)
		if len(mdDocLinks) > 0 {
			qlog.Debugf("    MD_LINKS (to .md doc): %#v", mdDocLinks)
		}
		for _, mdDocLink := range mdDocLinks {
			results := mdDocLinkPgRe.FindAllStringSubmatch(mdDocLink, -1)
			pageId := results[0][1]
			newReference := fmt.Sprintf("?pageId=%s", pageId)
			qlog.Debugf("    REPLACEMENT: %s -> %s", mdDocLink, newReference)
			newMdDocLink := strings.Replace(
				mdDocLink, fmt.Sprintf("%s.md", pageId), newReference, 1)
			qlog.Debugf("    NEW MD LINK: %s", newMdDocLink)
			line = strings.Replace(line, mdDocLink, newMdDocLink, -1)
		}

		// -----------------
		// Links to headings within .md documents
		// -----------------
		mdHeadingLinks := mdHeadingLinksRe.FindAllString(line, -1)
		if len(mdHeadingLinks) > 0 {
			qlog.Debugf("    MD_LINKS (to heading): %#v", mdHeadingLinks)
		}
		for _, mdHeadingtLink := range mdHeadingLinks {
			results := mdHeadingLinksPgRe.FindAllStringSubmatch(mdHeadingtLink, -1)
			pageId := results[0][1]
			headingId := results[0][2]
			newReference := fmt.Sprintf("?pageId=%s&amp;headingId=%s", pageId, headingId)
			qlog.Debugf("    REPLACEMENT: %s -> %s", mdHeadingtLink, newReference)
			newMdHeadingLink := strings.Replace(
				mdHeadingtLink, fmt.Sprintf("%s.md#%s", pageId, headingId), newReference, 1)
			qlog.Debugf("    NEW MD LINK: %s", newMdHeadingLink)
			line = strings.Replace(line, mdHeadingtLink, newMdHeadingLink, -1)
		}
		if line != lineOriginal {
			qlog.Debugf("    NEW LINE: %s", line)
		}
		linesOut = append(linesOut, line)
	}
	qlog.Debug("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-")
	return strings.Join(linesOut, "\n")
}

func (r *mdRendererT) CreateSubstitution(final string) string {
	// Create a substitution entry and return the placeholder
	placeholder := fmt.Sprintf(
		"<div class=\"PRE-PARSER-SUBSTITUTION-%d\"></div>", len(r.Substitutions)+1)
	r.Substitutions[placeholder] = final
	return placeholder
}

func (r *mdRendererT) PostParseHtml(html string) (string, error) {
	for placeholder, final := range r.Substitutions {
		html = strings.ReplaceAll(html, placeholder, final)
	}
	return html, nil
}
