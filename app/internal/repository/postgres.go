package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/uptrace/bun"
)

const defaultPageSize = 10

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

// ============================== USER REPO ==============================

// GetByID возвращает пользователя по id.
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

// List возвращает список пользователй с пагинацией.
func (r *PostgresUserRepo) List(ctx context.Context, first int32, after *string) ([]*models.User, *string, error) {
	if first <= 0 {
		first = defaultPageSize
	}

	users := make([]*models.User, 0, first)

	query := r.db.NewSelect().
		Model(&users).
		Order("id ASC").
		Limit(int(first))

	// ===================== Проверки пагинации =====================
	if after != nil && *after != "" {
		query = query.Where("id > ?", *after)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, nil, fmt.Errorf("список юзеров: %w", err)
	}
	// ==============================================================

	var last *string
	if len(users) > 0 {
		c := users[len(users)-1].ID
		last = &c
	}

	return users, last, nil
}

// ============================== POST REPO ==============================

// GetByID возвращает пост по id.
func (r *PostgresPostRepo) GetByID(ctx context.Context, id string) (*models.Post, error) {
	p := new(models.Post)

	err := r.db.NewSelect().
		Model(p).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("получение поста: %w", err)
	}
	return p, nil
}
