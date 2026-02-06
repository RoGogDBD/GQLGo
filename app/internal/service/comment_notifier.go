package service

import (
	"fmt"
	"sync"

	"github.com/RoGogDBD/GQLGo/internal/models"
)

type Logger interface {
	Printf(format string, args ...any)
}

type CommentNotifier struct {
	mu       sync.RWMutex
	byPostID map[string][]commentSubscriber
	logger   Logger
}

type commentSubscriber struct {
	stream chan *models.Comment
	done   chan struct{}
}

func NewCommentNotifier(logger Logger) *CommentNotifier {
	return &CommentNotifier{
		byPostID: make(map[string][]commentSubscriber),
		logger:   logger,
	}
}

func (n *CommentNotifier) Subscribe(postID string) (chan *models.Comment, func()) {
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

	return sub.stream, unSub
}

func (n *CommentNotifier) Publish(postID string, c *models.Comment) {
	if postID == "" {
		fmt.Errorf("пустой postID в CN")
		return
	}
	if c == nil {
		fmt.Errorf("пустой комментарий в CN")
		return
	}
	if c.PostID != "" && c.PostID != postID {
		fmt.Errorf("несовпадение postID в CN: %s и %s", c.PostID, postID)
		return
	}

	n.mu.RLock()
	subscribers := append([]commentSubscriber(nil), n.byPostID[postID]...)
	n.mu.RUnlock()

	for _, sub := range subscribers {
		if !n.trySend(sub, c) {
			fmt.Errorf("не удалось отправить коммент")
		}
	}
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
