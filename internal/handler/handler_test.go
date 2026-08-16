package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/handler"
	"bookmark-api/internal/model"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/router"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func newTestRouter(t *testing.T) *gin.Engine {
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

	h := handler.NewHandler(service.NewService(repo))
	return router.New(h)
}

func performRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	return response.Data
}

func TestBookmarkHTTPFlow(t *testing.T) {
	r := newTestRouter(t)

	createBody := map[string]string{
		"title": "Go documentation",
		"url":   "https://go.dev/doc/",
		"tags":  "go, docs",
		"note":  "official docs",
	}

	rec := performRequest(r, http.MethodPost, "/api/v1/bookmarks", createBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body=%s", rec.Code, rec.Body.String())
	}
	created := decodeData[model.Bookmark](t, rec)
	if created.ID < 1 || created.URL != "https://go.dev/doc" {
		t.Fatalf("unexpected created bookmark: %+v", created)
	}

	rec = performRequest(r, http.MethodPost, "/api/v1/bookmarks", createBody)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate status 409, got %d; body=%s", rec.Code, rec.Body.String())
	}

	rec = performRequest(r, http.MethodGet, "/api/v1/bookmarks?tag=go&keyword=Go", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
	list := decodeData[model.ListResult](t, rec)
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("unexpected filtered list: total=%d items=%d", list.Total, len(list.Items))
	}

	rec = performRequest(r, http.MethodPost, "/api/v1/bookmarks/1/click", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected click status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
	clicked := decodeData[model.Bookmark](t, rec)
	if clicked.ClickCount != 1 {
		t.Fatalf("expected click_count 1, got %d", clicked.ClickCount)
	}

	rec = performRequest(r, http.MethodPut, "/api/v1/bookmarks/1", map[string]string{
		"note": "updated note",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
	updated := decodeData[model.Bookmark](t, rec)
	if updated.Note != "updated note" {
		t.Fatalf("unexpected updated note: %q", updated.Note)
	}

	rec = performRequest(r, http.MethodDelete, "/api/v1/bookmarks/1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected delete status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
}
