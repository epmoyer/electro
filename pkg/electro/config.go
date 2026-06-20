package electro

const Version = "v3.10.0"

var config = configT{
	AppName: "electro",
	Version: Version,
}

type configT struct {
	AppName string
	Version string
}
