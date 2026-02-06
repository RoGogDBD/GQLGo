package repository

import (
	"sort"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

func CloneUser(u *models.User) *models.User {
	if u == nil {
		return nil
	}
	c := *u
	return &c
}

func ClonePost(p *models.Post) *models.Post {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}

func SortedKeys[T any](m map[string]T) []string {
	idList := make([]string, 0, len(m))
	for id := range m {
		idList = append(idList, id)
	}
	sort.Strings(idList)
	return idList
}

func SortCommentIDs(ids []string, comments map[string]*models.Comment, order models.CommentOrder) {
	oldest := order == models.CommentOrderOldest
	sort.Slice(ids, func(i, j int) bool {
		a := comments[ids[i]]
		b := comments[ids[j]]

		if a.CreatedAt.Equal(b.CreatedAt) {
			if oldest {
				return a.ID < b.ID
			}
			return a.ID > b.ID
		}
		if oldest {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		return a.CreatedAt.After(b.CreatedAt)
	})
}
