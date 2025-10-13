package electro

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
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
	IsStaticSite                    bool
	Level1HeadingsAreDocumentTitles bool

	// Runtime
	MenuHtml      string
	SiteDocuments map[string]siteDocumentT
	Substitutions map[string]string
	MenuBuilder   *menuBuilderT
	// FIXME: add SearchIndex
}

type menuBuilderT struct {
	Nodes               []menuNodeT
	CurrentDocumentName string
	IsFirstDivider      bool
}

// type menuItemT struct {
// 	DisplayText  string
// 	LinkUrl      string
// 	DocumentName string
// 	HeadingId    string
// 	Children     []menuItemT
// }

// type menuSectionT struct {
// 	DisplayText      string
// 	LastChildAtLevel []menuItemT
// 	Children         []menuItemT
// 	IsDivider        bool
// }

type menuNodeT struct {
	// Common to all node types
	NodeType    MenuNodeTypeT
	DisplayText string
	Children    []menuNodeT

	// NodeType:NodeMenuSection
	LastChildAtLevel []menuNodeT
	IsDivider        bool

	// NodeType:NodeMenuItem
	LinkUrl      string
	DocumentName string
	HeadingId    string
}

func newBuilder(pathOutputDir string,
	pathProjectDir string,
	pathThemeDir string,
	isStaticSite bool,
	level1HeadingsAreDocumentTitles bool,
) *builderT {
	return &builderT{
		// Config
		PathOutputDir:                   pathOutputDir,
		PathProjectDir:                  pathProjectDir,
		PathThemeDir:                    pathThemeDir,
		IsStaticSite:                    isStaticSite,
		Level1HeadingsAreDocumentTitles: level1HeadingsAreDocumentTitles,
		// Runtime
		SiteDocuments: make(map[string]siteDocumentT),
		Substitutions: make(map[string]string),
		MenuBuilder:   &menuBuilderT{},
	}
}

func (b *builderT) AddNavigationDescriptor(nd navigationDescriptorT) error {
	qlog.Infof("Adding navigation section: %q", nd.Section)
	isDivider := len(nd.Documents) == 0
	b.MenuBuilder.AddSection(nd.Section, isDivider)
	b.MenuHtml += "<ul class=\"menu-tree\">\n"
	for menuName, mdDocumentName := range nd.Documents {
		documentName := mdDocumentNameToDocumentName(mdDocumentName)
		pathMarkdown := filepath.Join(b.PathProjectDir, "docs", mdDocumentName)
		err := b.BuildDocument(pathMarkdown, documentName)
		if err != nil {
			return err
		}
		linkUrl := ""
		if b.IsStaticSite {
			linkUrl = documentName + ".html"
		}
		b.MenuBuilder.AddItem(0, menuName, "", linkUrl, documentName)
	}
	b.MenuHtml += "</ul>\n"
	return nil
}

func (b *builderT) BuildDocument(pathMarkdown string, documentName string) error {
	if !pathIsFile(pathMarkdown) {
		return fmt.Errorf("markdown document does not exist: %q", pathMarkdown)
	}
	// Read markdown file
	mdData, err := os.ReadFile(pathMarkdown)
	if err != nil {
		return fmt.Errorf("error reading markdown document %q: %w", pathMarkdown, err)
	}
	var bufHtmlBytes bytes.Buffer
	err = goldmark.Convert(mdData, &bufHtmlBytes)
	if err != nil {
		return fmt.Errorf("error converting markdown to HTML for document %q: %w", documentName, err)
	}

	// FIXME: implement
	// Fix inter-document links
	// Wrap images
	// Add id tags to headings
	// Add footer text
	// Update search

	b.SiteDocuments[documentName] = siteDocumentT{
		PathMarkdown: pathMarkdown,
		Html:         bufHtmlBytes.String(),
	}

	return nil
}

func (b *builderT) RenderSite() error {
	// FIXME: implement
	if b.Level1HeadingsAreDocumentTitles {
		b.MenuBuilder.CullItemsAbove(1)
	} else {
		b.MenuBuilder.CullItemsBelow(1)
	}
	// FIXME: Should we pass this flag as a command line arg, or show conditional on something else?
	b.MenuBuilder.Dump(true)
	b.MenuHtml = b.MenuBuilder.RenderHtml()

	// Finish implementation from this point forward.
	return nil
}

func (mb *menuBuilderT) AddSection(displayText string, isDivider bool) {
	section := newMenuSection(displayText, isDivider)
	mb.Nodes = append(mb.Nodes, *section)
}

func (mb *menuBuilderT) AddItem(
	level int,
	displayText string,
	headingId string,
	linkUrl string,
	documentName string,
) {
	if documentName != "" {
		mb.CurrentDocumentName = documentName
	}
	section := &mb.Nodes[len(mb.Nodes)-1]
	section.Add(level, displayText, headingId, linkUrl, documentName)
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
) []menuNodeT {
	if cullLevel == currentLevel+1 {
		return node.Children
	}
	var retainedItems []menuNodeT
	for i := range node.Children {
		retainedItems = append(
			retainedItems,
			mb.cullItemsAboveRecursive(cullLevel, currentLevel+1, &node.Children[i])...)
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
		node.Children = []menuNodeT{}
		return
	}

	for i := range node.Children {
		mb.cullItemsBelowRecursive(cullLevel, currentLevel+1, &node.Children[i])
	}
}

func (mb *menuBuilderT) Dump(display bool) {
	for _, section := range mb.Nodes {
		mb.DumpRecursive(section, display, 0)
	}
}

func (mb *menuBuilderT) DumpRecursive(node menuNodeT, display bool, level int) {
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

func (mb *menuBuilderT) RenderHtmlNode(node menuNodeT) string {
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

func (mb *menuBuilderT) renderHtmlForNodeChildren(level int, children []menuNodeT) string {
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
			caretItemContent = "<div class=\"caret-item caret-down\"></div>"
		} else {
			caretItemContent = "<div class=\"caret-item\"></div>"
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
		LastChildAtLevel: make([]menuNodeT, maxMenuDepth),
	}
}

func newMenuItem(displayText string, children []menuNodeT, headingId string, linkUrl string, documentName string) *menuNodeT {
	return &menuNodeT{
		NodeType:     NodeTypeMenuItem,
		DisplayText:  displayText,
		Children:     children,
		HeadingId:    headingId,
		LinkUrl:      linkUrl,
		DocumentName: documentName,
	}
}

func (ms *menuNodeT) Add(
	level int,
	displayText string,
	headingId string,
	linkUrl string,
	documentName string,
) {
	newItem := newMenuItem(displayText, []menuNodeT{}, headingId, linkUrl, documentName)
	if level == 0 {
		ms.Children = append(ms.Children, *newItem)
		ms.LastChildAtLevel[0] = *newItem
	} else {
		if level < len(ms.LastChildAtLevel) {
			parent := &ms.LastChildAtLevel[level-1]
			parent.Children = append(parent.Children, *newItem)
			// Clear "last child" of all levels deeper than this one.
			// NOTE: This is not strictly necessary, but it will
			//       defensively keep us from creating a weird tree if the
			//       input is badly formed.
			for i := level; i < len(ms.LastChildAtLevel); i++ {
				ms.LastChildAtLevel[i] = menuNodeT{}
			}
		} else {
			// Invalid level, ignore
		}
	}
}

func mdDocumentNameToDocumentName(mdDocumentName string) string {
	return strings.TrimSuffix(filepath.Base(mdDocumentName), filepath.Ext(mdDocumentName))
}
