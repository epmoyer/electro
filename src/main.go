package main

import (
	"app/pkg/electro"
	"flag"
	"fmt"
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

	versionInfo := config.AppName + " " + config.Version
	fmt.Println(versionInfo)
	fmt.Printf("    electro core package version: %s\n", electro.Version)
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
	err = electro.Init(flagNoEmbed)
	if err != nil {
		qlog.InfoPrint("Initialization error: " + err.Error())
		return
	}

	// ------------------------
	// Start
	// ------------------------
	err = electro.BuildProject(flagProject)
	if err != nil {
		qlog.ErrorPrint("Build error: " + err.Error())
		return
	}
	qlog.InfoPrint("Done")

}
