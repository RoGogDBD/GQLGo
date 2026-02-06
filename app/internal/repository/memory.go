package repository

import (
	"sync"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

type (
	MemoryUserRepo struct {
		// mu sync.Mutex
		mu    sync.RWMutex
		users map[string]*models.User
	}
	PostMemoryRepo struct {
		mu    sync.RWMutex
		posts map[string]*models.Post
		order []string
	}
	CommentMemoryRepo struct {
		mu       sync.RWMutex
		comments map[string]*models.Comment
		Post     map[string][]string
		Parent   map[string][]string
	}
)

// ==================== Конструкторы ====================
func NewMemoryUserRepo() *MemoryUserRepo {
	return &MemoryUserRepo{
		users: make(map[string]*models.User),
	}
}
func NewMemoryPostRepo() *PostMemoryRepo {
	return &PostMemoryRepo{
		posts: make(map[string]*models.Post),
		order: make([]string, 0),
	}
}
func NewMemoryCommentRepo() *CommentMemoryRepo {
	return &CommentMemoryRepo{
		comments: make(map[string]*models.Comment),
		Post:     make(map[string][]string),
		Parent:   make(map[string][]string),
	}
}

// ==================== UserRepo ====================
