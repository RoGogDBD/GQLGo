package config

import (
	"errors"
	"testing"
)

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
