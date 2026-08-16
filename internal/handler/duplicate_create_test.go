package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/handler"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/router"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func duplicateCreateRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo, err := repository.Open(":memory:")
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.Migrate(context.Background(), migrations.FS); err != nil {
		t.Fatalf("migrate repository: %v", err)
	}
	return router.New(handler.NewHandler(service.NewService(repo)))
}

func TestDuplicateCreateReturnsConflict(t *testing.T) {
	r := duplicateCreateRouter(t)

	body := map[string]string{
		"title": "Go documentation",
		"url":   "https://go.dev/doc/",
		"tags":  "go, docs",
	}

	first := performRequest(r, http.MethodPost, "/api/v1/bookmarks", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first create status 201, got %d; body=%s", first.Code, first.Body.String())
	}

	second := performRequest(r, http.MethodPost, "/api/v1/bookmarks", body)
	if second.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status 409, got %d; body=%s", second.Code, second.Body.String())
	}
}
