package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var flagVersion bool
var flagProject string // Project dir path or project json config file path
var flagLogDir string
var flagNoEmbed bool
var flagEnableLogging bool

func main() {

	flag.BoolVar(&flagVersion, "version", false, "Print the version and exit")
	flag.StringVar(&flagProject, "project", ".", "Project dir path or project json config file path")
	flag.StringVar(&flagLogDir, "logdir", "", "Log directory")
	flag.BoolVar(&flagNoEmbed, "noembed", false, "Do not use embedded filesystem")
	flag.BoolVar(&flagEnableLogging, "log", false, "Enable logging")
	flag.Parse()

	versionInfo := config.AppName + " v" + config.Version
	fmt.Println(versionInfo)
	if flagVersion {
		return
	}
	err := initLogging()
	if err != nil {
		fmt.Printf("ERROR: Error initializing logging: %s\n", err.Error())
		return
	}

	if flagNoEmbed {
		qlog.InfoPrint("🟣 Using local filesystem for content.")
	} else {
		qlog.InfoPrint("Using embedded filesystem for content.")
	}

	// ------------------------
	// Initialize local packages
	// ------------------------
	// FIXME: TBD

	// ------------------------
	// Start
	// ------------------------
	fmt.Println("🔴  Implementation TBD")
	err = buildProject(flagProject)
	if err != nil {
		qlog.ErrorPrint("Build error: " + err.Error())
		return
	}
	qlog.InfoPrint("Done")

}

func buildProject(pathCommandLineArg string) error {
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
		pathProjectFile = path.Join(pathProjectDir, config.ProjectFilename)
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
