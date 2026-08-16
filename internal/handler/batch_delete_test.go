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

func batchDeleteRouter(t *testing.T) *gin.Engine {
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

func TestBatchDeleteRejectsMissingIDWithoutPartialDelete(t *testing.T) {
	r := batchDeleteRouter(t)

	first := performRequest(r, http.MethodPost, "/api/v1/bookmarks", map[string]string{
		"title": "first",
		"url":   "https://example.com/first",
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("expected create first status 201, got %d; body=%s", first.Code, first.Body.String())
	}

	second := performRequest(r, http.MethodPost, "/api/v1/bookmarks", map[string]string{
		"title": "second",
		"url":   "https://example.com/second",
	})
	if second.Code != http.StatusCreated {
		t.Fatalf("expected create second status 201, got %d; body=%s", second.Code, second.Body.String())
	}

	deleteResp := performRequest(r, http.MethodPost, "/api/v1/bookmarks/batch-delete", map[string]any{
		"ids": []int64{1, 999},
	})
	if deleteResp.Code != http.StatusNotFound {
		t.Fatalf("expected missing-id batch delete status 404, got %d; body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	check := performRequest(r, http.MethodGet, "/api/v1/bookmarks/1", nil)
	if check.Code != http.StatusOK {
		t.Fatalf("expected first bookmark to remain after failed batch delete, got status %d; body=%s", check.Code, check.Body.String())
	}
}
