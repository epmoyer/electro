package electro

import (
	"encoding/json"
	"os"

	"github.com/epmoyer/electro/v3/pkg/orderedmap"
)

const projectFilename = "electro.json"

type OuputFormatT int

const (
	OutputFormatStaticSite OuputFormatT = iota
	OutputFormatSingleFile
)

type configElectroProjectT struct {
	MasterTitle                        string                  `json:"master_title"`
	Footer                             string                  `json:"footer"`
	OutputDirectory                    string                  `json:"output_directory"`
	Navigation                         []navigationDescriptorT `json:"navigation"`
	Theme                              string                  `json:"theme"`
	OutputFormat                       string                  `json:"output_format"`
	PathOutputSingleFileTargetRelative string                  `json:"output_single_file"`
	Watermark                          string                  `json:"watermark"`
	ExcludeFromSearch                  []string                `json:"exclude_from_search"`
	EnableNewlineToBreak               bool                    `json:"enable_newline_to_break"`
	Level1HeadingsAreDocumentTitles    bool                    `json:"level_1_headings_are_document_titles"`
	SideMenuHeadingCaptureStartDepth   int                     `json:"side_menu_heading_capture_start_depth"`
	StripFrontmatter                   bool                    `json:"strip_frontmatter"`
	NumberHeadings                     bool                    `json:"number_headings"`
	NumberHeadingsAtLevel              int                     `json:"number_headings_at_level"`
}

type navigationDescriptorT struct {
	Section   string                 `json:"section"`
	Documents *orderedmap.OrderedMap `json:"documents"`
}

func loadConfigElectroProject(pathProjectFile string) (*configElectroProjectT, error) {

	var project configElectroProjectT
	err := loadJsonFile(pathProjectFile, &project)
	if err != nil {
		return nil, err
	}
	if project.OutputFormat == "" {
		project.OutputFormat = "static_site"
	}
	if project.SideMenuHeadingCaptureStartDepth == 0 {
		// Default to start capturing from l2 (##) headings
		project.SideMenuHeadingCaptureStartDepth = 2
	}

	return &project, nil
}

func loadJsonFile(pathFile string, v interface{}) error {
	data, err := os.ReadFile(pathFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
