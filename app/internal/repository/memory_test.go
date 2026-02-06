package repository

import (
	"context"
	"testing"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

// Тест на создание поста и его погинацию.
func TestMemoryPostRepo_CreateList(t *testing.T) {
	tests := []struct {
		name      string
		createCnt int
		listFirst int32
		len       int
		cursor    bool
	}{
		{name: "Создание одного поста", createCnt: 1, listFirst: 10, len: 1, cursor: true},
		{name: "Без создания", createCnt: 0, listFirst: 10, len: 0, cursor: false},
		{name: "Создание трех постов", createCnt: 3, listFirst: 2, len: 2, cursor: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			storTTL := NewMemoryStorageWithTTL(0)
			repo := NewMemoryPostRepo(storTTL)

			for i := 0; i < tc.createCnt; i++ {
				_, err := repo.Create(context.Background(), models.CreatePostInput{
					AuthorID: "author",
					Title:    "title",
					Body:     "body",
				})
				if err != nil {
					t.Fatalf("при создинии поста: %v", err)
				}
			}

			list, cursor, err := repo.List(context.Background(), tc.listFirst, nil)
			if err != nil {
				t.Fatalf("список постов: %v", err)
			}
			if len(list) != tc.len {
				t.Fatalf("ожидалось %d постов, а получили %d", tc.len, len(list))
			}
			if tc.cursor && cursor == nil {
				t.Fatalf("ожидался cursor, got nil")
			}
			if !tc.cursor && cursor != nil {
				t.Fatalf("ожидалось nil, получили %v", cursor)
			}
		})
	}
}

// Тест на выборку корневых и дочерних комментариев.
func TestMemoryCommentRepo_ListByParent(t *testing.T) {
	tests := []struct {
		name        string
		createChild bool
		parentID    *string
		len         int
	}{
		{name: "Только корневой", createChild: false, parentID: nil, len: 1},
		{name: "Только дочерний", createChild: true, parentID: func() *string { id := "root"; return &id }(), len: 1},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			st := NewMemoryStorageWithTTL(0)
			repo := NewMemoryCommentRepo(st)

			root, err := repo.Create(context.Background(), "p", "u", nil, "root", 0)
			if err != nil {
				t.Fatalf("при создании корневого комментария: %v", err)
			}

			var parentID *string
			if tc.parentID != nil && *tc.parentID == "root" {
				parentID = &root.ID
			}

			if tc.createChild {
				if _, err := repo.Create(context.Background(), "p", "u", &root.ID, "child", 1); err != nil {
					t.Fatalf("при создании дочернего комментария: %v", err)
				}
			}

			list, _, err := repo.ListByParent(context.Background(), "p", parentID, 10, nil, models.CommentOrderNewest)
			if err != nil {
				t.Fatalf("список комментариев: %v", err)
			}
			if len(list) != tc.len {
				t.Fatalf("ожидалось %d комментариев, а получили %d", tc.len, len(list))
			}
		})
	}
}
