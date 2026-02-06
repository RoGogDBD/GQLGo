package graph

import (
	"context"
	"fmt"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

type (
	// CommentLister получение комментариев.
	CommentLister interface {
		ListByParent(ctx context.Context, postID string, parentID *string, first int32, after *string, order models.CommentOrder) ([]*models.Comment, *string, error)
	}

	// CommentMetaGetter получение мета-данных комментария.
	CommentMetaGetter interface {
		GetMeta(ctx context.Context, id string) (string, int, error)
	}
)

// NewPostConnection создает объект PostConnection.
func NewPostConnection(list []*models.Post, hasNext bool) *models.PostConnection {
	edges := make([]*models.PostEdge, 0, len(list))
	for _, p := range list {
		edges = append(edges, &models.PostEdge{
			Cursor: p.ID,
			Node:   p,
		})
	}

	var endCursor *string
	if len(list) > 0 {
		id := list[len(list)-1].ID
		endCursor = &id
	}
	return &models.PostConnection{
		Edges:      edges,
		PageInfo:   &models.PageInfo{HasNextPage: hasNext, EndCursor: endCursor},
		TotalCount: int32(len(edges)),
	}
}

// NewUserConnection создает объект UserConnection.
func NewUserConnection(list []*models.User, hasNext bool) *models.UserConnection {
	edges := make([]*models.UserEdge, 0, len(list))
	for _, u := range list {
		edges = append(edges, &models.UserEdge{
			Cursor: u.ID,
			Node:   u,
		})
	}
	var endCursor *string
	if len(list) > 0 {
		id := list[len(list)-1].ID
		endCursor = &id
	}
	return &models.UserConnection{
		Edges:      edges,
		PageInfo:   &models.PageInfo{HasNextPage: hasNext, EndCursor: endCursor},
		TotalCount: int32(len(edges)),
	}
}

// NewCommentConnection создает CommentConnection.
func NewCommentConnection(list []*models.Comment, hasNext bool) *models.CommentConnection {
	edges := make([]*models.CommentEdge, 0, len(list))
	for _, c := range list {
		edges = append(edges, &models.CommentEdge{
			Cursor: c.ID,
			Node:   c,
		})
	}
	var endCursor *string
	if len(list) > 0 {
		id := list[len(list)-1].ID
		endCursor = &id
	}
	return &models.CommentConnection{
		Edges:      edges,
		PageInfo:   &models.PageInfo{HasNextPage: hasNext, EndCursor: endCursor},
		TotalCount: int32(len(edges)),
	}
}

// ResolveCommentConnection применяет пагинацию и собирает CommentConnection.
func ResolveCommentConnection(ctx context.Context, repo CommentLister, postID string, parentID *string, first *int32, after *string, order *models.CommentOrder, defaultOrder models.CommentOrder) (*models.CommentConnection, error) {
	f := int32(20)
	if first != nil {
		f = *first
	}
	ord := defaultOrder
	if order != nil {
		ord = *order
	}

	list, _, err := repo.ListByParent(ctx, postID, parentID, f+1, after, ord)
	if err != nil {
		return nil, err
	}
	hasNext := int32(len(list)) > f
	if hasNext {
		list = list[:f]
	}
	return NewCommentConnection(list, hasNext), nil
}

// ResolveCommentDepth глубина нового комментария, относительно родителя.
func ResolveCommentDepth(ctx context.Context, repo CommentMetaGetter, postID string, parentID *string) (int, error) {
	if parentID == nil || *parentID == "" {
		return 0, nil
	}

	parentPostID, parentDepth, err := repo.GetMeta(ctx, *parentID)
	if err != nil {
		return 0, err
	}
	if parentPostID != postID {
		return 0, fmt.Errorf("родитель из другого поста")
	}
	return parentDepth + 1, nil
}
