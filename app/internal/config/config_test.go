package config

import (
	"errors"
	"testing"
)

// Тест конфиг из env.
func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantDSN      string
		wantAddr     string
		wantPostgres bool
	}{
		{
			name: "Все переменные",
			env: map[string]string{
				"DSN":          "postgres://user:pass@db:5432/app?sslmode=disable",
				"ADDR":         "0.0.0.0:8080",
				"USE_POSTGRES": "true",
			},
			wantDSN:      "postgres://user:pass@db:5432/app?sslmode=disable",
			wantAddr:     "0.0.0.0:8080",
			wantPostgres: true,
		},
		{
			name:         "Обязательные",
			env:          map[string]string{"DSN": "dsn"},
			wantDSN:      "dsn",
			wantAddr:     "localhost:8080",
			wantPostgres: false,
		},
		{
			name:         "USE_POSTGRES выключен",
			env:          map[string]string{"DSN": "dsn", "USE_POSTGRES": "false"},
			wantDSN:      "dsn",
			wantAddr:     "localhost:8080",
			wantPostgres: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg := LoadFromEnv()
			if cfg.DB.DSN != tc.wantDSN {
				t.Fatalf("ожидался DSN %q, а получили %q", tc.wantDSN, cfg.DB.DSN)
			}
			if cfg.Server.Addr != tc.wantAddr {
				t.Fatalf("ожидался ADDR %q, а получили %q", tc.wantAddr, cfg.Server.Addr)
			}
			if cfg.UsePostgres != tc.wantPostgres {
				t.Fatalf("ожидался UsePostgres %v, а получили %v", tc.wantPostgres, cfg.UsePostgres)
			}
		})
	}
}

// Тест на валидацию обязательных полей.
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantNoDSN bool
		wantNoAdr bool
	}{
		{
			name: "ОК",
			cfg: Config{
				Server: ServerConfig{Addr: "0.0.0.0:8080"},
				DB:     DataBase{DSN: "dsn"},
			},
		},
		{
			name:      "Нету DSN",
			cfg:       Config{Server: ServerConfig{Addr: "0.0.0.0:8080"}},
			wantNoDSN: true,
		},
		{
			name:      "Нету ADDR",
			cfg:       Config{DB: DataBase{DSN: "dsn"}},
			wantNoAdr: true,
		},
		{
			name:      "Нету DSN и ADDR",
			cfg:       Config{},
			wantNoDSN: true,
			wantNoAdr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantNoDSN && !errors.Is(err, ErrNoDSN) {
				t.Fatalf("ожидалось ErrNoDSN, а получили %v", err)
			}
			if tc.wantNoAdr && !errors.Is(err, ErrNoAddress) {
				t.Fatalf("ожидалось ErrNoAddress, а получили %v", err)
			}
			if !tc.wantNoDSN && !tc.wantNoAdr && err != nil {
				t.Fatalf("ошибка не ожидалась: %v", err)
			}
		})
	}
}
