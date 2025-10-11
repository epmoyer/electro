package main

import (
	"encoding/json"
	"os"
)

type configElectroProjectT struct {
	MasterTitle     string                            `json:"master_title"`
	Footer          string                            `json:"footer"`
	OutputDirectory string                            `json:"output_directory"`
	Navigation      []configElectroProjectNavigationT `json:"navigation"`
	Theme           string                            `json:"theme"`
	OutputFormat    string                            `json:"output_format"`
	Watermark       string                            `json:"watermark"`
}

type configElectroProjectNavigationT struct {
	Section   string            `json:"section"`
	Documents map[string]string `json:"documents"`
}

func loadConfigElectroProject(pathProjectFile string) (*configElectroProjectT, error) {

	var project configElectroProjectT
	err := loadJsonFile(pathProjectFile, &project)
	if err != nil {
		return nil, err
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
