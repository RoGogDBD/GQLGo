package config

import (
	"errors"
	"testing"
)

// Тест конфиг из env.
func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		DSN      string
		Addr     string
		Postgres bool
	}{
		{
			name: "Все переменные",
			env: map[string]string{
				"DSN":          "postgres://user:pass@db:5432/app?sslmode=disable",
				"ADDR":         "0.0.0.0:8080",
				"USE_POSTGRES": "true",
			},
			DSN:      "postgres://user:pass@db:5432/app?sslmode=disable",
			Addr:     "0.0.0.0:8080",
			Postgres: true,
		},
		{
			name:     "Обязательные",
			env:      map[string]string{"DSN": "dsn"},
			DSN:      "dsn",
			Addr:     "localhost:8080",
			Postgres: false,
		},
		{
			name:     "USE_POSTGRES выключен",
			env:      map[string]string{"DSN": "dsn", "USE_POSTGRES": "false"},
			DSN:      "dsn",
			Addr:     "localhost:8080",
			Postgres: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg := LoadFromEnv()
			if cfg.DB.DSN != tc.DSN {
				t.Fatalf("ожидался DSN %q, а получили %q", tc.DSN, cfg.DB.DSN)
			}
			if cfg.Server.Addr != tc.Addr {
				t.Fatalf("ожидался ADDR %q, а получили %q", tc.Addr, cfg.Server.Addr)
			}
			if cfg.UsePostgres != tc.Postgres {
				t.Fatalf("ожидался UsePostgres %v, а получили %v", tc.Postgres, cfg.UsePostgres)
			}
		})
	}
}

// Тест на валидацию обязательных полей.
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name  string
		cfg   Config
		NoDSN bool
		NoAdr bool
	}{
		{
			name: "ОК",
			cfg: Config{
				Server: ServerConfig{Addr: "0.0.0.0:8080"},
				DB:     DataBase{DSN: "dsn"},
			},
		},
		{
			name:  "Нету DSN",
			cfg:   Config{Server: ServerConfig{Addr: "0.0.0.0:8080"}},
			NoDSN: true,
		},
		{
			name:  "Нету ADDR",
			cfg:   Config{DB: DataBase{DSN: "dsn"}},
			NoAdr: true,
		},
		{
			name:  "Нету DSN и ADDR",
			cfg:   Config{},
			NoDSN: true,
			NoAdr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.NoDSN && !errors.Is(err, ErrNoDSN) {
				t.Fatalf("ожидалось ErrNoDSN, а получили %v", err)
			}
			if tc.NoAdr && !errors.Is(err, ErrNoAddress) {
				t.Fatalf("ожидалось ErrNoAddress, а получили %v", err)
			}
			if !tc.NoDSN && !tc.NoAdr && err != nil {
				t.Fatalf("ошибка не ожидалась: %v", err)
			}
		})
	}
}
