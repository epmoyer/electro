package electro

type builderT struct {
	PathOutputDir  string
	PathProjectDir string
	PathThemeDir   string
	MenuHtml       string
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

func newBuilder(pathOutputDir string, pathProjectDir string, pathThemeDir string) *builderT {
	return &builderT{
		PathOutputDir:  pathOutputDir,
		PathProjectDir: pathProjectDir,
		PathThemeDir:   pathThemeDir,
		MenuBuilder:    &menuBuilderT{},
	}
}

func (b *builderT) AddNavigationDescriptor(nd navigationDescriptorT) error {
	isDivider := nd.Documents == nil || len(nd.Documents) == 0
	b.MenuBuilder.AddSection(nd.Section, isDivider)
	b.MenuHtml += "<ul class=\"menu-tree\">\n"
}

func (mb *menuBuilderT) AddSection(displayText string, isDivider bool) {
	section := menuSectionT{
		DisplayText: displayText,
		IsDivider:   isDivider,
	}
	mb.Sections = append(mb.Sections, section)
}
