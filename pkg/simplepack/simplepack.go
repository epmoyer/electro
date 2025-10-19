package simplepack

import (
	"app/pkg/quicklog"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Version = "2.0.0"

const pathBuild = "project_build"
const pathEventData = pathBuild + "/app/js/event_data.js"
const pathTemp = pathBuild + "/temp"

type ReplacementDescriptorT struct {
	ContentType    string
	Regex          string
	ReplacementTag string
}

replacementDescriptors := []ReplacementDescriptorT{
	{
		ContentType:    "javascript",
		Regex:          `^\s*<script\s*type="text/javascript"\s*src="(.*)">\s*</script>\s$`,
		ReplacementTag: "script",
	},
	{
		ContentType:    "CSS",
		Regex:          `^\s*<link\s*rel="stylesheet"\s*type="text/css"\s*href="(.*)"\s*>\s*$`,
		ReplacementTag: "style",
	},
}

var qlog *quicklog.LoggerT = nil // Assigned at runtime

func init() {
	qlog = quicklog.GetLogger("default")
}

func SimplePack(pathFileIn string, pathFileOut string, enableMinify bool) error {

	qlog.Infof("SimplePack version: %q", Version)
	qlog.Infof("    Input file: %q", pathFileIn)
	qlog.Infof("    Output file: %q", pathFileOut)
	qlog.Infof("    Minify: %v", enableMinify)

	// Implementation goes here
	// read file
	data, err := os.ReadFile(pathFileIn)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	qlog.Debugf("Read %d lines", len(lines))

	newLines := []string{}
	for _, line := range lines {
		pathFileInParentDir := filepath.Dir(pathFileIn)
		expandedLines, err := expandLine(line, pathFileInParentDir, enableMinify)
		if err != nil {
			return fmt.Errorf("error expanding line: %w", err)
		}
		for _, expandedLine := range expandedLines {
			newLines = append(newLines, expandedLine)
		}
	}
	qlog.Infof("Writing: %q", pathFileOut)
	err = os.WriteFile(pathFileOut, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}
	return nil
}

func expandLine(line string, pathFileInParentDir string, enableMinify bool) ([]string, error) {

