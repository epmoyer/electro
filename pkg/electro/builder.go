package electro

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/net/html"
)

const maxMenuDepth = 6

type MenuNodeTypeT int

const (
	NodeTypeMenuItem MenuNodeTypeT = iota
	NodeTypeMenuSection
)

type siteDocumentT struct {
	PathMarkdown string
	Html         string
}

type builderT struct {
	// Config
	PathOutputDir                   string
	PathProjectDir                  string
	PathThemeDir                    string
	OutputFormat                    OuputFormatT
	Level1HeadingsAreDocumentTitles bool
	MasterTitle                     string
	Watermark                       string
	ExcludeFromSearch               []string
	StripFrontmatter                bool
	NumberHeadings                  bool
	NumberHeadingsAtLevel           int
	Footer                          string

	// Runtime
	MenuHtml             string
	OrderedDocumentnames []string
	SiteDocuments        map[string]siteDocumentT
	Substitutions        map[string]string
	MenuBuilder          *menuBuilderT
	SearchIndex          *searchIndexT
}

type menuBuilderT struct {
	Nodes               []*menuNodeT
	CurrentDocumentName string
	IsFirstDivider      bool
}

type menuNodeT struct {
	// Common to all node types
	NodeType    MenuNodeTypeT
	DisplayText string
	Children    []*menuNodeT

	// NodeType:NodeMenuSection
	LastChildAtLevel []*menuNodeT
	IsDivider        bool

	// NodeType:NodeMenuItem
	LinkUrl      string
	DocumentName string
	HeadingId    string
}

type searchIndexT struct {
	Config searchIndexConfigT     `json:"config"`
	Docs   []searchIndexDocumentT `json:"docs"`
}

type searchIndexConfigT struct {
	Lang            []string `json:"lang"`
	MinSearchLength int      `json:"min_search_length"`
	PrebuildIndix   bool     `json:"prebuild_index"`
	Separator       string   `json:"separator"`
}

type searchIndexDocumentT struct {
	Title    string `json:"title"`
	Location string `json:"location"`
	Heading  string `json:"heading"`
	Text     string `json:"text"`
}

func newBuilder(pathOutputDir string,
	pathProjectDir string,
	pathThemeDir string,
	outputFormat OuputFormatT,
	level1HeadingsAreDocumentTitles bool,
	masterTitle string,
	watermark string,
	excludeFromSearch []string,
	stripFrontmatter bool,
	numberHeadings bool,
	numberHeadingsAtLevel int,
	footer string,
) *builderT {

	// Set defaults
	if numberHeadingsAtLevel == 0 {
		numberHeadingsAtLevel = 1
	}

	searchIndex := searchIndexT{
		Config: searchIndexConfigT{
			Lang:            []string{"en"},
			MinSearchLength: 3,
			PrebuildIndix:   false,
			Separator:       `[\s\-]+`,
		},
		Docs: []searchIndexDocumentT{},
	}

	return &builderT{
		// Config
		PathOutputDir:                   pathOutputDir,
		PathProjectDir:                  pathProjectDir,
		PathThemeDir:                    pathThemeDir,
		OutputFormat:                    outputFormat,
		Level1HeadingsAreDocumentTitles: level1HeadingsAreDocumentTitles,
		MasterTitle:                     masterTitle,
		Watermark:                       watermark,
		ExcludeFromSearch:               excludeFromSearch,
		StripFrontmatter:                stripFrontmatter,
		NumberHeadings:                  numberHeadings,
		NumberHeadingsAtLevel:           numberHeadingsAtLevel,
		Footer:                          footer,
		SiteDocuments:                   make(map[string]siteDocumentT),
		Substitutions:                   make(map[string]string),
		MenuBuilder:                     &menuBuilderT{},
		SearchIndex:                     &searchIndex,
	}
}

