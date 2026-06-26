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
	"sort"
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
	PathMdSourceDir             string
	PathOutputDir               string
	Substitutions               map[string]string
	DoStripFrontmatter          bool
	DoNumberHeadings            bool
	NumberHeadingsAtLevel       int
	DoWrangleInterdocumentLinks bool
	TocItems                    []tocItemT
	headingTargets              map[string]headingTargetT
}

type tocItemT struct {
	HeadingLevel  int
	HeadingNumber string
	HeadingText   string
}

// headingTargetT holds the information needed to turn a wiki-style link
// ("[[Heading Name]]") into a markdown link to the referenced heading.
type headingTargetT struct {
	// displayText is the actual heading text, as authored, so that a mistyped
	// wiki link adopts the heading's correct capitalization in the output.
	displayText string
	// id is the final HTML id of the heading (including any heading number).
	id string
}

type TableCellBgColorDescriptorT struct {
	id             int
	bgColor        string
	isCssRendered  bool
	isPartialMatch bool
}

func NewMdRenderer(markdown string, filename string, pathProjectDir string, pathOutputDir string) *mdRendererT {
	return &mdRendererT{
		Markdown:                    markdown,
		Filename:                    filename,
		PathMdSourceDir:             pathProjectDir,
		PathOutputDir:               pathOutputDir,
		Substitutions:               make(map[string]string),
		DoStripFrontmatter:          true,
		DoNumberHeadings:            false,
		NumberHeadingsAtLevel:       2,
		DoWrangleInterdocumentLinks: false,
		TocItems:                    []tocItemT{},
		headingTargets:              make(map[string]headingTargetT),
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
	// Parse wiki-style links ([[Heading Name]])
	// -------------------------
	md = r.MdWrangleWikiLinks(md)

	// -------------------------
	// Parse CSV references (into tables)
	// -------------------------
	md = r.MdParseCsvReferences(md)

	// -------------------------
	// Parse as-run inline delimiters
	// -------------------------
	md = r.MdParseAsRunInline(md)

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
		tocHeadingNumber := ""
		tocHeadingText := headingText
		// renderedHeadingText is the heading text as it will appear in the
		// output (including any heading number). For unnumbered headings it is
		// simply the authored text.
		renderedHeadingText := headingText

		if level >= r.NumberHeadingsAtLevel && pragmaNumberHeadingsEnabled && doNumberHeadings {
			// Add heading number. Appendices are only honoured at the level the
			// heading manager numbers from (atLevel), so the heading number is a
			// letter ("A") only when the heading is at that level.
			isAppendix := strings.HasPrefix(strings.ToLower(headingText), "appendix")
			appendixHonored := isAppendix && level == r.NumberHeadingsAtLevel
			headingNumberText = headingManager.GetNextHeadingNumber(level, isAppendix)
			idWithoutHeadingNumber := headingTextToId(headingText)

			if appendixHonored {
				// Standardize the appendix word's capitalization and move the
				// letter after it, e.g. "# aPpendiX" -> "Appendix A" and
				// "# Appendix Foo Bar" -> "Appendix A: Foo Bar".
				rest := strings.TrimSpace(headingText[len("appendix"):])
				if rest != "" {
					renderedHeadingText = fmt.Sprintf("Appendix %s: %s", headingNumberText, rest)
				} else {
					renderedHeadingText = fmt.Sprintf("Appendix %s", headingNumberText)
				}
				// The letter is part of the heading text, so leave the TOC number
				// empty and let the standardized text carry it.
				tocHeadingNumber = ""
				tocHeadingText = renderedHeadingText
			} else {
				renderedHeadingText = fmt.Sprintf(
					"%s&nbsp;&nbsp;&nbsp;&nbsp;%s",
					headingNumberText,
					headingText)
				tocHeadingNumber = headingNumberText
				tocHeadingText = headingText
			}

			idWithHeadingNumber := headingTextToId(renderedHeadingText)
			headingIdToHeadingIdWithLineNumber[idWithoutHeadingNumber] = idWithHeadingNumber

			line = fmt.Sprintf("%s %s", pieces[0], renderedHeadingText)
		}
		if pragmaIncludeInTocEnabled {
			r.TocItems = append(r.TocItems, tocItemT{
				HeadingLevel:  level,
				HeadingNumber: tocHeadingNumber,
				HeadingText:   tocHeadingText,
			})
		}

		// Register this heading as a wiki-link target. We key on the authored
		// heading text (case-insensitive, whitespace-trimmed). If two headings
		// share a name the first one wins, which keeps resolution deterministic.
		wikiKey := strings.ToLower(strings.TrimSpace(headingText))
		if _, exists := r.headingTargets[wikiKey]; !exists {
			r.headingTargets[wikiKey] = headingTargetT{
				displayText: strings.TrimSpace(headingText),
				id:          headingTextToId(renderedHeadingText),
			}
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
//
// Additionally, the following pragma directives can be used to control table cell
// background colors:
//
// @pragma{table_cell_bg_color_by_content:pass, #d0d0f0}
// @pragma{table_cell_bg_color_by_content:fail, #f0d0d0}
// @pragma{table_cell_bg_color_by_content:warning, #f0e0d0}
// @pragma{table_cell_bg_color_by_content:n/a, #808080}
// @pragma{table_cell_bg_color_by_content_partial:TBD, #f0f000}
// @pragma{table_cell_bg_color_clear_all}
//
// The parser will apply/remove the background color pragma directives and the @table directives
// in the order in which they appear in the Markdown source.
func (r *mdRendererT) MdParseCsvReferences(md string) string {
	qlog.Trace()
	reTable := regexp.MustCompile(`^\s*@table\{(attachments/[^}]+)\}\s*$`)
	reTableCellBgColor := regexp.MustCompile(`^\s*@pragma\{table_cell_bg_color_by_content:(.*?),\s*([^}]+)\}\s*$`)
	reTableCellBgColorPartial := regexp.MustCompile(`^\s*@pragma\{table_cell_bg_color_by_content_partial:(.*?),\s*([^}]+)\}\s*$`)
	reTableCellBgColorClearAll := regexp.MustCompile(`^\s*@pragma\{table_cell_bg_color_clear_all\}\s*$`)

	renderPendingCss := func(descriptors map[string]TableCellBgColorDescriptorT) string {
		type pendingDescriptorT struct {
			cellContent string
			descriptor  TableCellBgColorDescriptorT
		}
		pending := []pendingDescriptorT{}
		for cellContent, descriptor := range descriptors {
			if !descriptor.isCssRendered {
				pending = append(pending, pendingDescriptorT{
					cellContent: cellContent,
					descriptor:  descriptor,
				})
			}
		}
		if len(pending) == 0 {
			return ""
		}
		sort.Slice(pending, func(i, j int) bool {
			return pending[i].descriptor.id < pending[j].descriptor.id
		})
		var b strings.Builder
		b.WriteString("<style>\n")
		for _, item := range pending {
			fmt.Fprintf(
				&b,
				".content-page td:has(.td_bg_custom_%d){\n    background-color: %s;\n}\n",
				item.descriptor.id,
				item.descriptor.bgColor,
			)
			descriptor := descriptors[item.cellContent]
			descriptor.isCssRendered = true
			descriptors[item.cellContent] = descriptor
		}
		b.WriteString("</style>")
		return b.String()
	}

	renderTableMarkdown := func(csvRelativePath string, descriptors map[string]TableCellBgColorDescriptorT) (string, bool) {
		csvAbsolutePath := path.Join(r.PathMdSourceDir, csvRelativePath)
		data, err := os.ReadFile(csvAbsolutePath)
		if err != nil {
			qlog.Debugf("Could not read CSV table reference %q: %v", csvAbsolutePath, err)
			return "", false
		}

		reader := csv.NewReader(strings.NewReader(string(data)))
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		if err != nil {
			qlog.Debugf("Could not parse CSV table reference %q: %v", csvAbsolutePath, err)
			return "", false
		}
		if len(records) == 0 {
			qlog.Debugf("CSV table reference %q is empty", csvAbsolutePath)
			return "", false
		}

		columnCount := 0
		for _, record := range records {
			if len(record) > columnCount {
				columnCount = len(record)
			}
		}
		if columnCount == 0 {
			qlog.Debugf("CSV table reference %q has no columns", csvAbsolutePath)
			return "", false
		}

		padRecord := func(record []string) []string {
			if len(record) >= columnCount {
				return record
			}
			out := make([]string, columnCount)
			copy(out, record)
			return out
		}

		findMatchingDescriptor := func(cell string) (TableCellBgColorDescriptorT, bool) {
			cellLookup := strings.ToLower(strings.TrimSpace(cell))
			if descriptor, ok := descriptors[cellLookup]; ok && !descriptor.isPartialMatch {
				return descriptor, true
			}

			var bestMatch TableCellBgColorDescriptorT
			found := false
			for descriptorCellContent, descriptor := range descriptors {
				if !descriptor.isPartialMatch {
					continue
				}
				if strings.Contains(cellLookup, descriptorCellContent) {
					if !found || descriptor.id < bestMatch.id {
						bestMatch = descriptor
						found = true
					}
				}
			}
			return bestMatch, found
		}

		sanitizeCell := func(cell string) string {
			descriptor, ok := findMatchingDescriptor(cell)

			cell = strings.ReplaceAll(cell, "\r\n", "\n")
			cell = strings.ReplaceAll(cell, "\r", "\n")
			cell = strings.ReplaceAll(cell, "\n", "<br>")
			cell = strings.ReplaceAll(cell, "|", `\|`)
			if ok {
				cell += fmt.Sprintf(`<span class="td_bg_custom_%d"></span>`, descriptor.id)
			}
			return cell
		}

		header := padRecord(records[0])
		var b strings.Builder
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
		return b.String(), true
	}

	lines := strings.Split(md, "\n")
	linesOut := []string{}
	tableCellBgColorDescriptors := map[string]TableCellBgColorDescriptorT{}
	nextTableCellBgColorDescriptorID := 1

	for _, line := range lines {
		if matches := reTableCellBgColor.FindStringSubmatch(line); matches != nil {
			cellContent := strings.ToLower(strings.TrimSpace(matches[1]))
			bgColor := strings.TrimSpace(matches[2])
			tableCellBgColorDescriptors[cellContent] = TableCellBgColorDescriptorT{
				id:             nextTableCellBgColorDescriptorID,
				bgColor:        bgColor,
				isCssRendered:  false,
				isPartialMatch: false,
			}
			nextTableCellBgColorDescriptorID++
			continue
		}

		if matches := reTableCellBgColorPartial.FindStringSubmatch(line); matches != nil {
			cellContent := strings.ToLower(strings.TrimSpace(matches[1]))
			bgColor := strings.TrimSpace(matches[2])
			tableCellBgColorDescriptors[cellContent] = TableCellBgColorDescriptorT{
				id:             nextTableCellBgColorDescriptorID,
				bgColor:        bgColor,
				isCssRendered:  false,
				isPartialMatch: true,
			}
			nextTableCellBgColorDescriptorID++
			continue
		}

		if reTableCellBgColorClearAll.MatchString(line) {
			tableCellBgColorDescriptors = map[string]TableCellBgColorDescriptorT{}
			continue
		}

		if matches := reTable.FindStringSubmatch(line); matches != nil {
			csvRelativePath := matches[1]
			tableMarkdown, ok := renderTableMarkdown(csvRelativePath, tableCellBgColorDescriptors)
			if !ok {
				linesOut = append(linesOut, line)
				continue
			}

			cssBlock := renderPendingCss(tableCellBgColorDescriptors)
			if cssBlock != "" {
				linesOut = append(linesOut, cssBlock)
			}
			linesOut = append(linesOut, tableMarkdown)
			continue
		}

		linesOut = append(linesOut, line)
	}

	return strings.Join(linesOut, "\n")
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

// MdWrangleWikiLinks turns wiki-style links of the form "[[Heading Name]]" into
// markdown links to the referenced heading within this document.
//
// The lookup is case-insensitive and tolerates extra whitespace just inside the
// brackets (e.g. "[[  Heading Name  ]]"). The link text we emit is the actual
// heading text we found, so that a mistyped wiki link adopts the heading's
// correct capitalization in the output.
//
// If no matching heading is found we leave the original text untouched (brackets
// and all) so the author has a chance to notice the mistake in the output
// document.
func (r *mdRendererT) MdWrangleWikiLinks(md string) string {
	qlog.Trace()
	wikiLinkRe := regexp.MustCompile(`\[\[(.+?)\]\]`)
	return wikiLinkRe.ReplaceAllStringFunc(md, func(match string) string {
		inner := wikiLinkRe.FindStringSubmatch(match)[1]
		key := strings.ToLower(strings.TrimSpace(inner))
		target, ok := r.headingTargets[key]
		if !ok {
			qlog.Debugf("    🔴 Wiki link target not found: %q", match)
			return match
		}
		newLink := fmt.Sprintf("[%s](#%s)", target.displayText, target.id)
		qlog.Debugf("    REPLACEMENT: %q -> %q", match, newLink)
		return newLink
	})
}

// MdParseAsRunInline converts "as-run" delimiters into HTML wrappers.
//
// A line consisting solely of `@(` (ignoring surrounding whitespace) that is not
// inside a fenced code block is converted into an opening `<div class="asrun">`
// block; a line consisting solely of `)@` is converted into the matching closing
// `</div>`.
//
// When the delimiters appear inline (not alone on their line, and not in a fenced
// code block) they are replaced with placeholders that PostParseHtml later turns
// into <span> wrappers. A `@(` is only recognized when preceded by whitespace or
// the start of the line, and a `)@` only when followed by whitespace or the end
// of the line. This "break" requirement means a delimiter wrapped in backticks
// (e.g. `@(foo)@`) is left untouched and renders normally.
func (r *mdRendererT) MdParseAsRunInline(md string) string {
	openInlineRe := regexp.MustCompile(`(^|\s)@\(`)
	closeInlineRe := regexp.MustCompile(`\)@(\s|$)`)
	outLines := []string{}
	inFencedBlock := false
	for _, line := range strings.Split(md, "\n") {
		// Detect fenced code blocks
		if strings.HasPrefix(line, "```") {
			inFencedBlock = !inFencedBlock
			outLines = append(outLines, line)
			continue
		}
		if inFencedBlock {
			outLines = append(outLines, line)
			continue
		}
		// Handle delimiters that are "alone" on their line (ignoring whitespace)
		switch strings.TrimSpace(line) {
		case "@(":
			outLines = append(outLines, "\n<div class=\"asrun\">\n")
			continue
		case ")@":
			outLines = append(outLines, "\n</div>\n")
			continue
		}
		// Handle inline delimiters
		line = openInlineRe.ReplaceAllString(line, "${1}{{placeholder-open-asrun}}")
		line = closeInlineRe.ReplaceAllString(line, "{{placeholder-close-asrun}}${1}")
		outLines = append(outLines, line)
	}
	return strings.Join(outLines, "\n")
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
	// Replace as-run inline placeholders
	// -------------------------
	html = strings.ReplaceAll(html, "{{placeholder-open-asrun}}", "<span class=\"asrun\">")
	html = strings.ReplaceAll(html, "{{placeholder-close-asrun}}", "</span>")

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
