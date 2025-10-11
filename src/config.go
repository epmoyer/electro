package main

var config = configT{
	AppTitle: "Electro",
	AppName:  "electro",
	Version:  "0.0.1",
}

type configT struct {
	AppTitle string
	AppName  string
	Version  string
}
