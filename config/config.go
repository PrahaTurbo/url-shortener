package config

import (
	"flag"
	"log"
	"net"
)

type Config struct {
	Host    string
	Port    string
	BaseURL string
}

func Load() Config {
	var c Config

	addr := flag.String("a", "localhost:8080", "input server address in a form host:port")
	baseURL := flag.String("b", "http://localhost:8080", "base address for short url")
	flag.Parse()

	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatal("error occured while parsing server address: ", err)
	}

	c.Host = host
	c.Port = port
	c.BaseURL = *baseURL

	return c
}
