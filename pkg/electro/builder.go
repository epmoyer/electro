package electro

import (
	"path/filepath"
	"strings"
)

const maxMenuDepth = 6

type builderT struct {
	PathOutputDir  string
	PathProjectDir string
	PathThemeDir   string
	MenuHtml       string
	IsStaticSite   bool
	SiteDocuments  map[string]string
	Substitutions  map[string]string
	MenuBuilder    *menuBuilderT
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

func newBuilder(pathOutputDir string, pathProjectDir string, pathThemeDir string, isStaticSite bool) *builderT {
	return &builderT{
		PathOutputDir:  pathOutputDir,
		PathProjectDir: pathProjectDir,
		PathThemeDir:   pathThemeDir,
		MenuBuilder:    &menuBuilderT{},
		IsStaticSite:   isStaticSite,
	}
}

func (b *builderT) AddNavigationDescriptor(nd navigationDescriptorT) error {
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

func (mb *menuBuilderT) AddSection(displayText string, isDivider bool) {
	section := menuSectionT{
		DisplayText: displayText,
		IsDivider:   isDivider,
	}
	mb.Sections = append(mb.Sections, section)
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
			if level == len(ms.LastChildAtLevel)-1 {
				ms.LastChildAtLevel[level] = newItem
			} else {
				ms.LastChildAtLevel = ms.LastChildAtLevel[:level]
				ms.LastChildAtLevel = append(ms.LastChildAtLevel, newItem)
			}
		} else {
			// Invalid level, ignore
		}
	}
}

func mdDocumentNameToDocumentName(mdDocumentName string) string {
	return strings.TrimSuffix(filepath.Base(mdDocumentName), filepath.Ext(mdDocumentName))
}
