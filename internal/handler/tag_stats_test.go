package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/handler"
	"bookmark-api/internal/model"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/router"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func tagStatsRouter(t *testing.T) *gin.Engine {
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

func TestTagStatsReturnsNonNilTags(t *testing.T) {
	r := tagStatsRouter(t)

	createResp := performRequest(r, http.MethodPost, "/api/v1/bookmarks", map[string]string{
		"title": "Go documentation",
		"url":   "https://go.dev/doc/",
		"tags":  "go, docs",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d; body=%s", createResp.Code, createResp.Body.String())
	}

	statsResp := performRequest(r, http.MethodGet, "/api/v1/bookmarks/stats/tags", nil)
	if statsResp.Code != http.StatusOK {
		t.Fatalf("expected tag stats status 200, got %d; body=%s", statsResp.Code, statsResp.Body.String())
	}

	stats := decodeData[model.TagStatsResult](t, statsResp)
	if stats.Tags == nil || len(stats.Tags) != 2 || stats.Total != 2 {
		t.Fatalf("unexpected tag stats: tags=%#v total=%d", stats.Tags, stats.Total)
	}
}
