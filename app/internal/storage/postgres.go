package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/RoGogDBD/GQLGo/internal/config/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Параметры для подключения к бд.
const (
	contextTimeout  = 5 * time.Second
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 5 * time.Minute

	maxOpenConns int32 = 25
	minIdleConns int32 = 10
)

// DBStorage хранилище данных с подключением к БД.
type DBStorage struct {
	pool *pgxpool.Pool
}

// NewDataStorage создает новое подключение к БД.
func NewDataStorage(dsn string) (*DBStorage, error) {
	if err := migrate.RunMigrations(dsn); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool config: %w", err)
	}

	// Конфиг.
	cfg.MaxConns = maxOpenConns
	cfg.MinConns = minIdleConns
	cfg.MaxConnLifetime = connMaxLifetime
	cfg.MaxConnIdleTime = connMaxIdleTime

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool open: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return &DBStorage{pool: pool}, nil
}

// Close закрывает соединение с БД.
func (s *DBStorage) Close() error {
	s.pool.Close()
	return nil
}
