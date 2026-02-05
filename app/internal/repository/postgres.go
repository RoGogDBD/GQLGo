package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/RoGogDBD/GQLGo/internal/utils/repository"
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

	PostgresCommentRepo struct {
		db *bun.DB
	}
)

type commentInsertRow struct {
	bun.BaseModel `bun:"table:comments"`

	ID            string    `bun:"id"`
	PostID        string    `bun:"post_id"`
	AuthorID      string    `bun:"author_id"`
	ParentID      *string   `bun:"parent_id"`
	Body          string    `bun:"body"`
	Depth         int       `bun:"depth"`
	ChildrenCount int       `bun:"children_count"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}

func NewPostgresUserRepo(db *bun.DB) (*PostgresUserRepo, error) {
	return &PostgresUserRepo{db: db}, nil
}
func NewPostgresPostRepo(db *bun.DB) (*PostgresPostRepo, error) {
	return &PostgresPostRepo{db: db}, nil
}
func NewPostgresCommentRepo(db *bun.DB) (*PostgresCommentRepo, error) {
	return &PostgresCommentRepo{db: db}, nil
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

	repository.ApplyAfterByID(query, after, "id")
	if err := query.Scan(ctx); err != nil {
		return nil, nil, fmt.Errorf("список юзеров: %w", err)
	}

	return users, repository.LastID(users, func(u *models.User) string { return u.ID }), nil
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

	repository.ApplyAfterByID(query, after, "p.id")

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

	return posts, repository.LastID(posts, func(p *models.Post) string { return p.ID }), nil
}

// SetCommentsEnabled включает или выключает комментарии для поста.
func (r *PostgresPostRepo) SetCommentsEnabled(ctx context.Context, postID string, enabled bool) (*models.Post, error) {
	if postID == "" {
		return nil, fmt.Errorf("требуется id поста")
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

// ============================== COMMENT REPO ==============================

// GetMeta возвращает минимальные данные о комментарии.
func (r *PostgresCommentRepo) GetMeta(ctx context.Context, id string) (string, int, error) {
	var meta struct {
		PostID string `bun:"post_id"`
		Depth  int    `bun:"depth"`
	}

	err := r.db.NewSelect().
		Table("comments").
		Column("post_id", "depth").
		Where("id = ?", id).
		Scan(ctx, &meta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, sql.ErrNoRows
		}
		return "", 0, fmt.Errorf("получение комментария: %w", err)
	}

	return meta.PostID, meta.Depth, nil
}

// Create создает комментарий и (если нужно) обновляет счетчик детей у родителя.
func (r *PostgresCommentRepo) Create(ctx context.Context, postID, authorID string, parentID *string, body string, depth int) (*models.Comment, error) {
	id := uuid.NewString()
	now := time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row := &commentInsertRow{
		ID:            id,
		PostID:        postID,
		AuthorID:      authorID,
		ParentID:      parentID,
		Body:          body,
		Depth:         depth,
		ChildrenCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err = tx.NewInsert().
		Model(row).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание комментария: %w", err)
	}

	if parentID != nil && *parentID != "" {
		_, err = tx.NewUpdate().
			Table("comments").
			Set("children_count = children_count + 1").
			Set("updated_at = ?", now).
			Where("id = ?", *parentID).
			Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("обновление родителя: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &models.Comment{
		ID:            id,
		PostID:        postID,
		Author:        &models.User{ID: authorID},
		Body:          body,
		ParentID:      parentID,
		Depth:         int32(depth),
		ChildrenCount: 0,
		Children: &models.CommentConnection{
			Edges:      []*models.CommentEdge{},
			PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: nil},
			TotalCount: 0,
		},
		CreatedAt: now,
	}, nil
}

func (r *PostgresCommentRepo) ListByParent(ctx context.Context, postID string, parentID *string, first int32, after *string, order models.CommentOrder) ([]*models.Comment, *string, error) {
	if postID == "" {
		return nil, nil, fmt.Errorf("требуется id поста")
	}
	if first <= 0 {
		first = defaultPageSize
	}
	if !order.IsValid() {
		order = models.CommentOrderNewest
	}

	comments := make([]*models.Comment, 0, first)

	query := r.db.NewSelect().
		TableExpr("comments AS c").
		Column(
			"c.id",
			"c.post_id",
			"c.parent_id",
			"c.body",
			"c.depth",
			"c.children_count",
			"c.created_at",
		).
		ColumnExpr("u.id AS author__id, u.username AS author__username").
		Join("JOIN users AS u ON u.id = c.author_id").
		Where("c.post_id = ?", postID).
		Limit(int(first))

	if parentID == nil || *parentID == "" {
		query.Where("c.parent_id IS NULL")
	} else {
		query.Where("c.parent_id = ?", *parentID)
	}

	if after != nil && *after != "" {
		switch order {
		case models.CommentOrderNewest:
			query.Where("(c.created_at, c.id) < (SELECT created_at, id FROM comments WHERE id = ?)", *after)
		case models.CommentOrderOldest:
			query.Where("(c.created_at, c.id) > (SELECT created_at, id FROM comments WHERE id = ?)", *after)
		}
	}

	switch order {
	case models.CommentOrderNewest:
		query.Order("c.created_at DESC", "c.id DESC")
	case models.CommentOrderOldest:
		query.Order("c.created_at ASC", "c.id ASC")
	}

	if err := query.Scan(ctx, &comments); err != nil {
		return nil, nil, fmt.Errorf("список комментариев: %w", err)
	}

	for _, c := range comments {
		if c.Children == nil {
			c.Children = &models.CommentConnection{
				Edges:      []*models.CommentEdge{},
				PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: nil},
				TotalCount: 0,
			}
		}
	}

	return comments, repository.LastID(comments, func(c *models.Comment) string { return c.ID }), nil
}
