package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/RoGogDBD/GQLGo/internal/repository"
)

type PostService struct {
	repo repository.PostRepo
}

func NewPostService(repo repository.PostRepo) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) Create(ctx context.Context, in models.CreatePostInput) (*models.Post, error) {
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

	return s.repo.Create(ctx, in)
}
