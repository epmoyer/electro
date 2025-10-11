package main

var config = configT{
	AppTitle:        "Electro",
	AppName:         "electro",
	Version:         "0.0.1",
	ProjectFilename: "electro.json",
}

type configT struct {
	AppTitle        string
	AppName         string
	Version         string
	ProjectFilename string
}
