package electro

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"

	// D2 diagram support
	goldmarkd2 "github.com/FurqanSoftware/goldmark-d2"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
)

type mdRendererT struct {
	Markdown                    string
	Filename                    string
	PathProjectDir              string
	PathOutputDir               string
	Substitutions               map[string]string
	DoStripFrontmatter          bool
	DoNumberHeadings            bool
	NumberHeadingsAtLevel       int
	DoWrangleInterdocumentLinks bool
	TocItems                    []tocItemT
}

type tocItemT struct {
	HeadingLevel  int
	HeadingNumber string
	HeadingText   string
}

func NewMdRenderer(markdown string, filename string, pathProjectDir string, pathOutputDir string) *mdRendererT {
	return &mdRendererT{
		Markdown:                    markdown,
		Filename:                    filename,
		PathProjectDir:              pathProjectDir,
		PathOutputDir:               pathOutputDir,
		Substitutions:               make(map[string]string),
		DoStripFrontmatter:          true,
		DoNumberHeadings:            false,
		NumberHeadingsAtLevel:       2,
		DoWrangleInterdocumentLinks: false,
		TocItems:                    []tocItemT{},
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
	pathMdPreParsed := path.Join(r.PathOutputDir, "md_pre_parsed", r.Filename)
	err = writeStringToFileEnsureDir(pathMdPreParsed, md)
	if err != nil {
		return "", fmt.Errorf("error writing pre-parsed markdown to %q: %w", pathMdPreParsed, err)
	}

	// -------------------------
	// Render markdown to HTML
	// -------------------------

	// Prepare D2 diagram layout and theme option
	themeID := d2themescatalog.CoolClassics.ID // take address of a local var (always addressable)

	layout := func(ctx context.Context, g *d2graph.Graph) error {
		// nil opts = defaults (or use &d2dagrelayout.ConfigurableOpts{} if you prefer explicit)
		return d2dagrelayout.Layout(ctx, g, nil)
	}

	var bufHtmlBytes bytes.Buffer
	mdConverter := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
			),
			&goldmarkd2.Extender{
				Layout:  layout,
				ThemeID: &themeID,
			},
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
	md = r.MdTightenBulletLists(md)

	// -------------------------
	// Parse headings (numbering, TOC data generation)
	// -------------------------
	md, err = r.MdParseHeadings(md, r.DoNumberHeadings)
	if err != nil {
		return "", fmt.Errorf("error numbering headings: %w", err)
	}

	// -------------------------
	// Parse notices
	// -------------------------
	md, err = r.MdParseNotices(md)
	if err != nil {
		return "", fmt.Errorf("error parsing notices: %w", err)
	}

	// -------------------------
	// Parse fields
	// -------------------------
	md, err = r.MdParseFields(md)
	if err != nil {
		return "", fmt.Errorf("error parsing fields: %w", err)
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
	// Parse CSV references (into tables)
	// -------------------------
	md = r.MdParseCsvReferences(md)

	// -------------------------
	// Generate Table of Contents
	// -------------------------
	md = r.MdGenerateTableOfContents(md)

	return md, nil
}

func (r *mdRendererT) MdGenerateTableOfContents(md string) string {
	tocIndent := strings.Repeat("&nbsp;", 4)
	toc_lines := []string{"", "# Table of Contents", "<div class=\"toc-body\">"}
	for _, tocItem := range r.TocItems {
		indent := strings.Repeat(tocIndent, tocItem.HeadingLevel-1)
		pageId := strings.TrimSuffix(r.Filename, filepath.Ext(r.Filename))
		toc_line := fmt.Sprintf(
			"%s<div class=\"toc-number toc{%d}\">%s</div>"+
				"<a class=\"toc-link\" href=\"?pageId=%s&headingId=%s\">"+
				tocItem.HeadingText+
				"</a><br>",
			indent,
			tocItem.HeadingLevel,
			tocItem.HeadingNumber,
			pageId,
			headingTextToId(tocItem.HeadingNumber+" "+tocItem.HeadingText))
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

func (r *mdRendererT) MdTightenBulletLists(md string) string {
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

func (r *mdRendererT) MdParseHeadings(md string, doNumberHeadings bool) (string, error) {
	pragmaNumberHeadingsRe := regexp.MustCompile(`@pragma\{number_headings:(?P<setting>\S+)\}`)
	pragmaNumberHeadingsEnabled := true
	pragmaIncludeInTocRe := regexp.MustCompile(`@pragma\{include_in_toc:(?P<setting>\S+)\}`)
	pragmaIncludeInTocEnabled := true
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
		//   @pragma{number_headings:true}
		//   # Document Info
		//   ...
		//   @pragma{number_headings:false}
		//   # Introduction
		if matches := pragmaNumberHeadingsRe.FindStringSubmatch(line); matches != nil {
			setting := matches[pragmaNumberHeadingsRe.SubexpIndex("setting")]
			if setting == "false" {
				pragmaNumberHeadingsEnabled = false
			} else if setting == "true" {
				pragmaNumberHeadingsEnabled = true
			}
			// We do not include the pragma line in the output.
			// By convention, pragmas are always on their own line.	Any other text
			// appearing on the pragma line will not be rendered.
			continue
		}
		if matches := pragmaIncludeInTocRe.FindStringSubmatch(line); matches != nil {
			setting := matches[pragmaIncludeInTocRe.SubexpIndex("setting")]
			if setting == "false" {
				pragmaIncludeInTocEnabled = false
			} else if setting == "true" {
				pragmaIncludeInTocEnabled = true
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
		if len(pieces) < 2 {
			// Malformed heading, skip
			renumberedLines = append(renumberedLines, line)
			continue
		}
		level := len(pieces[0])
		headingText := pieces[1]
		headingNumberText := ""

		if level >= r.NumberHeadingsAtLevel && pragmaNumberHeadingsEnabled && doNumberHeadings {
			// Add heading number
			headingNumberText = headingManager.GetNextHeadingNumber(level)
			idWithoutHeadingNumber := headingTextToId(headingText)
			idWithHeadingNumber := headingTextToId(headingNumberText + " " + headingText)
			headingIdToHeadingIdWithLineNumber[idWithoutHeadingNumber] = idWithHeadingNumber

			line = fmt.Sprintf(
				"%s %s&nbsp;&nbsp;&nbsp;&nbsp;%s",
				pieces[0],
				headingNumberText,
				headingText)
		}
		if pragmaIncludeInTocEnabled {
			r.TocItems = append(r.TocItems, tocItemT{
				HeadingLevel:  level,
				HeadingNumber: headingNumberText,
				HeadingText:   headingText,
			})
		}
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

	// -------------------------
	// Begin Notice: @block{<noticeType>}
	// -------------------------
	reNoticeStartNew := regexp.MustCompile(`@block\{(\S*)\}`)
	noticeTypesNew := reNoticeStartNew.FindAllStringSubmatch(md, -1)
	for _, match := range noticeTypesNew {
		// fmt.Printf("*** Notice New: %#v\n", match)
		// This is the full directive, e.g. "@block{info}"
		noticeDirective := match[0]
		// This is the type of notice, e.g. "info"
		noticeType := match[1]

		// Skip @block{end} as it's handled separately below
		if noticeType == "end" {
			continue
		}

		htmlNoticeStart, err := buildHtmlSnippetNoticeStart(noticeType)
		if err != nil {
			return "", fmt.Errorf("error building notice snippet for type %q: %w", noticeType, err)
		}
		sub := r.CreateSubstitution(htmlNoticeStart)
		md = strings.ReplaceAll(md, noticeDirective, sub+"\n")
	}

	// -------------------------
	// Begin Notice: LEGACY: {{% notice <noticeType> %}}
	// -------------------------
	reNoticeStart := regexp.MustCompile(`{{% notice (\S*) %}}`)
	noticeTypes := reNoticeStart.FindAllStringSubmatch(md, -1)
	for _, match := range noticeTypes {
		// fmt.Printf("*** Notice Legacy: %#v\n", match)
		// This is the full directive, e.g. "{{% notice info %}}"
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
		// e.g. Without it inline code blocks in the first line after a notice start were
		// not rendered if the source contained no blank line between the notice start and
		// the notice content.
		md = strings.ReplaceAll(md, noticeDirective, sub+"\n")
	}

	// -------------------------
	// End Notice: @block{end}
	// -------------------------
	noticeEndDirectiveNew := "@block{end}"
	if strings.Contains(md, noticeEndDirectiveNew) {
		sub := r.CreateSubstitution(snippetHtmlNoticeEnd)
		md = strings.ReplaceAll(md, noticeEndDirectiveNew, sub)
	}

	// -------------------------
	// End Notice: LEGACY: {{% /notice %}}
	// -------------------------
	noticeEndDirectiveLegacy := "{{% /notice %}}"
	if strings.Contains(md, noticeEndDirectiveLegacy) {
		sub := r.CreateSubstitution(snippetHtmlNoticeEnd)
		md = strings.ReplaceAll(md, noticeEndDirectiveLegacy, sub)
	}

	return md, nil
}

// MdParseCsvReferences parses CSV references in the markdown and replaces them with table markup.
// CSV references have the form `@table{attachments/<csv_filename>}`.
// The first row of the CSV is treated as the table header.
// The remaining rows are treated as table data.
// Nelwlines are converted to "<br>" tags.
// The Markdown table format is not padded to be pretty; only to be syntactically correct.
func (r *mdRendererT) MdParseCsvReferences(md string) string {
	qlog.Trace()
	reTable := regexp.MustCompile(`@table\{(attachments/[^}]+)\}`)
	matches := reTable.FindAllStringSubmatch(md, -1)
	for _, match := range matches {
		tableDirective := match[0]
		csvRelativePath := match[1]
		// NOTE: The attachments dir has not yet been copied to the output dir at the time
		// when we render the markdown to HTML, so we need to get the CSV file from the project dir.
		csvAbsolutePath := path.Join(r.PathProjectDir, "docs", csvRelativePath)

		data, err := os.ReadFile(csvAbsolutePath)
		if err != nil {
			qlog.Debugf("Could not read CSV table reference %q: %v", csvAbsolutePath, err)
			continue
		}

		reader := csv.NewReader(strings.NewReader(string(data)))
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		if err != nil {
			qlog.Debugf("Could not parse CSV table reference %q: %v", csvAbsolutePath, err)
			continue
		}
		if len(records) == 0 {
			qlog.Debugf("CSV table reference %q is empty", csvAbsolutePath)
			continue
		}

		columnCount := 0
		for _, record := range records {
			if len(record) > columnCount {
				columnCount = len(record)
			}
		}
		if columnCount == 0 {
			qlog.Debugf("CSV table reference %q has no columns", csvAbsolutePath)
			continue
		}

		sanitizeCell := func(cell string) string {
			cell = strings.ReplaceAll(cell, "\r\n", "\n")
			cell = strings.ReplaceAll(cell, "\r", "\n")
			cell = strings.ReplaceAll(cell, "\n", "<br>")
			cell = strings.ReplaceAll(cell, "|", `\|`)
			// Today we "brute force" cells containing "n/a" to use the gray background.
			// FIXME: make this association a @pragma so that we can assign background colors in the
			// markdown source to cells containing specific values.
			if strings.Contains(strings.ToLower(cell), "n/a") {
				cell += `<span class="td_bg_gray"></span>`
			}
			return cell
		}
		padRecord := func(record []string) []string {
			if len(record) >= columnCount {
				return record
			}
			out := make([]string, columnCount)
			copy(out, record)
			return out
		}

		header := padRecord(records[0])
		var b strings.Builder
		b.WriteString("\n")
		b.WriteString("|")
		for _, cell := range header {
			b.WriteString(sanitizeCell(cell))
			b.WriteString("|")
		}
		b.WriteString("\n|")
		for i := 0; i < columnCount; i++ {
			b.WriteString("---|")
		}

		for _, record := range records[1:] {
			b.WriteString("\n|")
			for _, cell := range padRecord(record) {
				b.WriteString(sanitizeCell(cell))
				b.WriteString("|")
			}
		}
		b.WriteString("\n")

		md = strings.ReplaceAll(md, tableDirective, b.String())
	}

	return md
}

func (r *mdRendererT) MdParseFields(md string) (string, error) {
	// Substitute field directives with field text

	// @field{<field_name>}
	reField := regexp.MustCompile(`@field\{(\S*?)\}`)
	fieldDirectives := reField.FindAllStringSubmatch(md, -1)
	for _, match := range fieldDirectives {
		// This is the full directive, e.g. "@field{username}"
		fieldDirective := match[0]
		// This is the name of the field, e.g. "username"
		fieldName := match[1]

		fieldText, ok := fieldManagerGetFieldText(fieldName)
		if !ok {
			return "", fmt.Errorf("field %q not found", fieldName)
		}

		md = strings.ReplaceAll(md, fieldDirective, fieldText)
	}

	return md, nil
}

func fieldManagerGetFieldText(fieldName string) (string, bool) {
	if fieldName == "app_version" {
		return config.Version, true
	}
	if fieldName == "app_name" {
		return config.AppName, true
	}
	if fieldName == "datetime_now" {
		// Form: 2026-01-31T08:50:07-08:00
		now := time.Now()
		formatted := now.Format(time.RFC3339)
		return formatted, true
	}
	return "", false
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
	// -------------------------
	// Apply substitutions
	// -------------------------
	for placeholder, final := range r.Substitutions {
		html = strings.ReplaceAll(html, placeholder, final)
	}

	// -------------------------
	// Process pragmas
	// -------------------------
	pragmaInjectHeadingClassRe := regexp.MustCompile(`@pragma\{inject_heading_class:(?P<setting>\S*)\}`)
	pragmaInjectHeadingClass := ""
	reHeading := regexp.MustCompile(`<(h\d)(\s[^>]*)?>`)
	lines := strings.Split(html, "\n")
	linesOut := []string{}
	for _, line := range lines {
		if matches := pragmaInjectHeadingClassRe.FindStringSubmatch(line); matches != nil {
			setting := matches[pragmaInjectHeadingClassRe.SubexpIndex("setting")]
			pragmaInjectHeadingClass = setting
			// We do not include the pragma line in the output.
			// By convention, pragmas are always on their own line.	Any other text
			// appearing on the pragma line will not be rendered.
			continue
		}
		if pragmaInjectHeadingClass != "" {
			// Inject class into headings
			if reHeading.MatchString(line) {
				line = reHeading.ReplaceAllString(line, fmt.Sprintf("<$1 class=\"%s\"$2>", pragmaInjectHeadingClass))
			}
		}
		linesOut = append(linesOut, line)
	}
	html = strings.Join(linesOut, "\n")
	return html, nil
}
