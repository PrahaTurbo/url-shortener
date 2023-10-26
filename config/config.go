package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

// Config represents the configuration of the application.
type Config struct {
	Addr            string `json:"server_address"`    // The server address, in the form host:port.
	BaseURL         string `json:"base_url"`          // The base URL to which the server responds.
	LogLevel        string `json:"log_level"`         // The level of logs that should be displayed. Options include "info", "error", and "debug".
	StorageFilePath string `json:"file_storage_path"` // The path to the file where the server will store short URL data.
	DatabaseDSN     string `json:"database_dsn"`      // The SQL database DSN (Data Source Name) to connect to the database.
	JWTSecret       string // The secret key used in JWT for authentication.
	TrustedSubnet   string `json:"trusted_subnet"` // Trusted subnet
	EnableHTTPS     bool   `json:"enable_https"`   // Enable HTTPS on server
}

// Load reads command-line flags and environment variables to populate a Config object.
func Load() Config {
	var c Config

	addr := flag.String("a", "localhost:8080", "input server address in a form host:port")
	baseURL := flag.String("b", "http://localhost:8080", "base address for short url")
	logLevel := flag.String("l", "info", "log lever")
	storageFilePath := flag.String("f", "/tmp/short-url-db.json", "path to storage file")
	databaseDSN := flag.String("d", "", "sql database dsn")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS on server")
	configPath := flag.String("c", "", "path to config file")
	trustedSubnet := flag.String("t", "", "trusted subnet")
	flag.Parse()

	if err := c.loadJSON(*configPath); err != nil {
		log.Println("can't load config file: ", err)
	}

	c.Addr = *addr
	c.BaseURL = *baseURL
	c.LogLevel = *logLevel
	c.StorageFilePath = *storageFilePath
	c.DatabaseDSN = *databaseDSN
	c.EnableHTTPS = *enableHTTPS
	c.TrustedSubnet = *trustedSubnet

	c.loadEnvVars()

	return c
}

func (c *Config) loadJSON(path string) error {
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		path = envConfig
	}

	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
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

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		val, err := strconv.ParseBool(envEnableHTTPS)
		if err != nil {
			log.Fatal(err)
		}

		c.EnableHTTPS = val
	}

	if envJWTSecret := os.Getenv("JWT_SECRET_KEY"); envJWTSecret != "" {
		c.JWTSecret = envJWTSecret
	} else {
		c.JWTSecret = "key_for_studying"
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		c.TrustedSubnet = envTrustedSubnet
	}
}
