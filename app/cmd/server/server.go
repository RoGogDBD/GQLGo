package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/RoGogDBD/GQLGo/internal/config"
	"github.com/RoGogDBD/GQLGo/internal/handler"
	"github.com/RoGogDBD/GQLGo/internal/qraphql/graph"
	"github.com/RoGogDBD/GQLGo/internal/repository"
	"github.com/RoGogDBD/GQLGo/internal/storage"
)

const (
	msgNoDSNConfig  = "config error: отсутствует DSN"
	msgNoAddrConfig = "config error: отсутствует ADDR"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if err := run(logger); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("error: %v", err)
		}
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	// ===================== Кофигурация =====================
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrNoDSN) {
			logger.Printf(msgNoDSNConfig)
		}
		if errors.Is(err, config.ErrNoAddress) {
			logger.Printf(msgNoAddrConfig)
		}
		return err
	}

	// ===================== Хранилище =====================
	st, err := storage.NewDataStorage(cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer st.Close()

	userRepo, err := repository.NewPostgresUserRepo(st.DB())
	if err != nil {
		return err
	}
	postRepo, err := repository.NewPostgresPostRepo(st.DB())
	if err != nil {
		return err
	}
	resolver := &graph.Resolver{
		UserRepo: userRepo,
		PostRepo: postRepo,
	}

	router := handler.NewRouter(resolver)
	logger.Printf("connect to %s for GraphQL playground", cfg.Server.Addr)
	return router.Run(cfg.Server.Addr)
}
