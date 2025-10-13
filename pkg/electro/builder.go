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
	Sections            []menuSectionT
	CurrentDocumentName string
	IsFirstDivider      bool
}

type menuItemT struct {
	DisplayText  string
	LinkUrl      string
	DocumentName string
	HeadingId    string
	Children     []menuItemT
}

type menuSectionT struct {
	DisplayText      string
	LastChildAtLevel []menuItemT
	Children         []menuItemT
	IsDivider        bool
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
	isDivider := nd.Documents == nil || len(nd.Documents) == 0
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
	b.MenuBuilder.Dump()
	return nil
}

func (mb *menuBuilderT) AddSection(displayText string, isDivider bool) {
	section := newMenuSection(displayText, isDivider)
	mb.Sections = append(mb.Sections, *section)
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
	section := &mb.Sections[len(mb.Sections)-1]
	section.Add(level, displayText, headingId, linkUrl, documentName)
}

func (mb *menuBuilderT) CullItemsAbove(level int) {
	// Remove all menu items ABOVE level (i.e. at an indent LESS than level).
	// NOTE: level is the "item" level, and does not include the section.
	qlog.Infof("CullItemsAbove(): %d", level)
	for i := range mb.Sections {
		// NOTE: level is the "item" level depth, and does not include the section, but we
		//       are recursing a tree where level 0 is the section node, so we add
		//       1 to the passed in level.
		mb.Sections[i].Children = mb.cullItemsAboveRecursive(
			level+1, 0, &menuItemT{Children: mb.Sections[i].Children})
	}
}

func (mb *menuBuilderT) cullItemsAboveRecursive(
	cullLevel,
	currentLevel int,
	node *menuItemT,
) []menuItemT {
	if cullLevel == currentLevel+1 {
		return node.Children
	}
	var retainedItems []menuItemT
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
	for i := range mb.Sections {
		// NOTE: level is the "item" level depth, and does not include the section, but we
		//       are recursing a tree where level 0 is the section node, so we add
		//       1 to the passed in level.
		tempNode := menuItemT{Children: mb.Sections[i].Children}
		mb.cullItemsBelowRecursive(level+1, 0, &tempNode)
		mb.Sections[i].Children = tempNode.Children
	}
}

func (mb *menuBuilderT) cullItemsBelowRecursive(cullLevel, currentLevel int, node *menuItemT) {
	if cullLevel == currentLevel {
		node.Children = []menuItemT{}
		return
	}

	for i := range node.Children {
		mb.cullItemsBelowRecursive(cullLevel, currentLevel+1, &node.Children[i])
	}
}

func newMenuSection(displayText string, isDivider bool) *menuSectionT {
	return &menuSectionT{
		DisplayText:      displayText,
		IsDivider:        isDivider,
		LastChildAtLevel: make([]menuItemT, maxMenuDepth),
	}
}

func (ms *menuSectionT) Add(
	level int,
	displayText string,
	headingId string,
	linkUrl string,
	documentName string,
) {
	newItem := menuItemT{
		DisplayText:  displayText,
		HeadingId:    headingId,
		LinkUrl:      linkUrl,
		DocumentName: documentName,
	}
	if level == 0 {
		ms.Children = append(ms.Children, newItem)
		ms.LastChildAtLevel[0] = newItem
	} else {
		if level < len(ms.LastChildAtLevel) {
			parent := &ms.LastChildAtLevel[level-1]
			parent.Children = append(parent.Children, newItem)
			// Clear "last child" of all levels deeper than this one.
			// NOTE: This is not strictly necessary, but it will
			//       defensively keep us from creating a weird tree if the
			//       input is badly formed.
			for i := level; i < len(ms.LastChildAtLevel); i++ {
				ms.LastChildAtLevel[i] = menuItemT{}
			}
		} else {
			// Invalid level, ignore
		}
	}
}

func mdDocumentNameToDocumentName(mdDocumentName string) string {
	return strings.TrimSuffix(filepath.Base(mdDocumentName), filepath.Ext(mdDocumentName))
}
