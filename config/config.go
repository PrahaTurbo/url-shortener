package config

import (
	"flag"
	"os"
)

// Config represents the configuration of the application.
type Config struct {
	Addr            string // The server address, in the form host:port.
	BaseURL         string // The base URL to which the server responds.
	LogLevel        string // The level of logs that should be displayed. Options include "info", "error", and "debug".
	StorageFilePath string // The path to the file where the server will store short URL data.
	DatabaseDSN     string // The SQL database DSN (Data Source Name) to connect to the database.
	JWTSecret       string // The secret key used in JWT for authentication.
}

// Load reads command-line flags and environment variables to populate a Config object.
func Load() Config {
	var c Config

	addr := flag.String("a", "localhost:8080", "input server address in a form host:port")
	baseURL := flag.String("b", "http://localhost:8080", "base address for short url")
	logLevel := flag.String("l", "info", "log lever")
	storageFilePath := flag.String("f", "/tmp/short-url-db.json", "path to storage file")
	databaseDSN := flag.String("d", "", "sql database dsn")
	flag.Parse()

	c.Addr = *addr
	c.BaseURL = *baseURL
	c.LogLevel = *logLevel
	c.StorageFilePath = *storageFilePath
	c.DatabaseDSN = *databaseDSN

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

	if envStorageFilePath := os.Getenv("FILE_STORAGE_PATH"); envStorageFilePath != "" {
		c.StorageFilePath = envStorageFilePath
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		c.DatabaseDSN = envDatabaseDSN
	}

	if envJWTSecret := os.Getenv("JWT_SECRET_KEY"); envJWTSecret != "" {
		c.JWTSecret = envJWTSecret
	} else {
		c.JWTSecret = "key_for_studying"
	}
}