func (b *builderT) AddNavigationDescriptor(nd navigationDescriptorT) error {
	qlog.Infof("Adding navigation section: %q", nd.Section)
	isDivider := len(nd.Documents.Keys()) == 0
	b.MenuBuilder.AddSection(nd.Section, isDivider)
	b.MenuHtml += "<ul class=\"menu-tree\">\n"
	menuNames := nd.Documents.Keys()
	for _, menuName := range menuNames {
		mdDocumentName, _ := nd.Documents.Get(menuName)
		documentName := mdDocumentNameToDocumentName(mdDocumentName.(string))
		pathMarkdown := filepath.Join(b.PathProjectDir, "docs", mdDocumentName.(string))
		err := b.BuildDocument(pathMarkdown, documentName)
		if err != nil {
			return err
		}
		linkUrl := ""
		if b.OutputFormat == OutputFormatStaticSite {
			linkUrl = documentName + ".html"
		}
		b.MenuBuilder.AddItem(0, menuName, "", linkUrl, documentName)
		b.BuildSubheadingMenus(documentName)
	}
	b.MenuHtml += "</ul>\n"

	// For debugging:
	// qlog.Debugf("Menu structure after adding section %#v:", b.MenuBuilder.Nodes)
	// b.MenuBuilder.Dump(true)

	return nil
}

func (b *builderT) BuildSubheadingMenus(documentName string) {
	documentHtml := b.SiteDocuments[documentName].Html

	// Parse HTML to extract h2 and h3 headings
	headings := extractHeadings(documentHtml, []string{"h2", "h3"})
	qlog.Debugf("Extracted headings from %s: %+v", documentName, headings)

	for _, heading := range headings {
		// Determine the level: h2 = level 1, h3 = level 2
		level := 1
		if heading.Tag == "h3" {
			level = 2
		}

		// Generate heading ID from text
		headingId := headingTextToId(heading.Text)

		// Create link URL for the heading
		linkUrl := ""
		if b.OutputFormat == OutputFormatStaticSite {
			linkUrl = documentName + ".html#" + headingId
		}

		// Add the heading to the menu
		b.MenuBuilder.AddItem(level, heading.Text, headingId, linkUrl, documentName)
	}
}

func (b *builderT) BuildDocument(pathMarkdown string, documentName string) error {
	if !pathIsFile(pathMarkdown) {
		return fmt.Errorf("markdown document does not exist: %q", pathMarkdown)
	}

	// -------------------------
	// Read markdown file
	// -------------------------
	mdData, err := os.ReadFile(pathMarkdown)
	if err != nil {
		return fmt.Errorf("error reading markdown document %q: %w", pathMarkdown, err)
	}
	md := string(mdData)
	// Normalize line endings on load
	md = strings.ReplaceAll(md, "\r\n", "\n")

	// -------------------------
	// Pre-parser
	// -------------------------
	md, err = b.PreParseMarkdown(md)
	if err != nil {
		return fmt.Errorf("error pre-parsing markdown %q content: %w", pathMarkdown, err)
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
		return fmt.Errorf("error converting markdown to HTML for document %q: %w", documentName, err)
	}
	html := bufHtmlBytes.String()

	// -------------------------
	// Post-parser
	// -------------------------
	html, err = b.PostParseHtml(html)
	if err != nil {
		return fmt.Errorf("error post-processing HTML for document %q: %w", documentName, err)
	}

	// -------------------------
	// Modify HTML
	// -------------------------

	// Fix inter-document links
	linkRe := regexp.MustCompile(`<a href="\S*.md(?:\#\S*)?">`)
	links := linkRe.FindAllString(html, -1)
	for _, link := range links {
		html = strings.ReplaceAll(
			html, link, strings.ReplaceAll(link, ".md", ".html"))
	}

	// Wrap images
	imgTagRe := regexp.MustCompile(`<img .*?>`)
	imgTags := imgTagRe.FindAllString(html, -1)
	for _, imgTag := range imgTags {
		html = strings.ReplaceAll(
			html, imgTag, fmt.Sprintf("<div class=\"img-wrapper\">%s</div>", imgTag))
	}

	// Add id tags to headings
	html = addIdTagsToHeadings(html)

	// Add footer text
	html += ("<div class=\"no-indent\"><hr />\n" +
		"<div class=\"footer\">\n" +
		b.Footer +
		"</div>\n</div>\n")

	// FIXME:search: implement
	// Update search

	b.OrderedDocumentnames = append(b.OrderedDocumentnames, documentName)
	b.SiteDocuments[documentName] = siteDocumentT{
		PathMarkdown: pathMarkdown,
		Html:         html,
	}

	return nil
}

