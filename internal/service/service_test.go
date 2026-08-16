package service_test

import (
	"context"
	"errors"
	"testing"

	"bookmark-api/internal/model"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/service"
	"bookmark-api/migrations"
)

func newTestService(t *testing.T) *service.Service {
	t.Helper()

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

	return service.NewService(repo)
}

func TestCreateAndGet(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	bookmark, err := svc.Create(ctx, model.CreateBookmarkInput{
		Title: "  Go documentation  ",
		URL:   "https://go.dev/doc/",
		Tags:  " go, docs, go ",
		Note:  "official docs",
	})
	if err != nil {
		t.Fatalf("create bookmark: %v", err)
	}

	if bookmark.ID < 1 {
		t.Fatalf("expected positive id, got %d", bookmark.ID)
	}
	if bookmark.Title != "Go documentation" {
		t.Fatalf("expected trimmed title, got %q", bookmark.Title)
	}
	if bookmark.URL != "https://go.dev/doc" {
		t.Fatalf("expected normalized url, got %q", bookmark.URL)
	}
	if bookmark.Tags != "docs,go" {
		t.Fatalf("expected normalized tags, got %q", bookmark.Tags)
	}

	got, err := svc.GetByID(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("get bookmark: %v", err)
	}
	if got.ID != bookmark.ID {
		t.Fatalf("expected id %d, got %d", bookmark.ID, got.ID)
	}
}

func TestCreateRejectsDuplicateURL(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	input := model.CreateBookmarkInput{
		Title: "Example",
		URL:   "https://example.com/docs/",
		Tags:  "example",
	}
	if _, err := svc.Create(ctx, input); err != nil {
		t.Fatalf("create first bookmark: %v", err)
	}

	_, err := svc.Create(ctx, model.CreateBookmarkInput{
		Title: "Example again",
		URL:   "https://example.com/docs",
		Tags:  "duplicate",
	})
	if !errors.Is(err, service.ErrDuplicateURL) {
		t.Fatalf("expected ErrDuplicateURL, got %v", err)
	}
}

func TestCreateRejectsInvalidURL(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	tests := []string{
		"",
		"not a url",
		"ftp://example.com/file",
		"http://",
		"http://user:pass@example.com",
	}

	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			_, err := svc.Create(ctx, model.CreateBookmarkInput{
				Title: "Invalid",
				URL:   rawURL,
			})
			if !errors.Is(err, service.ErrInvalidURL) {
				t.Fatalf("expected ErrInvalidURL for %q, got %v", rawURL, err)
			}
		})
	}
}

func TestListFiltersAndPagination(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	items := []model.CreateBookmarkInput{
		{Title: "Go Blog", URL: "https://go.dev/blog", Tags: "go, news", Note: "golang article"},
		{Title: "Python Guide", URL: "https://python.org", Tags: "python, news", Note: "python article"},
		{Title: "SQLite Docs", URL: "https://sqlite.org", Tags: "database", Note: "storage docs"},
	}
	for _, input := range items {
		if _, err := svc.Create(ctx, input); err != nil {
			t.Fatalf("create bookmark %q: %v", input.URL, err)
		}
	}

	byTag, err := svc.List(ctx, model.ListFilter{Tag: "go"})
	if err != nil {
		t.Fatalf("list by tag: %v", err)
	}
	if byTag.Total != 1 || len(byTag.Items) != 1 || byTag.Items[0].Title != "Go Blog" {
		t.Fatalf("unexpected tag filter result: total=%d items=%d", byTag.Total, len(byTag.Items))
	}

	byKeyword, err := svc.List(ctx, model.ListFilter{Keyword: "sqlite"})
	if err != nil {
		t.Fatalf("list by keyword: %v", err)
	}
	if byKeyword.Total != 1 || len(byKeyword.Items) != 1 || byKeyword.Items[0].Title != "SQLite Docs" {
		t.Fatalf("unexpected keyword filter result: total=%d items=%d", byKeyword.Total, len(byKeyword.Items))
	}

	pageOne, err := svc.List(ctx, model.ListFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if pageOne.Total != 3 || len(pageOne.Items) != 2 || pageOne.Page != 1 || pageOne.PageSize != 2 {
		t.Fatalf("unexpected first page: total=%d items=%d page=%d page_size=%d", pageOne.Total, len(pageOne.Items), pageOne.Page, pageOne.PageSize)
	}

	pageTwo, err := svc.List(ctx, model.ListFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(pageTwo.Items) != 1 {
		t.Fatalf("expected one item on second page, got %d", len(pageTwo.Items))
	}
}

func TestUpdateClickAndDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, model.CreateBookmarkInput{
		Title: "Old title",
		URL:   "https://example.com",
		Tags:  "old",
		Note:  "old note",
	})
	if err != nil {
		t.Fatalf("create bookmark: %v", err)
	}

	note := "updated note"
	tags := "go, docs"
	updated, err := svc.Update(ctx, created.ID, model.UpdateBookmarkInput{
		Note: &note,
		Tags: &tags,
	})
	if err != nil {
		t.Fatalf("update bookmark: %v", err)
	}
	if updated.Note != note || updated.Tags != "docs,go" {
		t.Fatalf("unexpected update result: %+v", updated)
	}

	for i := 1; i <= 2; i++ {
		clicked, err := svc.IncrementClickCount(ctx, created.ID)
		if err != nil {
			t.Fatalf("increment click count %d: %v", i, err)
		}
		if clicked.ClickCount != int64(i) {
			t.Fatalf("expected click count %d, got %d", i, clicked.ClickCount)
		}
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete bookmark: %v", err)
	}
	if _, err := svc.GetByID(ctx, created.ID); !errors.Is(err, repository.ErrBookmarkNotFound) {
		t.Fatalf("expected ErrBookmarkNotFound after delete, got %v", err)
	}
}
