package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/config"
	"bookmark-api/internal/handler"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/router"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	repo, err := repository.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open repository: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(context.Background(), migrations.FS); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc)
	engine := router.New(h)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("bookmark api listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
