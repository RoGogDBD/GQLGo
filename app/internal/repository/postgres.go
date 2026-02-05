package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/uptrace/bun"
)

type (
	PostgresUserRepo struct {
		db *bun.DB
	}

	PostgresPostRepo struct {
		db *bun.DB
	}
)

func NewPostgresUserRepo(db *bun.DB) (*PostgresUserRepo, error) {
	return &PostgresUserRepo{db: db}, nil
}

func NewPostgresPostRepo(db *bun.DB) (*PostgresPostRepo, error) {
	return &PostgresPostRepo{db: db}, nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	u := new(models.User)

	err := r.db.NewSelect().
		Model(u).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}
	return u, nil
}
