package config

const (
	host = "localhost"
	port = "8080"
)

type Config struct {
	Host string
	Port string
}

func New() Config {
	return Config{Host: host, Port: port}
}
