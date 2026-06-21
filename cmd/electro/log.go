package main

import (
	"fmt"
	"github.com/epmoyer/quicklog/v2"
	"os"
	"strings"
)

var qlog *quicklog.LoggerT = nil // Assigned at runtime

func initLogging() error {
	if flagLogDir != "" {
		pathDirLogs = flagLogDir
	}

	isDisabled := false
	if flagEnableLogging {
		isDisabled = false

		// Disable logEnabled if pathDirLogs does not exist
		if _, err := os.Stat(pathDirLogs); os.IsNotExist(err) {
			isDisabled = true
			return fmt.Errorf("log directory does not exist: %q", pathDirLogs)
		}
	}
	if isDisabled {
		fmt.Println("Logging is disabled.")
	} else {
		fmt.Printf("Logging to: %q\n", pathDirLogs)
	}

	// ------------------------
	// Start logger
	// ------------------------
	thisComputerName := getComputerName()
	loggingConfig := quicklog.ConfigT{
		LoggerId:   "default",
		Directory:  pathDirLogs,
		Filename:   thisComputerName + "." + config.AppName + ".log",
		Level:      quicklog.LogLevelTrace,
		MaxSize:    50,
		MaxBackups: 5,
		IsDisabled: isDisabled,
	}
	qlog = quicklog.ConfigureLogger(loggingConfig)
	qlog.Info(config.AppName + " v" + config.Version)

	return nil
}

func getComputerName() string {
	HostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	computerName := strings.Replace(HostName, ".local", "", 1)
	fmt.Printf("HostName:%#v computerName:%#v\n", HostName, computerName)
	return computerName
}
