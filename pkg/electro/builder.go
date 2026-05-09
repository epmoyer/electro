package electro

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	PathOutputDir                    string
	PathProjectDir                   string
	PathThemeDir                     string
	OutputFormat                     OuputFormatT
	Level1HeadingsAreDocumentTitles  bool
	MasterTitle                      string
	Watermark                        string
	ExcludeFromSearch                []string
	DoStripFrontmatter               bool
	NumberHeadings                   bool
	NumberHeadingsAtLevel            int
	SideMenuHeadingCaptureStartDepth int
	Footer                           string

	// Runtime
	MenuHtml             string
	OrderedDocumentnames []string
	SiteDocuments        map[string]siteDocumentT
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
	sideMenuHeadingCaptureStartDepth int,
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
		PathOutputDir:                    pathOutputDir,
		PathProjectDir:                   pathProjectDir,
		PathThemeDir:                     pathThemeDir,
		OutputFormat:                     outputFormat,
		Level1HeadingsAreDocumentTitles:  level1HeadingsAreDocumentTitles,
		MasterTitle:                      masterTitle,
		Watermark:                        watermark,
		ExcludeFromSearch:                excludeFromSearch,
		DoStripFrontmatter:               stripFrontmatter,
		NumberHeadings:                   numberHeadings,
		NumberHeadingsAtLevel:            numberHeadingsAtLevel,
		SideMenuHeadingCaptureStartDepth: sideMenuHeadingCaptureStartDepth,
		Footer:                           footer,
		SiteDocuments:                    make(map[string]siteDocumentT),
		MenuBuilder:                      &menuBuilderT{},
		SearchIndex:                      &searchIndex,
	}
}

