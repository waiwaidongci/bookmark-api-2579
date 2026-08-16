package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/handler"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/router"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func exportLeakRouter(t *testing.T) *gin.Engine {
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

func TestExportDoesNotLeakConnection(t *testing.T) {
	r := exportLeakRouter(t)

	createResp := performRequest(r, http.MethodPost, "/api/v1/bookmarks", map[string]string{
		"title": "export test",
		"url":   "https://example.com/export-test",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d; body=%s", createResp.Code, createResp.Body.String())
	}

	secondResp := performRequest(r, http.MethodPost, "/api/v1/bookmarks", map[string]string{
		"title": "export test two",
		"url":   "https://example.com/export-test-two",
	})
	if secondResp.Code != http.StatusCreated {
		t.Fatalf("expected second create status 201, got %d; body=%s", secondResp.Code, secondResp.Body.String())
	}

	exportResp := performRequest(r, http.MethodGet, "/api/v1/bookmarks/export?limit=1", nil)
	if exportResp.Code != http.StatusOK {
		t.Fatalf("expected export status 200, got %d; body=%s", exportResp.Code, exportResp.Body.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookmarks?page=1&page_size=20", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected list after export status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
}
