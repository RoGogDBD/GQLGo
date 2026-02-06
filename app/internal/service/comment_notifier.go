package service

import (
	"errors"
	"sync"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

var (
	ErrNilComment     = errors.New("пустой комментарий")
	ErrSendFailed     = errors.New("не удалось отправить коммент")
	ErrEmptyPostID    = errors.New("пустой postID")
	ErrPostIDMismatch = errors.New("postID комментария не совпадает с postID публикации")
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type (
	CommentNotifier struct {
		mu       sync.RWMutex
		byPostID map[string][]commentSubscriber
		logger   Logger
	}

	commentSubscriber struct {
		stream chan *models.Comment
		done   chan struct{}
	}
)

func NewCommentNotifier(logger Logger) *CommentNotifier {
	return &CommentNotifier{
		byPostID: make(map[string][]commentSubscriber),
		logger:   logger,
	}
}

func (n *CommentNotifier) Subscribe(postID string) (chan *models.Comment, func(), error) {
	if postID == "" {
		return nil, nil, ErrEmptyPostID
	}
	sub := commentSubscriber{
		stream: make(chan *models.Comment, 1),
		done:   make(chan struct{}),
	}

	n.mu.Lock()
	n.byPostID[postID] = append(n.byPostID[postID], sub)
	n.mu.Unlock()

	unSub := func() {
		n.mu.Lock()
		subscribers := n.byPostID[postID]
		for i, s := range subscribers {
			if s.stream == sub.stream {
				close(s.done)
				close(s.stream)
				subscribers = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
		if len(subscribers) == 0 {
			delete(n.byPostID, postID)
		} else {
			n.byPostID[postID] = subscribers
		}
		n.mu.Unlock()
	}

	return sub.stream, unSub, nil
}

func (n *CommentNotifier) Publish(postID string, c *models.Comment) error {
	var errs []error
	if postID == "" {
		errs = append(errs, ErrEmptyPostID)
	}
	if c == nil {
		errs = append(errs, ErrNilComment)
	}
	if c != nil && c.PostID != "" && c.PostID != postID {
		errs = append(errs, ErrPostIDMismatch)
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	n.mu.RLock()
	subscribers := append([]commentSubscriber(nil), n.byPostID[postID]...)
	n.mu.RUnlock()

	for _, sub := range subscribers {
		if !n.trySend(sub, c) {
			errs = append(errs, ErrSendFailed)
		}
	}
	return errors.Join(errs...)
}

func (n *CommentNotifier) trySend(sub commentSubscriber, c *models.Comment) (sent bool) {
	defer func() {
		if recover() != nil {
			sent = false
		}
	}()

	select {
	case <-sub.done:
		return false
	case sub.stream <- c:
		return true
	default:
		return false
	}
}
