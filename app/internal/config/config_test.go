package config

import (
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
