package graph

import "github.com/RoGogDBD/GQLGo/internal/models"

// NewPostConnection создает объект PostConnection.
func NewPostConnection(list []*models.Post, endCursor *string) *models.PostConnection {
	edges := make([]*models.PostEdge, 0, len(list))
	for _, p := range list {
		edges = append(edges, &models.PostEdge{
			Cursor: p.ID,
			Node:   p,
		})
	}
	return &models.PostConnection{
		Edges:      edges,
		PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: endCursor},
		TotalCount: int32(len(edges)),
	}
}

// NewUserConnection создает объект UserConnection.
func NewUserConnection(list []*models.User, endCursor *string) *models.UserConnection {
	edges := make([]*models.UserEdge, 0, len(list))
	for _, u := range list {
		edges = append(edges, &models.UserEdge{
			Cursor: u.ID,
			Node:   u,
		})
	}
	return &models.UserConnection{
		Edges:      edges,
		PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: endCursor},
		TotalCount: int32(len(edges)),
	}
}
