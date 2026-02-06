package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RoGogDBD/GQLGo/internal/config/migrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Параметры для конфигурации к бд.
const (
	contextTimeout  = 5 * time.Second
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 5 * time.Minute

	maxOpenConns = 25
	maxIdleConns = 10
)

// DBStorage хранилище с подключением к БД.
type DBStorage struct {
	sqldb *sql.DB
	db    *bun.DB
}

// NewDataStorage создает подключение к БД.
func NewDataStorage(dsn string) (*DBStorage, error) {
	if err := migrate.RunMigrations(dsn); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Конфиг.
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxLifetime(connMaxLifetime)
	sqldb.SetConnMaxIdleTime(connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	return &DBStorage{
		sqldb: sqldb,
		db:    db,
	}, nil
}

// Close закрывает соединение с БД.
func (s *DBStorage) Close() error {
	return s.sqldb.Close()
}

// DB возвращает объект БД.
func (s *DBStorage) DB() *bun.DB {
	return s.db
}
