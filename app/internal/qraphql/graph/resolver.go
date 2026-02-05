//go:generate go tool gqlgen generate

package graph

import (
	"github.com/RoGogDBD/GQLGo/internal/repository"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require
// here.

//// Test data.
//var data = []*models.User{
//	{ID: "1", Username: "Vasa"},
//	{ID: "2", Username: "Petya"},
//	{ID: "3", Username: "Slon"},
//}

type Resolver struct {
	UserRepo    repository.UserRepo
	PostRepo    repository.PostRepo
	CommentRepo repository.CommentRepo
}
