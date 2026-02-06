package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/RoGogDBD/GQLGo/internal/logging"
	"github.com/RoGogDBD/GQLGo/internal/models"
	"github.com/RoGogDBD/GQLGo/internal/utils/repository"
	"github.com/google/uuid"
)

var (
	ErrEmptyID      = errors.New("неверный id")
	ErrNotFound     = errors.New("не найдено")
	ErrAlreadyExist = errors.New("уже существует")
	ErrNilEntity    = errors.New("пустая сущность")
	ErrBadCursor    = errors.New("неверный курсор")
)

type MemoryStorage struct {
	// mu sync.Mutex
	mu             sync.RWMutex
	users          map[string]*models.User
	userCreated    map[string]time.Time
	posts          map[string]*models.Post
	postCreated    map[string]time.Time
	postOrder      []string
	comments       map[string]*models.Comment
	commentCreated map[string]time.Time
	byPost         map[string][]string
	byParent       map[string][]string

	ttl           time.Duration
	lastPrune     time.Time
	pruneInterval time.Duration
	logger        logging.Logger
}

type (
	MemoryUserRepo    struct{ st *MemoryStorage }
	MemoryPostRepo    struct{ st *MemoryStorage }
	MemoryCommentRepo struct{ st *MemoryStorage }
)

// ==================== Конструктор ====================
func NewMemoryStorage(logger logging.Logger) *MemoryStorage {
	return NewMemoryStorageWithTTL(24*time.Hour, logger)
}

func NewMemoryStorageWithTTL(ttl time.Duration, logger logging.Logger) *MemoryStorage {
	return &MemoryStorage{
		users:          map[string]*models.User{},
		userCreated:    map[string]time.Time{},
		posts:          map[string]*models.Post{},
		postCreated:    map[string]time.Time{},
		postOrder:      make([]string, 0),
		comments:       map[string]*models.Comment{},
		commentCreated: map[string]time.Time{},
		byPost:         map[string][]string{},
		byParent:       map[string][]string{},
		ttl:            ttl,
		pruneInterval:  time.Minute,
		logger:         logger,
	}
}

// ==================== Конструктор репозиториев ====================
func NewMemoryPostRepo(st *MemoryStorage, logger logging.Logger) *MemoryPostRepo {
	st.logger = logger
	return &MemoryPostRepo{st: st}
}
func NewMemoryUserRepo(st *MemoryStorage, logger logging.Logger) *MemoryUserRepo {
	st.logger = logger
	return &MemoryUserRepo{st: st}
}
func NewMemoryCommentRepo(st *MemoryStorage, logger logging.Logger) *MemoryCommentRepo {
	st.logger = logger
	return &MemoryCommentRepo{st: st}
}

func (st *MemoryStorage) maybePrune(now time.Time) {
	if st.ttl <= 0 {
		return
	}
	if !st.lastPrune.IsZero() && now.Sub(st.lastPrune) < st.pruneInterval {
		return
	}
	st.lastPrune = now
	st.pruneExpired(now)
}

func (st *MemoryStorage) logErr(msg string, err error) {
	if st.logger != nil && err != nil {
		st.logger.Errorf("%s: %v", msg, err)
	}
}

