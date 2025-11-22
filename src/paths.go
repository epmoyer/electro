package main

import "path"

var pathDirData = "data"

// NOTE: The logs dir is used only for development, when the `-log` command line flag is passed.
// In production, logs messages are quietly discarded.  Generally in production the logs
// dir does not exist, because the data dir is an embedded filesystem so we don't write to it.
// It is however possible to log in production by setting the -logdir command line flag.
var pathDirLogs = path.Join(pathDirData, "logs")