func (b *builderT) PreParseMarkdown(md string) (string, error) {
	var err error
	// -------------------------
	// Strip frontmatter
	// -------------------------
	if b.StripFrontmatter {
		md, err = b.stripFrontmatter(md)
		if err != nil {
			return "", fmt.Errorf("error stripping frontmatter: %w", err)
		}
	}

	// -------------------------
	// Tighten bullet lists
	// -------------------------
	md = b.MdTightenlBulletLists(md)

	// -------------------------
	// Number headings
	// -------------------------
	if b.NumberHeadings {
		md, err = b.MdAddHeadingNumbers(md)
		if err != nil {
			return "", fmt.Errorf("error numbering headings: %w", err)
		}
	}

	// -------------------------
	// Parse notices
	// -------------------------
	md, err = b.MdParseNotices(md)
	if err != nil {
		return "", fmt.Errorf("error parsing notices: %w", err)
	}

	// -------------------------
	// Parse checklists
	// -------------------------
	md = b.MdParseChecklists(md)

	// -------------------------
	// Parse interdocument links
	// -------------------------
	if b.OutputFormat == OutputFormatSingleFile {
		md = b.MdWrangleInterDocumentLinks(md)
	}
	return md, nil
}

func (b *builderT) stripFrontmatter(md string) (string, error) {
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

func (b *builderT) MdWrangleInterDocumentLinks(md string) string {
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

func (b *builderT) MdParseChecklists(md string) string {
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

func (b *builderT) MdTightenlBulletLists(md string) string {
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

func (b *builderT) MdAddHeadingNumbers(md string) (string, error) {
	headingManager := newHeadingManager(b.NumberHeadingsAtLevel)
	headingIdToHeadingIdWithLineNumber := make(map[string]string)
	renumberedLines := []string{}
	inFencedBlock := false
	lines := strings.Split(md, "\n")
	for _, line := range lines {
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

		if level < b.NumberHeadingsAtLevel {
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
	// Replace heading links to maktch re-numbered headings
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

func (b *builderT) MdParseNotices(md string) (string, error) {
	// Parse custom notice blockss
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
		sub := b.CreateSubstitution(htmlNoticeStart)
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
		sub := b.CreateSubstitution(snippetHtmlNoticeEnd)
		md = strings.ReplaceAll(md, noticeEndDirective, sub)
	}

	return md, nil
}

func (b *builderT) CreateSubstitution(final string) string {
	// Create a substitution entry and return the placeholder
	placeholder := fmt.Sprintf("<div class=\"PRE-PARSER-SUBSTITUTION-%d\"></div>", len(b.Substitutions)+1)
	b.Substitutions[placeholder] = final
	return placeholder
}

func (b *builderT) PostParseHtml(html string) (string, error) {
	for placeholder, final := range b.Substitutions {
		html = strings.ReplaceAll(html, placeholder, final)
	}
	return html, nil
}

func (b *builderT) addDocumentToSearch(documentName string, documentHtml string) error {
	qlog.Trace()
	b.addSearchItem("test", "test", "test", "test")
	return nil
}

func (b *builderT) addSearchItem(
	title string,
	location string,
	heading string,
	text string,
) {
	searchDoc := searchIndexDocumentT{
		Title:    title,
		Location: location,
		Heading:  heading,
		Text:     text,
	}
	b.SearchIndex.Docs = append(b.SearchIndex.Docs, searchDoc)
}

func (b *builderT) RenderSite() error {
	if b.Level1HeadingsAreDocumentTitles {
		b.MenuBuilder.CullItemsAbove(1)
	} else {
		b.MenuBuilder.CullItemsBelow(1)
	}
	// FIXME: Should we pass this flag as a command line arg, or show conditional on something else?
	b.MenuBuilder.Dump(true)
	b.MenuHtml = b.MenuBuilder.RenderHtml()

	// -------------------
	// Copy CSS files
	// -------------------
	cssFiles := []string{
		"base_electro_doc.css",
		"base_electro_ui.css",
		"base_pygments_monokai.css",
		"fonts.css",
		"fontawesome.css",
	}

	for _, filename := range cssFiles {
		srcPath := filepath.Join(b.PathThemeDir, filename)
		dstPath := filepath.Join(b.PathOutputDir, filename)
		err := copyFile(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("error copying CSS file %s: %w", filename, err)
		}
	}

	// -------------------
	// Copy CSS overlay
	// -------------------
	overlaySrcPath := filepath.Join(b.PathProjectDir, "docs", "overlay.css")
	overlayDstPath := filepath.Join(b.PathOutputDir, "overlay.css")
	if !pathIsFile(overlaySrcPath) {
		// Create empty overlay.css if it doesn't exist
		err := os.WriteFile(overlayDstPath, []byte(""), 0644)
		if err != nil {
			return fmt.Errorf("error creating empty overlay.css: %w", err)
		}
	} else {
		err := copyFile(overlaySrcPath, overlayDstPath)
		if err != nil {
			return fmt.Errorf("error copying overlay.css: %w", err)
		}
	}

	// FIXME: Append customizations and mixins to overlay.css

	// -------------------
	// Copy Images
	// -------------------
	imgSrcDir := filepath.Join(b.PathProjectDir, "docs", "img")
	imgDstDir := filepath.Join(b.PathOutputDir, "img")
	err := copyDirectoryContents(imgSrcDir, imgDstDir)
	if err != nil {
		qlog.Debugf("Note: Could not copy images directory: %v", err)
	}

	// -------------------
	// Copy Fonts
	// -------------------
	fontsSrcDir := filepath.Join(b.PathThemeDir, "fonts")
	fontsDstDir := filepath.Join(b.PathOutputDir, "fonts")
	err = copyDirectoryContents(fontsSrcDir, fontsDstDir)
	if err != nil {
		return fmt.Errorf("error copying fonts: %w", err)
	}

	// -------------------
	// Copy Attachments
	// -------------------
	attachSrcDir := filepath.Join(b.PathProjectDir, "docs", "attachments")
	attachDstDir := filepath.Join(b.PathOutputDir, "attachments")
	err = copyDirectoryContents(attachSrcDir, attachDstDir)
	if err != nil {
		qlog.Debugf("Note: Could not copy attachments directory: %v", err)
	}

	// -------------------
	// Copy Favicon
	// -------------------
	faviconSrcPath := filepath.Join(b.PathThemeDir, "favicon.ico")
	faviconDstPath := filepath.Join(b.PathOutputDir, "img", "favicon.ico")
	err = copyFile(faviconSrcPath, faviconDstPath)
	if err != nil {
		qlog.Debugf("Note: Could not copy favicon: %v", err)
	}

	// -------------------
	// Copy JavaScript
	// -------------------
	jsSrcDir := filepath.Join(b.PathThemeDir, "js")
	jsDstDir := filepath.Join(b.PathOutputDir, "js")
	err = copyDirectoryContents(jsSrcDir, jsDstDir)
	if err != nil {
		return fmt.Errorf("error copying JavaScript files: %w", err)
	}

	// -------------------
	// Build search results doc (placeholder)
	// -------------------
	// TODO: Implement search results document building
	err = b.BuildDocument(pathSearchResultsMd, "search")
	if err != nil {
		return err
	}

	// -------------------
	// Build site pages
	// -------------------
	templatePath := filepath.Join(b.PathThemeDir, "template.html")
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("error reading template file: %w", err)
	}
	templateHtml := string(templateData)

	if b.OutputFormat == OutputFormatSingleFile {
		htmlPages := ""
		htmlStyle := ""
		for _, documentName := range b.OrderedDocumentnames {
			siteDocument := b.SiteDocuments[documentName]
			htmlPages += "<div class=\"content-page\" id=\"" + documentName + "\"" + htmlStyle + ">\n"
			htmlPages += siteDocument.Html + "\n"
			htmlPages += "</div>\n"
			// Start all subsequent pages hidden
			htmlStyle = " style=\"display:none;\""
		}
		pathSiteDocument := filepath.Join(b.PathOutputDir, "index.raw.html")
		err = b.renderDocument(templateHtml, pathSiteDocument, htmlPages, "Document")
		if err != nil {
			return fmt.Errorf("error rendering single file document: %w", err)
		}
	} else {
		for _, documentName := range b.OrderedDocumentnames {
			siteDocument := b.SiteDocuments[documentName]
			outputPath := filepath.Join(b.PathOutputDir, documentName+".html")
			documentHtml := "<div class=\"content-page\">" + siteDocument.Html + "</div>"

			err = b.renderDocument(templateHtml, outputPath, documentHtml, documentName)
			if err != nil {
				return fmt.Errorf("error rendering document %s: %w", documentName, err)
			}
		}
	}

	// Build index.html (first document or main page)
	if len(b.SiteDocuments) > 0 {
		// Use first document as index
		var firstDoc siteDocumentT
		var firstName string
		for name, doc := range b.SiteDocuments {
			firstDoc = doc
			firstName = name
			break
		}
		indexPath := filepath.Join(b.PathOutputDir, "index.html")
		err = b.renderDocument(templateHtml, indexPath, firstDoc.Html, firstName)
		if err != nil {
			return fmt.Errorf("error rendering index.html: %w", err)
		}
	}

	// -------------------
	// Save search index
	// -------------------
	searchDir := filepath.Join(b.PathOutputDir, "search")
	err = os.MkdirAll(searchDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating search directory: %w", err)
	}

	searchIndexPath := filepath.Join(searchDir, "search_index.js")
	// TODO: Implement proper search index building
	searchJs := `App.searchData = {
    "config": {
        "lang": ["en"],
        "min_search_length": 3,
        "prebuild_index": false,
        "separator": "[\\s\\-]+"
    },
    "docs": []
};`

	err = os.WriteFile(searchIndexPath, []byte(searchJs), 0644)
	if err != nil {
		return fmt.Errorf("error writing search index: %w", err)
	}

	return nil
}

func (mb *menuBuilderT) AddSection(displayText string, isDivider bool) {
	section := newMenuSection(displayText, isDivider)
	mb.Nodes = append(mb.Nodes, section)
}

func (mb *menuBuilderT) AddItem(
	level int,
	displayText string,
	headingId string,
	linkUrl string,
	documentName string,
) {
	qlog.Debugf("AddItem(): level=%d displayText=%q headingId=%q linkUrl=%q documentName=%q", level, displayText, headingId, linkUrl, documentName)
	if documentName != "" {
		mb.CurrentDocumentName = documentName
	}
	currentNode := mb.Nodes[len(mb.Nodes)-1]
	qlog.Debugf("Current node before Add: %+v", currentNode)
	currentNode.Add(level, displayText, headingId, linkUrl, documentName)
}

func (mb *menuBuilderT) CullItemsAbove(level int) {
	// Remove all menu items ABOVE level (i.e. at an indent LESS than level).
	// NOTE: level is the "item" level, and does not include the section.
	qlog.Infof("CullItemsAbove(): %d", level)
	for i := range mb.Nodes {
		// NOTE: level is the "item" level depth, and does not include the section, but we
		//       are recursing a tree where level 0 is the section node, so we add
		//       1 to the passed in level.
		mb.Nodes[i].Children = mb.cullItemsAboveRecursive(
			level+1,
			0,
			newMenuItem("", mb.Nodes[i].Children, "", "", ""),
		)
	}
}

func (mb *menuBuilderT) cullItemsAboveRecursive(
	cullLevel,
	currentLevel int,
	node *menuNodeT,
) []*menuNodeT {
	if cullLevel == currentLevel+1 {
		return node.Children
	}
	var retainedItems []*menuNodeT
	for i := range node.Children {
		retainedItems = append(
			retainedItems,
			mb.cullItemsAboveRecursive(cullLevel, currentLevel+1, node.Children[i])...)
	}
	return retainedItems
}

func (mb *menuBuilderT) CullItemsBelow(level int) {
	// Remove all menu items BELOW level (i.e. at an indent GREATER than level).
	// NOTE: level is the "item" level, and does not include the section.
	qlog.Infof("CullItemsBelow(): %d", level)
	for i := range mb.Nodes {
		// NOTE: level is the "item" level depth, and does not include the section, but we
		//       are recursing a tree where level 0 is the section node, so we add
		//       1 to the passed in level.
		tempNode := newMenuItem("", mb.Nodes[i].Children, "", "", "")
		mb.cullItemsBelowRecursive(level+1, 0, tempNode)
		mb.Nodes[i].Children = tempNode.Children
	}
}

func (mb *menuBuilderT) cullItemsBelowRecursive(cullLevel, currentLevel int, node *menuNodeT) {
	if cullLevel == currentLevel {
		node.Children = []*menuNodeT{}
		return
	}

	for i := range node.Children {
		mb.cullItemsBelowRecursive(cullLevel, currentLevel+1, node.Children[i])
	}
}

func (mb *menuBuilderT) Dump(display bool) {
	for _, section := range mb.Nodes {
		mb.DumpRecursive(section, display, 0)
	}
}

func (mb *menuBuilderT) DumpRecursive(node *menuNodeT, display bool, level int) {
	indent := strings.Repeat("    ", level)
	var prefix string
	if node.NodeType == NodeTypeMenuSection {
		prefix = "@ "
	} else {
		prefix = "- "
	}
	text := indent + prefix + " " + node.DisplayText
	if node.NodeType == NodeTypeMenuItem {
		text += fmt.Sprintf(":: %s, %s, %s", node.DocumentName, node.LinkUrl, node.HeadingId)
	}
	qlog.Debug(text)
	if display {
		fmt.Println(text)
	}
	for i := range node.Children {
		mb.DumpRecursive(node.Children[i], display, level+1)
	}
}

func (mb *menuBuilderT) RenderHtml() string {
	var html string
	for _, node := range mb.Nodes {
		html += mb.RenderHtmlNode(node)
	}
	return html
}

func (mb *menuBuilderT) RenderHtmlNode(node *menuNodeT) string {
	var html string
	if node.IsDivider {
		if mb.IsFirstDivider {
			mb.IsFirstDivider = false
		} else {
			html += "<hr>\n"
		}
	}
	var headingClass string
	if node.IsDivider {
		headingClass = "section-heading-divider"
	} else {
		headingClass = "section-heading"
	}
	if node.DisplayText != "" {
		html += fmt.Sprintf("<div class=\"%s\">%s</div>\n", headingClass, node.DisplayText)
	}
	html += "<ul class=\"menu-tree\">\n"
	html += mb.renderHtmlForNodeChildren(0, node.Children)
	html += "</ul>\n"
	return html
}

func (mb *menuBuilderT) renderHtmlForNodeChildren(level int, children []*menuNodeT) string {
	var html string

	for _, child := range children {
		// Submenu items
		var submenuHtml string
		var caretVisible bool
		if level == 0 && len(child.Children) > 0 {
			submenuHtml = "<ul class=\"nested\">\n" +
				mb.renderHtmlForNodeChildren(level+1, child.Children) +
				"</ul>\n"
			caretVisible = true
		} else {
			submenuHtml = ""
			caretVisible = false
		}

		// Build classes
		var classes []string
		if level == 0 {
			classes = []string{"level-0"}
		} else {
			classes = []string{"no-child"}
		}
		classList := strings.Join(classes, " ")
		classStatement := fmt.Sprintf("class=\"%s\"", classList)

		// Build heading ID statement
		headingIdStatement := ""
		if child.HeadingId != "" {
			headingIdStatement = fmt.Sprintf("data-target-heading-id=%s", child.HeadingId)
		}

		// Build the menu item HTML
		html += "<li>\n"
		html += fmt.Sprintf("<span %s id=\"menuitem_doc_%s\" data-document-name=\"%s\" %s>\n",
			classStatement,
			child.DocumentName,
			child.DocumentName,
			headingIdStatement)

		html += mb.formatMenuHeading(
			child.DisplayText,
			level == 0,    // includeCaretSpace
			caretVisible,  // caretVisible
			child.LinkUrl, // linkUrl
			level > 0,     // isLevelTwo
		)

		html += "</span>\n"
		html += submenuHtml
		html += "</li>\n"
	}

	return html
}

func (mb *menuBuilderT) formatMenuHeading(
	text string,
	includeCaretSpace bool,
	caretVisible bool,
	linkUrl string,
	isLevelTwo bool,
) string {
	// Given a menu heading, split it into two divs if it has a numeric prefix.
	// For headings that start with a section number (e.g. "1.5 Study Results") we
	// will split the heading into two pieces and wrap them each in a div.

	var caretItemContent string
	var numberItemContent string
	textItemContent := text
	var coreContent string

	if includeCaretSpace {
		if caretVisible {
			caretItemContent = "<i class=\"caret fa fa-angle-right\"></i>"
		} else {
			caretItemContent = "<i class=\"caret-placeholder\"></i>"
		}
	}

	// Replace non-breaking space with regular space
	text = strings.ReplaceAll(text, "\u00a0", " ")

	if strings.Contains(text, " ") {
		numberedParts := mb.splitIfNumbered(text)
		if numberedParts != nil {
			numberItemContent = fmt.Sprintf("<div class=\"number-item\">%s</div>", numberedParts[0])
			textItemContent = numberedParts[1]
		}
	}

	if numberItemContent != "" {
		// We have a numbered heading
		classes := "number-item"
		if isLevelTwo {
			numberItemContent = fmt.Sprintf(
				"<div class=\"%s\">%s</div>", classes, numberItemContent)
		}
	}
	textItemContent = fmt.Sprintf("<div class=\"text-item\">%s</div>", textItemContent)
	coreContent = fmt.Sprintf("<div class=\"core\">%s%s</div>", numberItemContent, textItemContent)

	if linkUrl != "" {
		coreContent = fmt.Sprintf("<a href=\"%s\">%s</a>", linkUrl, coreContent)
	}

	html := fmt.Sprintf("<div class=\"menu-item-container\">%s%s</div>", caretItemContent, coreContent)

	qlog.Debugf("formatMenuHeading: result: %q", html)
	return html
}

func (mb *menuBuilderT) splitIfNumbered(text string) []string {
	// Given a menu heading, split it into two strings if it has a numeric prefix.
	// For headings that start with a section number (e.g. "1.5 Study Results") we
	// will split the heading into two pieces and return a slice.
	// Otherwise return nil

	pieces := strings.Fields(text)
	if len(pieces) == 0 {
		return nil
	}

	headingNumber := pieces[0]

	// Check if it matches a numeric pattern like "1.2.3"
	isNumbered := true
	for _, char := range headingNumber {
		if char != '.' && (char < '0' || char > '9') {
			isNumbered = false
			break
		}
	}

	var result []string
	if isNumbered && len(pieces) > 1 {
		result = []string{headingNumber, strings.Join(pieces[1:], " ")}
	}

	qlog.Debugf("splitIfNumbered(): headingNumber=%s text=%s => %v", headingNumber, text, result)
	return result
}

func newMenuSection(displayText string, isDivider bool) *menuNodeT {
	return &menuNodeT{
		NodeType:         NodeTypeMenuSection,
		DisplayText:      displayText,
		IsDivider:        isDivider,
		LastChildAtLevel: make([]*menuNodeT, maxMenuDepth),
	}
}

func newMenuItem(displayText string, children []*menuNodeT, headingId string, linkUrl string, documentName string) *menuNodeT {
	return &menuNodeT{
		NodeType:     NodeTypeMenuItem,
		DisplayText:  displayText,
		Children:     children,
		HeadingId:    headingId,
		LinkUrl:      linkUrl,
		DocumentName: documentName,
	}
}

func (mn *menuNodeT) Add(
	level int,
	displayText string,
	headingId string,
	linkUrl string,
	documentName string,
) {
	newItem := newMenuItem(displayText, []*menuNodeT{}, headingId, linkUrl, documentName)
	if level == 0 {
		mn.Children = append(mn.Children, newItem)
		mn.LastChildAtLevel[0] = newItem
	} else {
		if level < len(mn.LastChildAtLevel) {
			parent := mn.LastChildAtLevel[level-1]
			parent.Children = append(parent.Children, newItem)
			// Clear "last child" of all levels deeper than this one.
			// NOTE: This is not strictly necessary, but it will
			//       defensively keep us from creating a weird tree if the
			//       input is badly formed.
			// for i := level; i < len(mn.LastChildAtLevel); i++ {
			// 	mn.LastChildAtLevel[i] = &menuNodeT{}
			// }
			mn.LastChildAtLevel[level] = newItem
		} else {
			// Invalid level, ignore
		}
	}
}

func mdDocumentNameToDocumentName(mdDocumentName string) string {
	return strings.TrimSuffix(filepath.Base(mdDocumentName), filepath.Ext(mdDocumentName))
}

func (b *builderT) renderDocument(templateHtml, outputPath, contentHtml, documentName string) error {
	qlog.Infof("Building %s...", outputPath)

	masterTitle := b.MasterTitle
	watermark := b.Watermark

	documentHtml := strings.ReplaceAll(templateHtml, "{{% content %}}", contentHtml)
	documentHtml = strings.ReplaceAll(documentHtml, "{{% master_title %}}", masterTitle)
	documentHtml = strings.ReplaceAll(documentHtml, "{{% master_title_nav %}}", strings.ReplaceAll(masterTitle, "<br>", " "))
	documentHtml = strings.ReplaceAll(documentHtml, "{{% sidebar_menu %}}", b.MenuHtml)
	documentHtml = strings.ReplaceAll(documentHtml, "{{% current_document_name %}}", documentName)
	// NOTE: These single quotes are intentional.  We are replacing a single quoted string in the template with a JS true or false.
	documentHtml = strings.ReplaceAll(documentHtml, "'{{% single_file %}}'", boolToJsTrueFalseText(b.OutputFormat == OutputFormatSingleFile))
	documentHtml = strings.ReplaceAll(documentHtml, "{{% watermark %}}", watermark)
	documentHtml = strings.ReplaceAll(documentHtml, "{{% electro_version %}}", config.Version)
	documentHtml = strings.ReplaceAll(documentHtml, "{{% year %}}", time.Now().Format("2006"))
	documentHtml = strings.ReplaceAll(documentHtml, "{{% timestamp %}}", time.Now().Format(time.RFC3339))

	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	err = os.WriteFile(outputPath, []byte(documentHtml), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	return nil
}

func boolToJsTrueFalseText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func addIdTagsToHeadings(html string) string {
	headingsRe := regexp.MustCompile(`<h\d>.*<\/h\d>`)
	headings := headingsRe.FindAllString(html, -1)
	for _, heading := range headings {
		core := heading[4 : len(heading)-5]
		tagStart := heading[:3]
		id := headingTextToId(core)
		replacement := strings.Replace(
			heading,
			tagStart,
			fmt.Sprintf("%s id=\"%s\"", tagStart, id),
			1)
		// fmt.Printf("🟣  heading: %q, core: %q, id: %q, replacement: %q\n",
		// 	heading, core, id, replacement)
		html = strings.Replace(html, heading, replacement, -1)
	}
	return html
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func copyDirectoryContents(srcDir, dstDir string) error {
	// Check if source directory exists
	if !pathIsDir(srcDir) {
		return fmt.Errorf("source directory does not exist: %s", srcDir)
	}

	// Create destination directory
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			err = copyDirectoryContents(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// headingT represents an HTML heading element
type headingT struct {
	Tag  string
	Text string
}

// extractHeadings parses HTML and extracts headings of specified tags
func extractHeadings(htmlContent string, tags []string) []headingT {
	var headings []headingT

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		qlog.Errorf("Error parsing HTML: %v", err)
		return headings
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, tag := range tags {
				if n.Data == tag {
					text := extractTextContent(n)
					if text != "" {
						headings = append(headings, headingT{
							Tag:  tag,
							Text: text,
						})
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return headings
}

// extractTextContent extracts text content from an HTML node
func extractTextContent(n *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(text.String())
}
