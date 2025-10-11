package electro

type builderT struct {
	PathOutputDir  string
	PathProjectDir string
	PathThemeDir   string
	MenuHtml       string
	SiteDocuments  map[string]string
	Substitutions  map[string]string
}

type menuBuilderT struct {
	Sections            []string
	CurrentDocumentName string
	IsFirstDivider      bool
}
