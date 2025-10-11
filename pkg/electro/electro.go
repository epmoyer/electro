package electro

import (
	"app/pkg/quicklog"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var qlog *quicklog.LoggerT = nil // Assigned at runtime

const pathDirThemes = "../pkg/electro/data/themes"

func init() {
	qlog = quicklog.GetLogger("default")
}

func BuildProject(pathCommandLineArg string) error {
	var pathProjectDir string
	var pathProjectFile string

	// Fail if path does not exist
	if !pathExists(pathCommandLineArg) {
		return fmt.Errorf("path does not exist: %q", pathCommandLineArg)
	}
	if pathIsDir(pathCommandLineArg) {
		// ------------------------
		// Directory was passed
		// ------------------------
		pathProjectDir = pathCommandLineArg
		pathProjectFile = path.Join(pathProjectDir, projectFilename)
		if !pathExists(pathProjectFile) {
			return fmt.Errorf("project file does not exist: %q", pathProjectFile)
		}
		if pathIsDir(pathProjectFile) {
			return fmt.Errorf("project file is a directory, must be a .json file: %q", pathProjectFile)
		}
	} else {
		// ------------------------
		// File was passed
		// ------------------------
		if path.Ext(pathCommandLineArg) != ".json" {
			return fmt.Errorf("project file must be a .json file: %q", pathCommandLineArg)
		}
		pathProjectDir = path.Dir(pathCommandLineArg)
		pathProjectFile = pathCommandLineArg
	}
	qlog.InfoPrintf("Project dir: %q", pathProjectDir)
	qlog.InfoPrintf("Project file: %q", pathProjectFile)

	// Load configProject config
	configProject, err := loadConfigElectroProject(pathProjectFile)
	if err != nil {
		return fmt.Errorf("error loading project config: %w", err)
	}

	qlog.Infof("Project config: %#v", configProject)

	// -----------------------
	// Determine output directory
	// -----------------------
	if configProject.OutputDirectory == "" {
		return fmt.Errorf("output directory not specified in project config")
	}
	pathOutputDir := filepath.Join(pathProjectDir, configProject.OutputDirectory)
	if !pathIsDir(pathOutputDir) {
		return fmt.Errorf("output directory does not exist: %q", pathOutputDir)
	}
	qlog.InfoPrintf("Using output directory: %q", pathOutputDir)

	// -----------------------
	// Determine theme dir
	// -----------------------
	pathThemeDirectory := filepath.Join(pathDirThemes, configProject.Theme)
	if !pathIsDir(pathThemeDirectory) {
		return fmt.Errorf("theme directory does not exist: %q", pathThemeDirectory)
	}
	qlog.InfoPrintf("Using theme: %q", configProject.Theme)

	// -----------------------
	// Build project
	// -----------------------
	builder := newBuilder(pathOutputDir, pathProjectDir, pathThemeDirectory)

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func pathIsDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
