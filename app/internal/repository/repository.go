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
)
