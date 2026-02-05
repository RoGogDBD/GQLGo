package repository

import (
	"context"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

type (
	UserRepo interface {
		GetByID(ctx context.Context, id string) (*models.User, error)
		List(ctx context.Context, first int32, after *string) ([]*models.User, *string, error)
	}

	PostRepo interface {
		GetByID(ctx context.Context, id string) (*models.Post, error)
		Create(ctx context.Context, in models.CreatePostInput) (*models.Post, error)
		List(ctx context.Context, first int32, after *string) ([]*models.Post, *string, error)
		SetCommentsEnabled(ctx context.Context, postID string, enabled bool) (*models.Post, error)
	}

	CommentRepo interface {
		GetMeta(ctx context.Context, id string) (postID string, depth int, err error)
		Create(ctx context.Context, postID, authorID string, parentID *string, body string, depth int) (*models.Comment, error)
		ListByParent(ctx context.Context, postID string, parentID *string, first int32, after *string, order models.CommentOrder) ([]*models.Comment, *string, error)
	}
)
