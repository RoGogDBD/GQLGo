package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/google/uuid"
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
	p.Author = &models.User{}

	err := r.db.NewSelect().
		TableExpr("posts AS p").
		Column("p.id", "p.title", "p.body", "p.comments_enabled").
		ColumnExpr("u.id AS author__id, u.username AS author__username").
		Join("JOIN users AS u ON u.id = p.author_id").
		Where("p.id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("получение поста: %w", err)
	}
	p.Comments = &models.CommentConnection{
		Edges:      []*models.CommentEdge{},
		PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: nil},
		TotalCount: 0,
	}
	return p, nil
}

func (r *PostgresPostRepo) Create(ctx context.Context, in models.CreatePostInput) (*models.Post, error) {
	// ===================== Валидация входных данных =====================
	if in.AuthorID == "" {
		return nil, fmt.Errorf("требуется id автора")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("требуется заголовок")
	}
	if len(title) > 100 {
		return nil, fmt.Errorf("заголовок слишком длинный")
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, fmt.Errorf("требуется тело поста")
	}
	if len(body) > 2000 {
		return nil, fmt.Errorf("тело длинное (<= 2000 симв.)")
	}
	// ==============================================================

	commentsEnabled := true
	if in.CommentsEnabled != nil {
		commentsEnabled = *in.CommentsEnabled
	}

	id := uuid.NewString()

	_, err := r.db.NewRaw(`
		INSERT INTO posts (id, title, body, comments_enabled, author_id)
		VALUES (?, ?, ?, ?, ?)
	`, id, title, body, commentsEnabled, in.AuthorID).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание поста: %w", err)
	}

	return r.GetByID(ctx, id)
}

// List возвращает список постов с пагинацией.
func (r *PostgresPostRepo) List(ctx context.Context, first int32, after *string) ([]*models.Post, *string, error) {
	if first <= 0 {
		first = defaultPageSize
	}

	posts := make([]*models.Post, 0, first)

	query := r.db.NewSelect().
		TableExpr("posts AS p").
		Column("p.id", "p.title", "p.body", "p.comments_enabled").
		ColumnExpr("u.id AS author__id, u.username AS author__username").
		Join("JOIN users AS u ON u.id = p.author_id").
		Order("p.id ASC").
		Limit(int(first))

	if after != nil && *after != "" {
		query = query.Where("p.id > ?", *after)
	}

	if err := query.Scan(ctx, &posts); err != nil {
		return nil, nil, fmt.Errorf("список постов: %w", err)
	}

	for _, p := range posts {
		if p.Comments == nil {
			p.Comments = &models.CommentConnection{
				Edges:      []*models.CommentEdge{},
				PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: nil},
				TotalCount: 0,
			}
		}
	}

	var last *string
	if len(posts) > 0 {
		c := posts[len(posts)-1].ID
		last = &c
	}

	return posts, last, nil
}

// SetCommentsEnabled включает или выключает комментарии для поста.
func (r *PostgresPostRepo) SetCommentsEnabled(ctx context.Context, postID string, enabled bool) (*models.Post, error) {
	if postID == "" {
		return nil, fmt.Errorf("post id is required")
	}

	res, err := r.db.NewUpdate().
		Table("posts").
		Set("comments_enabled = ?", enabled).
		Set("updated_at = now()").
		Where("id = ?", postID).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление комментариев: %w", err)
	}

	if res != nil {
		if rows, err := res.RowsAffected(); err == nil && rows == 0 {
			return nil, nil
		}
	}

	return r.GetByID(ctx, postID)
}
