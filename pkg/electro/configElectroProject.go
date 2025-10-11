package electro

import (
	"encoding/json"
	"os"
)

const projectFilename = "electro.json"

type configElectroProjectT struct {
	MasterTitle          string                  `json:"master_title"`
	Footer               string                  `json:"footer"`
	OutputDirectory      string                  `json:"output_directory"`
	Navigation           []navigationDescriptorT `json:"navigation"`
	Theme                string                  `json:"theme"`
	OutputFormat         string                  `json:"output_format"`
	Watermark            string                  `json:"watermark"`
	EnableNewlineToBreak bool                    `json:"enable_newline_to_break"`
}

type navigationDescriptorT struct {
	Section   string            `json:"section"`
	Documents map[string]string `json:"documents"`
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
