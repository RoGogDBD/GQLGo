package config

import "os"

type Config struct {
	Server ServerConfig
	DB     DataBase
}

type (
	// ServerConfig конфиг серва.
	ServerConfig struct {
		Addr string
	}
	// DataBase конфиг бд.
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
	}

	if v := os.Getenv("DSN"); v != "" {
		cfg.DB.DSN = v
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Server.Addr = v
	}

	return cfg
}
