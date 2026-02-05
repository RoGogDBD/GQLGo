package config

import (
	"errors"
)

var (
	ErrNoDSN     = errors.New("dsn не установлен")
	ErrNoAddress = errors.New("addr не установлен")
)

// Validate проверяет на параметры кофига.
func (c Config) Validate() error {
	var errs []error

	if c.DB.DSN == "" {
		errs = append(errs, ErrNoDSN)
	}
	if c.Server.Addr == "" {
		errs = append(errs, ErrNoAddress)
	}

	return errors.Join(errs...)
}

// TODO: Можно добавить более строгую валидацию параметров.
