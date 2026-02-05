package repository

import (
	"fmt"

	"github.com/uptrace/bun"
)

// ApplyAfterByID пагинация по id.
func ApplyAfterByID(query *bun.SelectQuery, after *string, column string) {
	if after != nil && *after != "" {
		query.Where(fmt.Sprintf("%s > ?", column), *after)
	}
}

// LastID id последнего элемента из списка.
func LastID[T any](items []T, get func(T) string) *string {
	if len(items) == 0 {
		return nil
	}
	id := get(items[len(items)-1])
	return &id
}
