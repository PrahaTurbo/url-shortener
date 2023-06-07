package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string
	BaseURL string
}

func Load() Config {
	var c Config

	addr := flag.String("a", "localhost:8080", "input server address in a form host:port")
	baseURL := flag.String("b", "http://localhost:8080", "base address for short url")
	flag.Parse()

	c.Addr = *addr
	c.BaseURL = *baseURL

	c.loadEnvVars()

	return c
}

func (c *Config) loadEnvVars() {
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		c.Addr = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURL = envBaseURL
	}
}
