package service_test

import (
	"context"
	"errors"
	"testing"

	"bookmark-api/internal/model"
)

func TestListRespectsCanceledContext(t *testing.T) {
	svc := newTestService(t)

	ctx, cancel := context.WithCancel(context.Background())
	if _, err := svc.Create(ctx, model.CreateBookmarkInput{
		Title: "context test",
		URL:   "https://example.com/context-test",
	}); err != nil {
		t.Fatalf("create bookmark: %v", err)
	}
	cancel()

	_, err := svc.List(ctx, model.ListFilter{Page: 1, PageSize: 20})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
