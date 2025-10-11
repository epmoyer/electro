package main

import (
	"flag"
	"fmt"
	"os"
	"path"
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

func buildProject(projectPath string) error {
	var pathProjectDir string

	// Fail if path does not exist
	if !pathExists(projectPath) {
		return fmt.Errorf("path does not exist: %q", projectPath)
	}
	if pathIsDir(projectPath) {
		// ------------------------
		// Directory was passed
		// ------------------------
		pathProjectDir = projectPath
	} else {
		// ------------------------
		// File was passed
		// ------------------------
		if path.Ext(projectPath) != ".json" {
			return fmt.Errorf("project file must be a .json file: %q", projectPath)
		}
		pathProjectDir = path.Dir(projectPath)
	}
	fmt.Printf("Project dir: %q\n", pathProjectDir)
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
