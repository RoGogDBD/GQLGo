package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server      ServerConfig
	DB          DataBase
	UsePostgres bool
}

type (
	ServerConfig struct {
		Addr string
	}
	DataBase struct {
		DSN string
	}

	// TODO: Другие шняги JWT и т.д.
)

func LoadFromEnv() Config {
	cfg := Config{
		Server: ServerConfig{
			Addr: "localhost:8080",
		},
		UsePostgres: false,
	}

	if v := os.Getenv("DSN"); v != "" {
		cfg.DB.DSN = v
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := os.Getenv("USE_POSTGRES"); v != "" {
		cfg.UsePostgres = v == "true" || v == "1" || v == "yes" || v == "y"
	} else if v := os.Getenv("POSTGRES"); v != "" {
		cfg.UsePostgres = v == "true" || v == "1" || v == "yes" || v == "y"
	}

	return cfg
}

// Load конфиг из env и валидация.
func Load() (Config, error) {
	cfg := LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("загрузка конфига env: %w", err)
	}
	return cfg, nil
}