func removeID(ids []string, id string) []string {
	for i := 0; i < len(ids); i++ {
		if ids[i] == id {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}

func (st *MemoryStorage) pruneExpired(now time.Time) {
	// posts
	if st.ttl > 0 {
		for id, ts := range st.postCreated {
			if now.Sub(ts) > st.ttl {
				delete(st.posts, id)
				delete(st.postCreated, id)
				st.postOrder = removeID(st.postOrder, id)
				// remove comments for this post
				for _, cid := range st.byPost[id] {
					st.deleteCommentLocked(cid)
				}
				delete(st.byPost, id)
			}
		}
		// users
		for id, ts := range st.userCreated {
			if now.Sub(ts) > st.ttl {
				delete(st.users, id)
				delete(st.userCreated, id)
			}
		}
		// comments
		for id, ts := range st.commentCreated {
			if now.Sub(ts) > st.ttl {
				st.deleteCommentLocked(id)
			}
		}
	}
}

func (st *MemoryStorage) deleteCommentLocked(id string) {
	c := st.comments[id]
	if c == nil {
		return
	}
	delete(st.comments, id)
	delete(st.commentCreated, id)
	if c.ParentID != nil {
		parentKey := *c.ParentID
		st.byParent[parentKey] = removeID(st.byParent[parentKey], id)
		if parent := st.comments[parentKey]; parent != nil && parent.ChildrenCount > 0 {
			parent.ChildrenCount--
		}
	} else {
		st.byParent[""] = removeID(st.byParent[""], id)
	}
	st.byPost[c.PostID] = removeID(st.byPost[c.PostID], id)
}

func paginateIDs(ids []string, after *string, first int32) []string {
	if first <= 0 {
		first = DefaultPageSize
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

// ======================== POST REPO ========================
func (r *MemoryPostRepo) GetByID(ctx context.Context, id string) (*models.Post, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == "" {
		r.st.logErr("получение поста", ErrEmptyID)
		return nil, ErrEmptyID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	p := r.st.posts[id]
	if p == nil {
		return nil, nil
	}

	return repository.ClonePost(p), nil
}

func (r *MemoryPostRepo) Create(ctx context.Context, in models.CreatePostInput) (*models.Post, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	commentsEnabled := true
	if in.CommentsEnabled != nil {
		commentsEnabled = *in.CommentsEnabled
	}

	p := &models.Post{
		ID:              uuid.NewString(),
		Title:           in.Title,
		Body:            in.Body,
		CommentsEnabled: commentsEnabled,
		Author:          &models.User{ID: in.AuthorID},
	}
	if p == nil {
		return nil, ErrNilEntity
	}
	if p.ID == "" {
		r.st.logErr("создание поста", ErrEmptyID)
		return nil, ErrEmptyID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	defer r.st.mu.Unlock()
	r.st.maybePrune(now)

	if r.st.posts[p.ID] != nil {
		r.st.logErr("создание поста", ErrAlreadyExist)
		return nil, ErrAlreadyExist
	}

	cp := repository.ClonePost(p)
	r.st.posts[cp.ID] = cp
	r.st.postCreated[cp.ID] = now
	r.st.postOrder = append(r.st.postOrder, cp.ID)

	return repository.ClonePost(cp), nil
}

func (r *MemoryPostRepo) List(ctx context.Context, first int32, after *string) ([]*models.Post, *string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	ids := paginateIDs(r.st.postOrder, after, first)
	posts := make([]*models.Post, 0, len(ids))
	for _, id := range ids {
		if p := r.st.posts[id]; p != nil {
			posts = append(posts, repository.ClonePost(p))
		}
	}
	return posts, repository.LastID(posts, func(p *models.Post) string { return p.ID }), nil
}

func (r *MemoryPostRepo) SetCommentsEnabled(ctx context.Context, postID string, enabled bool) (*models.Post, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if postID == "" {
		r.st.logErr("обновление комментариев", ErrEmptyID)
		return nil, ErrEmptyID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	defer r.st.mu.Unlock()
	r.st.maybePrune(now)

	p := r.st.posts[postID]
	if p == nil {
		return nil, nil
	}
	p.CommentsEnabled = enabled
	return repository.ClonePost(p), nil
}

// ======================== USER REPO ========================
func (r *MemoryUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == "" {
		r.st.logErr("получение пользователя", ErrEmptyID)
		return nil, ErrEmptyID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	u := r.st.users[id]
	if u == nil {
		return nil, nil
	}
	return repository.CloneUser(u), nil
}

func (r *MemoryUserRepo) List(ctx context.Context, first int32, after *string) ([]*models.User, *string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	ids := paginateIDs(repository.SortedKeys(r.st.users), after, first)
	users := make([]*models.User, 0, len(ids))
	for _, id := range ids {
		if u := r.st.users[id]; u != nil {
			users = append(users, repository.CloneUser(u))
		}
	}
	return users, repository.LastID(users, func(u *models.User) string { return u.ID }), nil
}

// ======================== COMMENT REPO ========================
func (r *MemoryCommentRepo) GetMeta(ctx context.Context, id string) (string, int, error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	if id == "" {
		r.st.logErr("получение комментария", ErrEmptyID)
		return "", 0, ErrEmptyID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	c := r.st.comments[id]
	if c == nil {
		return "", 0, sql.ErrNoRows
	}
	return c.PostID, int(c.Depth), nil
}

func (r *MemoryCommentRepo) Create(ctx context.Context, postID, authorID string, parentID *string, body string, depth int) (*models.Comment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if postID == "" || authorID == "" {
		r.st.logErr("создание комментария", ErrEmptyID)
		return nil, ErrEmptyID
	}

	timeNow := time.Now().UTC()
	id := uuid.NewString()

	comment := &models.Comment{
		ID:            id,
		PostID:        postID,
		Author:        &models.User{ID: authorID},
		Body:          body,
		ParentID:      parentID,
		Depth:         int32(depth),
		ChildrenCount: 0,
		Children: &models.CommentConnection{
			Edges:      []*models.CommentEdge{},
			PageInfo:   &models.PageInfo{HasNextPage: false, EndCursor: nil},
			TotalCount: 0,
		},
		CreatedAt: timeNow,
	}

	r.st.mu.Lock()
	defer r.st.mu.Unlock()
	r.st.maybePrune(timeNow)

	r.st.comments[id] = comment
	r.st.commentCreated[id] = timeNow
	r.st.byPost[postID] = append(r.st.byPost[postID], id)
	parentKey := ""
	if parentID != nil {
		parentKey = *parentID
	}
	r.st.byParent[parentKey] = append(r.st.byParent[parentKey], id)

	if parentID != nil && *parentID != "" {
		if parent := r.st.comments[*parentID]; parent != nil {
			parent.ChildrenCount++
		}
	}

	return comment, nil
}

func (r *MemoryCommentRepo) ListByParent(ctx context.Context, postID string, parentID *string, first int32, after *string, order models.CommentOrder) ([]*models.Comment, *string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if postID == "" {
		r.st.logErr("список комментариев", ErrEmptyID)
		return nil, nil, ErrEmptyID
	}
	if first <= 0 {
		first = DefaultPageSize
	}
	if !order.IsValid() {
		order = models.CommentOrderNewest
	}

	parentKey := ""
	if parentID != nil {
		parentKey = *parentID
	}

	now := time.Now().UTC()
	r.st.mu.Lock()
	r.st.maybePrune(now)
	r.st.mu.Unlock()

	r.st.mu.RLock()
	defer r.st.mu.RUnlock()

	ids := append([]string(nil), r.st.byParent[parentKey]...)
	if len(ids) == 0 {
		return []*models.Comment{}, nil, nil
	}

	filtered := make([]string, 0, len(ids))
	for _, id := range ids {
		if c := r.st.comments[id]; c != nil && c.PostID == postID {
			filtered = append(filtered, id)
		}
	}

	repository.SortCommentIDs(filtered, r.st.comments, order)
	//func sortCommentIDs(ids []string, comments map[string]*models.Comment, order models.CommentOrder) {
	//	sort.Slice(ids, func(i, j int) bool {
	//		return CommentLess(comments[ids[i]], comments[ids[j]], order)
	//	})
	//}

	filtered = paginateIDs(filtered, after, first)

	out := make([]*models.Comment, 0, len(filtered))
	for _, id := range filtered {
		if c := r.st.comments[id]; c != nil {
			out = append(out, c)
		}
	}

	return out, repository.LastID(out, func(c *models.Comment) string { return c.ID }), nil
}
