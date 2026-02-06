package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RoGogDBD/GQLGo/internal/config"
	"github.com/RoGogDBD/GQLGo/internal/handler"
	"github.com/RoGogDBD/GQLGo/internal/logger"
	"github.com/RoGogDBD/GQLGo/internal/qraphql/graph"
	"github.com/RoGogDBD/GQLGo/internal/repository"
	"github.com/RoGogDBD/GQLGo/internal/service"
	"github.com/RoGogDBD/GQLGo/internal/storage"
)

const (
	msgNoDSNConfig  = "config error: отсутствует DSN"
	msgNoAddrConfig = "config error: отсутствует ADDR"
)

func main() {
	// ===================== Логгер =====================
	logger, cleanup, err := config.NewLogger()
	if err != nil {
		_, _ = os.Stdout.WriteString("ошибка при инициализации логера\n")
		os.Exit(1)
	}
	defer cleanup()

	// ===================== Запуск сервера =====================
	if err := run(logger); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error: %v", err)
		}
		os.Exit(1)
	}
}

func run(logger logger.Logger) error {
	// ===================== Кофигурация =====================
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrNoDSN) {
			logger.Errorf(msgNoDSNConfig)
		}
		if errors.Is(err, config.ErrNoAddress) {
			logger.Errorf(msgNoAddrConfig)
		}
		return err
	}

	// ===================== Хранилище =====================
	var (
		userRepo    repository.UserRepo
		postRepo    repository.PostRepo
		commentRepo repository.CommentRepo
		cleanup     func() error
	)

	switch cfg.UsePostgres {
	case false:
		st := repository.NewMemoryStorage()
		userRepo = repository.NewMemoryUserRepo(st)
		postRepo = repository.NewMemoryPostRepo(st)
		commentRepo = repository.NewMemoryCommentRepo(st)
		cleanup = func() error { return nil }
	default:
		st, err := storage.NewDataStorage(cfg.DB.DSN)
		if err != nil {
			return err
		}
		cleanup = st.Close
		userRepo, err = repository.NewPostgresUserRepo(st.DB())
		if err != nil {
			return err
		}
		postRepo, err = repository.NewPostgresPostRepo(st.DB())
		if err != nil {
			return err
		}
		commentRepo, err = repository.NewPostgresCommentRepo(st.DB())
		if err != nil {
			return err
		}
	}
	defer cleanup()

	postService := service.NewPostService(postRepo)
	resolver := &graph.Resolver{
		UserRepo:        userRepo,
		PostRepo:        postRepo,
		CommentRepo:     commentRepo,
		CommentNotifier: service.NewCommentNotifier(logger),
		Logger:          logger,
		PostService:     postService,
	}

	router := handler.NewRouter(resolver)
	logger.Infof("connect to %s for GraphQL playground", cfg.Server.Addr)

	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Infof("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
