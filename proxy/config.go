package proxy

var NewConfig = Config{
	Local: Network{
		Host: "0.0.0.0",
		Port: 25566,
	},
	Remote: Network{
		Host: "127.0.0.1",
		Port: 25565,
	},
}

type Config struct {
	Local  Network
	Remote Network
}

type Network struct {
	Host string
	Port int
}
