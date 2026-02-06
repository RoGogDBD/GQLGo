package models

import "time"

// ============================== COMMENTS ==============================
type (
	AddCommentInput struct {
		PostID   string  `json:"postId"`
		AuthorID string  `json:"authorId"`
		ParentID *string `json:"parentId,omitempty"`
		Body     string  `json:"body"`
	}

	Comment struct {
		ID            string             `json:"id"`
		PostID        string             `json:"postId"`
		Post          *Post              `json:"post"`
		Author        *User              `json:"author"`
		Body          string             `json:"body"`
		ParentID      *string            `json:"parentId,omitempty"`
		Depth         int32              `json:"depth"`
		ChildrenCount int32              `json:"childrenCount"`
		Children      *CommentConnection `json:"children"`
		CreatedAt     time.Time          `json:"-"`
	}

	CommentConnection struct {
		Edges      []*CommentEdge `json:"edges"`
		PageInfo   *PageInfo      `json:"pageInfo"`
		TotalCount int32          `json:"totalCount"`
	}

	CommentEdge struct {
		Cursor string   `json:"cursor"`
		Node   *Comment `json:"node"`
	}
)

// ============================== PAGINATION ==============================
type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

// ============================== POSTS ==============================
type (
	Post struct {
		ID              string             `json:"id"`
		Title           string             `json:"title"`
		Body            string             `json:"body"`
		Author          *User              `json:"author"`
		CommentsEnabled bool               `json:"commentsEnabled"`
		Comments        *CommentConnection `json:"comments"`
	}

	CreatePostInput struct {
		AuthorID        string `json:"authorId"`
		Title           string `json:"title"`
		Body            string `json:"body"`
		CommentsEnabled *bool  `json:"commentsEnabled,omitempty"`
	}

	PostConnection struct {
		Edges      []*PostEdge `json:"edges"`
		PageInfo   *PageInfo   `json:"pageInfo"`
		TotalCount int32       `json:"totalCount"`
	}

	PostEdge struct {
		Cursor string `json:"cursor"`
		Node   *Post  `json:"node"`
	}
)

// ============================== USERS ==============================
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