func (b *builderT) AddNavigationDescriptor(nd navigationDescriptorT) error {
	qlog.Infof("Adding navigation section: %q", nd.Section)
	isDivider := len(nd.Documents.Keys()) == 0
	b.MenuBuilder.AddSection(nd.Section, isDivider)
	b.MenuHtml += "<ul class=\"menu-tree\">\n"
	menuNames := nd.Documents.Keys()
	fsysProject := os.DirFS(b.PathProjectDir)
	for _, menuName := range menuNames {
		mdDocumentName, _ := nd.Documents.Get(menuName)
		documentName := mdDocumentNameToDocumentName(mdDocumentName.(string))
		pathMarkdown := path.Join("docs", mdDocumentName.(string))
		err := b.BuildDocument(fsysProject, pathMarkdown, documentName)
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

	// Build list of two html heading tags to include in menu, of the form "h2", "h3", etc.
	includedLevels := []string{}
	for i := b.SideMenuHeadingCaptureStartDepth; i <= b.SideMenuHeadingCaptureStartDepth+1; i++ {
		includedLevels = append(includedLevels, fmt.Sprintf("h%d", i))
	}
	qlog.Infof("Building subheading menus for %s including levels: %+v", documentName, includedLevels)

	// Parse HTML to extract h2 and h3 headings
	headings := extractHeadings(documentHtml, includedLevels)
	qlog.Debugf("Extracted headings from %s: %+v", documentName, headings)

	for _, heading := range headings {

		// Determine the menuLevel: h2 = menuLevel 1, h3 = menuLevel 2
		// Determine menu level from position in includedLevels
		menuLevel := 1
		for i, level := range includedLevels {
			if heading.Tag == level {
				menuLevel = i + 1
				break
			}
		}

		// menuLevel := 1
		// if heading.Tag == "h3" {
		// 	menuLevel = 2
		// }

		// Generate heading ID from text
		headingId := headingTextToId(heading.Text)

		// Create link URL for the heading
		linkUrl := ""
		if b.OutputFormat == OutputFormatStaticSite {
			linkUrl = documentName + ".html#" + headingId
		}

		// Add the heading to the menu
		b.MenuBuilder.AddItem(menuLevel, heading.Text, headingId, linkUrl, documentName)
	}
}

func (b *builderT) BuildDocument(fsys fs.FS, pathMarkdown string, documentName string) error {
	if !pathIsFileFS(fsys, pathMarkdown) {
		return fmt.Errorf("markdown document does not exist: %q", pathMarkdown)
	}

	// -------------------------
	// Read markdown file
	// -------------------------
	mdData, err := fs.ReadFile(fsys, pathMarkdown)
	if err != nil {
		return fmt.Errorf("error reading markdown document %q: %w", pathMarkdown, err)
	}
	md := string(mdData)
	// Normalize line endings on load
	md = strings.ReplaceAll(md, "\r\n", "\n")

	// -------------------------
	// Render markdown to HTML
	// -------------------------
	filename := filepath.Base(pathMarkdown)
	renderer := NewMdRenderer(md, filename, b.PathOutputDir)
	renderer.DoStripFrontmatter = b.DoStripFrontmatter
	renderer.DoNumberHeadings = b.NumberHeadings
	renderer.NumberHeadingsAtLevel = b.NumberHeadingsAtLevel
	renderer.DoWrangleInterdocumentLinks = b.OutputFormat == OutputFormatSingleFile
	var html string
	html, err = renderer.Render()
	if err != nil {
		return fmt.Errorf("error rendering markdown document %q: %w", documentName, err)
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

	// Add document to search
	markdownFilename := filepath.Base(pathMarkdown)
	if !slices.Contains(b.ExcludeFromSearch, markdownFilename) {
		b.addDocumentToSearch(documentName, html)
	}

	// Update search

	b.OrderedDocumentnames = append(b.OrderedDocumentnames, documentName)
	b.SiteDocuments[documentName] = siteDocumentT{
		PathMarkdown: pathMarkdown,
		Html:         html,
	}

	return nil
}

func (b *builderT) addDocumentToSearch(documentName string, documentHtml string) error {
	qlog.Trace()

	// --------------------------------------------------------------------------------
	// This is SUBTLE and IMPORTANT:
	// Documents may contain special substitution placeholders of the form {{% placeholder %}}.
	// The use of such substitution placeholders can be OUTSIDE THE SCOPE of electro, so you
	// may find NO OTHER MENTION of them in the electro codebase.   On example is the Dynamo app
	// which uses a {{% doc_control_stamp_here %}} placeholder to mark the location where a
	// document control stamp gets inserted AFTER HTML generation.  It is VITALLY important that
	// we NOT include placeholders in the search index.  If we do, then the placeholder will
	// appear TWICE in the output HTML (once in the body, and once in the search data).
	// If the placeholder substitution ends up substituting the version in the search index,
	// then unpredictable things can happen (including breaking search because the index syntax
	// gets broken.  We have also seek knock-on effects such as breaking internal document
	// links.
	//
	// For that reason we MUST STRIP OUT ALL PLACEHOLDERS from the HTML before we add
	// it to the search index.
	// ---------------------------------------------------------------------------------

	// Strip out all placeholders of the form {{% placeholder %}}
	documentHtml = regexp.MustCompile(`\{\{%\s*.*?\s*%\}\}`).ReplaceAllString(documentHtml, "")

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(documentHtml))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find document title from first h1
	documentTitle := ""
	doc.Find("h1").First().Each(func(i int, s *goquery.Selection) {
		documentTitle = strings.TrimSpace(s.Text())
	})

	if documentTitle == "" {
		qlog.Infof("No h1 tag found in %s. Cannot extract document title for search.", documentName)
		documentTitle = "(Unknown)"
	}

	baseLocation := fmt.Sprintf("%s.html", documentName)
	currentLocation := baseLocation
	currentHeadingText := ""
	sectionText := ""

	// Iterate through h2, h3, p, li, th, td elements
	doc.Find("h2, h3, p, li, th, td").Each(func(i int, s *goquery.Selection) {
		tagName := goquery.NodeName(s)

		if tagName == "p" || tagName == "li" || tagName == "th" || tagName == "td" {
			// Paragraph, list item, or table cell
			text := strings.TrimSpace(s.Text())
			if sectionText != "" {
				sectionText += " "
			}
			sectionText += text
		} else {
			// Heading (h2 or h3)
			if sectionText != "" {
				b.addSearchItem(documentTitle, currentLocation, currentHeadingText, sectionText)
				sectionText = ""
			}

			currentHeadingText = strings.TrimSpace(s.Text())
			headingId := headingTextToId(currentHeadingText)
			qlog.Debugf("   %s : %s : %s", tagName, currentHeadingText, headingId)
			currentLocation = fmt.Sprintf("%s#%s", baseLocation, headingId)
		}
	})

	// Commit any remaining text
	if sectionText != "" {
		b.addSearchItem(documentTitle, currentLocation, currentHeadingText, sectionText)
	}

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
	// b.MenuBuilder.Dump(true)
	b.MenuHtml = b.MenuBuilder.RenderHtml()

	// -------------------
	// Copy CSS files
	// -------------------
	cssFiles := []string{
		"base_electro_doc.css",
		"base_electro_ui.css",
		"fonts.css",
		"fontawesome.css",
	}

	for _, filename := range cssFiles {
		fmt.Printf("Copying CSS file: %s\n", filename)
		srcPath := path.Join(b.PathThemeDir, filename)
		dstPath := path.Join(b.PathOutputDir, filename)
		err := copyFileFromFS(embeddedDataFS, srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("error copying CSS file %s: %w", filename, err)
		}
	}

	// -------------------
	// Copy CSS overlay
	// -------------------
	overlaySrcPath := path.Join(b.PathProjectDir, "docs", "overlay.css")
	overlayDstPath := path.Join(b.PathOutputDir, "overlay.css")
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
	imgSrcDir := path.Join(b.PathProjectDir, "docs", "img")
	imgDstDir := path.Join(b.PathOutputDir, "img")
	err := copyDirectoryContents(imgSrcDir, imgDstDir)
	if err != nil {
		qlog.Debugf("Note: Could not copy images directory: %v", err)
	}

	// -------------------
	// Copy Fonts
	// -------------------
	fontsSrcDir := path.Join(b.PathThemeDir, "fonts")
	fontsDstDir := path.Join(b.PathOutputDir, "fonts")
	err = copyDirectoryContentsFromFS(embeddedDataFS, fontsSrcDir, fontsDstDir)
	if err != nil {
		return fmt.Errorf("error copying fonts: %w", err)
	}

	// -------------------
	// Copy Attachments
	// -------------------
	attachSrcDir := path.Join(b.PathProjectDir, "docs", "attachments")
	attachDstDir := path.Join(b.PathOutputDir, "attachments")
	err = copyDirectoryContents(attachSrcDir, attachDstDir)
	if err != nil {
		qlog.Debugf("Note: Could not copy attachments directory: %v", err)
	}

	// -------------------
	// Copy Favicon
	// -------------------
	faviconSrcPath := path.Join(b.PathThemeDir, "favicon.ico")
	faviconDstPath := path.Join(b.PathOutputDir, "img", "favicon.ico")
	err = copyFileFromFS(embeddedDataFS, faviconSrcPath, faviconDstPath)
	if err != nil {
		qlog.Debugf("Note: Could not copy favicon: %v", err)
	}

	// -------------------
	// Copy JavaScript
	// -------------------

	// Copy site resources js
	jsSiteResourcesSrcDir := path.Join(pathDirSiteResourcesJs)
	jsSiteResourcesDstDir := path.Join(b.PathOutputDir, "js")
	err = copyDirectoryContentsFromFS(embeddedDataFS, jsSiteResourcesSrcDir, jsSiteResourcesDstDir)
	if err != nil {
		return fmt.Errorf("error copying core JavaScript files: %w", err)
	}

	// Copy theme js
	jsSrcDir := path.Join(b.PathThemeDir, "js")
	jsDstDir := path.Join(b.PathOutputDir, "js")
	err = copyDirectoryContentsFromFS(embeddedDataFS, jsSrcDir, jsDstDir)
	if err != nil {
		return fmt.Errorf("error copying JavaScript files: %w", err)
	}

	// -------------------
	// Build search results doc (empty placeholder for runtime search resutlts)
	// -------------------
	err = b.BuildDocument(dataFS, pathSearchResultsMd, "search")
	if err != nil {
		return err
	}

	// -------------------
	// Build site pages
	// -------------------
	templatePath := path.Join(b.PathThemeDir, "template.html")
	templateData, err := fs.ReadFile(embeddedDataFS, templatePath)
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
		pathSiteDocument := path.Join(b.PathOutputDir, "index.raw.html")
		err = b.renderDocument(templateHtml, pathSiteDocument, htmlPages, "Document")
		if err != nil {
			return fmt.Errorf("error rendering single file document: %w", err)
		}
	} else {
		for _, documentName := range b.OrderedDocumentnames {
			siteDocument := b.SiteDocuments[documentName]
			outputPath := path.Join(b.PathOutputDir, documentName+".html")
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
		indexPath := path.Join(b.PathOutputDir, "index.html")
		err = b.renderDocument(templateHtml, indexPath, firstDoc.Html, firstName)
		if err != nil {
			return fmt.Errorf("error rendering index.html: %w", err)
		}
	}

	// -------------------
	// Save search index
	// -------------------
	searchDir := path.Join(b.PathOutputDir, "search")
	err = os.MkdirAll(searchDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating search directory: %w", err)
	}

	searchIndexPath := path.Join(searchDir, "search_index.js")
	searchIndexJsonBytes, err := json.MarshalIndent(b.SearchIndex, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling search index to JSON: %w", err)
	}
	searchJs := fmt.Sprintf("App.searchData = %s;", string(searchIndexJsonBytes))

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

	// tocHtml := b.generateTocHtml(contentHtml)
	// documentHtml = strings.ReplaceAll(documentHtml, "{{% table_of_contents %}}", tocHtml)

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

// func generateTocHtml(contentHtml string) string {
// 	reHeading := regexp.MustCompile(`<h[\d].*?>.*?<\/h[\d]>`)
// 	// TODO: Implement TOC generation logic
// 	lines := strings.Split(contentHtml, "\n")
// 	for _, line := range lines {
// 		if reHeading.MatchString(line) {

// 	return ""
// }

func boolToJsTrueFalseText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func addIdTagsToHeadings(html string) string {
	// Match headings with optional attributes
	headingsRe := regexp.MustCompile(`<h([1-6])([^>]*)>(.*?)</h[1-6]>`)

	html = headingsRe.ReplaceAllStringFunc(html, func(heading string) string {
		matches := headingsRe.FindStringSubmatch(heading)
		if len(matches) != 4 {
			return heading
		}

		level := matches[1]         // "1", "2", etc.
		existingAttrs := matches[2] // existing attributes (may be empty or contain class, etc.)
		content := matches[3]       // heading text content

		// Verify closing tag matches opening tag level
		expectedClosing := fmt.Sprintf("</h%s>", level)
		if !strings.HasSuffix(heading, expectedClosing) {
			return heading // Mismatched tags, don't modify
		}

		// Check if id attribute already exists
		// if strings.Contains(existingAttrs, "id=") {
		// 	return heading
		// }

		id := headingTextToId(content)

		// Build new opening tag with id
		newOpenTag := fmt.Sprintf("<h%s id=\"%s\"%s>", level, id, existingAttrs)
		return fmt.Sprintf("%s%s</h%s>", newOpenTag, content, level)
	})

	return html
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
