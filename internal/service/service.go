package service

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"
	"unicode/utf8"

	"bookmark-api/internal/model"
	"bookmark-api/internal/repository"
)

var (
	ErrInvalidTitle     = errors.New("title must be between 1 and 255 characters")
	ErrInvalidURL       = errors.New("url must be a valid http or https URL")
	ErrDuplicateURL     = errors.New("url already exists")
	ErrInvalidTags      = errors.New("tags must be at most 255 characters")
	ErrInvalidNote      = errors.New("note must be at most 2000 characters")
	ErrNoFieldsToUpdate = errors.New("at least one field must be provided")
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input model.CreateBookmarkInput) (model.Bookmark, error) {
	bookmark := model.Bookmark{
		Title: strings.TrimSpace(input.Title),
		Tags:  normalizeTags(input.Tags),
		Note:  strings.TrimSpace(input.Note),
	}

	normalizedURL, err := normalizeURL(input.URL)
	if err != nil {
		return model.Bookmark{}, err
	}
	bookmark.URL = normalizedURL

	if err := validateBookmark(bookmark); err != nil {
		return model.Bookmark{}, err
	}

	exists, err := s.repo.ExistsByURL(ctx, bookmark.URL, 0)
	if err != nil {
		return model.Bookmark{}, err
	}
	if exists {
		return model.Bookmark{}, ErrDuplicateURL
	}

	return s.repo.Create(ctx, bookmark)
}

func (s *Service) GetByID(ctx context.Context, id int64) (model.Bookmark, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter model.ListFilter) (model.ListResult, error) {
	filter.Tag = strings.TrimSpace(filter.Tag)
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return model.ListResult{}, err
	}

	return model.ListResult{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *Service) Update(ctx context.Context, id int64, input model.UpdateBookmarkInput) (model.Bookmark, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Bookmark{}, err
	}

	hasField := false
	if input.Title != nil {
		current.Title = strings.TrimSpace(*input.Title)
		hasField = true
	}
	if input.URL != nil {
		normalizedURL, err := normalizeURL(*input.URL)
		if err != nil {
			return model.Bookmark{}, err
		}
		current.URL = normalizedURL
		hasField = true
	}
	if input.Tags != nil {
		current.Tags = normalizeTags(*input.Tags)
		hasField = true
	}
	if input.Note != nil {
		current.Note = strings.TrimSpace(*input.Note)
		hasField = true
	}

	if !hasField {
		return model.Bookmark{}, ErrNoFieldsToUpdate
	}
	if err := validateBookmark(current); err != nil {
		return model.Bookmark{}, err
	}

	exists, err := s.repo.ExistsByURL(ctx, current.URL, current.ID)
	if err != nil {
		return model.Bookmark{}, err
	}
	if exists {
		return model.Bookmark{}, ErrDuplicateURL
	}

	return s.repo.Update(ctx, current)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) IncrementClickCount(ctx context.Context, id int64) (model.Bookmark, error) {
	return s.repo.IncrementClickCount(ctx, id)
}

func (s *Service) Export(ctx context.Context) ([]model.Bookmark, error) {
	return s.repo.Export(ctx)
}

func validateBookmark(bookmark model.Bookmark) error {
	titleLength := utf8.RuneCountInString(bookmark.Title)
	if titleLength < 1 || titleLength > 255 {
		return ErrInvalidTitle
	}
	if utf8.RuneCountInString(bookmark.Tags) > 255 {
		return ErrInvalidTags
	}
	if utf8.RuneCountInString(bookmark.Note) > 2000 {
		return ErrInvalidNote
	}
	if bookmark.URL == "" {
		return ErrInvalidURL
	}
	return nil
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}
	if parsed.Hostname() == "" {
		return "", ErrInvalidURL
	}
	if parsed.User != nil {
		return "", ErrInvalidURL
	}

	parsed.Scheme = scheme
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}
	parsed.Fragment = ""

	return parsed.String(), nil
}

func normalizeTags(raw string) string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return strings.Join(normalized, ",")
}
