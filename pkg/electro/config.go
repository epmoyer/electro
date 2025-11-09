package electro

var config = configT{
	AppName: "electro",
	Version: "3.0.0",
}

type configT struct {
	AppName string
	Version string
}
