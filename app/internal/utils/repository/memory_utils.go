package repository

import (
	"sort"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

const defaultPageSize = 10

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

func PaginateIDs(ids []string, after *string, first int32) []string {
	if first <= 0 {
		first = defaultPageSize
	}

	start := 0
	if after != nil && *after != "" {
		for i, id := range ids {
			if id == *after {
				start = i + 1
				break
			}
		}
	}

	end := start + int(first)
	if end > len(ids) {
		end = len(ids)
	}
	if start > len(ids) {
		start = len(ids)
	}
	return ids[start:end]
}

func RemoveID(ids []string, id string) []string {
	for i := 0; i < len(ids); i++ {
		if ids[i] == id {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}
