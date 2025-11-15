package electro

const Version = "3.0.0"

var config = configT{
	AppName: "electro",
	Version: Version,
}

type configT struct {
	AppName string
	Version string
}
