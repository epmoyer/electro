package electro

import (
	"path/filepath"
	"strings"
)

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
	LastChildAtLeven []menuItemT
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
		b.MenuBuilder.AddItem(0, menuName, linkUrl, documentName)
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

func mdDocumentNameToDocumentName(mdDocumentName string) string {
	return strings.TrimSuffix(filepath.Base(mdDocumentName), filepath.Ext(mdDocumentName))
}
