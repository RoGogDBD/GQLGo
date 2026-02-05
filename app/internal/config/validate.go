package config

import "fmt"

// Validate проверяет на наличие значений.
func (c Config) Validate() error {
	if c.DB.DSN == "" {
		return fmt.Errorf("DB.DSN отсутсвует")
	}
	if c.Server.Addr == "" {
		return fmt.Errorf("SERVER.ADDR отсутсвует")
	}
	return nil
}

// TODO: Можно добавить более строгую валидацию параметров.
