package service

import (
	"context"
	"strings"
	"testing"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

type postRepoStub struct {
	createCalled bool
}

func (s *postRepoStub) GetByID(context.Context, string) (*models.Post, error) {
	return nil, nil
}

func (s *postRepoStub) Create(_ context.Context, in models.CreatePostInput) (*models.Post, error) {
	s.createCalled = true
	return &models.Post{ID: "p1", Title: in.Title, Body: in.Body, Author: &models.User{ID: in.AuthorID}}, nil
}

func (s *postRepoStub) List(context.Context, int32, *string) ([]*models.Post, *string, error) {
	return nil, nil, nil
}

func (s *postRepoStub) SetCommentsEnabled(context.Context, string, bool) (*models.Post, error) {
	return nil, nil
}

// Тест на базовую валидацию входных данных при создании поста.
func TestPostService_Create_Table(t *testing.T) {
	tests := []struct {
		name  string
		input models.CreatePostInput
		err   bool
		call  bool
	}{
		{
			name:  "Нет автора",
			input: models.CreatePostInput{Title: "t", Body: "b"},
			err:   true,
		},
		{
			name:  "Пустой заголовок",
			input: models.CreatePostInput{AuthorID: "u", Title: "   ", Body: "b"},
			err:   true,
		},
		{
			name:  "Заголовок слишком длинный",
			input: models.CreatePostInput{AuthorID: "u", Title: strings.Repeat("a", 101), Body: "b"},
			err:   true,
		},
		{
			name:  "Пустое тело",
			input: models.CreatePostInput{AuthorID: "u", Title: "t", Body: "   "},
			err:   true,
		},
		{
			name:  "Тело слишком длинное",
			input: models.CreatePostInput{AuthorID: "u", Title: "t", Body: strings.Repeat("a", 2001)},
			err:   true,
		},
		{
			name:  "Успешное создание",
			input: models.CreatePostInput{AuthorID: "u", Title: "t", Body: "b"},
			err:   false,
			call:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &postRepoStub{}
			svc := NewPostService(repo)

			_, err := svc.Create(context.Background(), tc.input)
			if tc.err && err == nil {
				t.Fatalf("ожидалась - nil")
			}
			if !tc.err && err != nil {
				t.Fatalf("ошибка: %v", err)
			}
			if tc.call && !repo.createCalled {
				t.Fatalf("ожидался - repo.Create")
			}
			if !tc.call && repo.createCalled {
				t.Fatalf("ошибка - repo.Create")
			}
		})
	}
}
