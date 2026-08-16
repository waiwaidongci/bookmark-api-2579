package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"bookmark-api/internal/model"
)

var ErrBookmarkNotFound = errors.New("bookmark not found")

type Repository struct {
	db *sql.DB
}

func Open(dbPath string) (*Repository, error) {
	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) DB() *sql.DB {
	return r.db
}

func (r *Repository) queryContext(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}

func (r *Repository) Migrate(ctx context.Context, migrationFS fs.FS) error {
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		data, err := fs.ReadFile(migrationFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := r.db.ExecContext(r.queryContext(ctx), string(data)); err != nil {
			return fmt.Errorf("execute migration %s: %w", name, err)
		}
	}

	return nil
}

func (r *Repository) Create(ctx context.Context, bookmark model.Bookmark) (model.Bookmark, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	bookmark.CreatedAt = now
	bookmark.UpdatedAt = now

	result, err := r.db.ExecContext(
		r.queryContext(ctx),
		`INSERT INTO bookmarks (title, url, tags, note, click_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bookmark.Title,
		bookmark.URL,
		bookmark.Tags,
		bookmark.Note,
		bookmark.ClickCount,
		bookmark.CreatedAt,
		bookmark.UpdatedAt,
	)
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("insert bookmark: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("read inserted bookmark id: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (model.Bookmark, error) {
	var bookmark model.Bookmark
	err := r.db.QueryRowContext(
		r.queryContext(ctx),
		`SELECT id, title, url, tags, note, click_count, created_at, updated_at
		 FROM bookmarks WHERE id = ?`,
		id,
	).Scan(
		&bookmark.ID,
		&bookmark.Title,
		&bookmark.URL,
		&bookmark.Tags,
		&bookmark.Note,
		&bookmark.ClickCount,
		&bookmark.CreatedAt,
		&bookmark.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Bookmark{}, ErrBookmarkNotFound
	}
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("get bookmark: %w", err)
	}
	return bookmark, nil
}

func (r *Repository) ExistsByURL(ctx context.Context, url string, excludeID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		r.queryContext(ctx),
		`SELECT EXISTS(
			SELECT 1 FROM bookmarks WHERE url = ? AND id <> ?
		)`,
		url,
		excludeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check bookmark url: %w", err)
	}
	return exists == 1, nil
}

func (r *Repository) List(ctx context.Context, filter model.ListFilter) ([]model.Bookmark, int64, error) {
	where := []string{"1 = 1"}
	args := make([]any, 0, 6)

	if filter.Tag != "" {
		where = append(where, "instr(',' || tags || ',', ',' || ? || ',') > 0")
		args = append(args, filter.Tag)
	}

	if filter.Keyword != "" {
		pattern := "%" + filter.Keyword + "%"
		where = append(where, "(title LIKE ? OR url LIKE ? OR note LIKE ? OR tags LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern)
	}

	whereSQL := strings.Join(where, " AND ")
	var total int64
	if err := r.db.QueryRowContext(r.queryContext(ctx), "SELECT COUNT(*) FROM bookmarks WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookmarks: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := `SELECT id, title, url, tags, note, click_count, created_at, updated_at
		FROM bookmarks WHERE ` + whereSQL + ` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), filter.PageSize, offset)

	rows, err := r.db.QueryContext(r.queryContext(ctx), query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query bookmarks: %w", err)
	}
	defer rows.Close()

	items := make([]model.Bookmark, 0)
	for rows.Next() {
		var bookmark model.Bookmark
		if err := rows.Scan(
			&bookmark.ID,
			&bookmark.Title,
			&bookmark.URL,
			&bookmark.Tags,
			&bookmark.Note,
			&bookmark.ClickCount,
			&bookmark.CreatedAt,
			&bookmark.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan bookmark: %w", err)
		}
		items = append(items, bookmark)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate bookmarks: %w", err)
	}

	return items, total, nil
}

func (r *Repository) Update(ctx context.Context, bookmark model.Bookmark) (model.Bookmark, error) {
	bookmark.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		r.queryContext(ctx),
		`UPDATE bookmarks
		 SET title = ?, url = ?, tags = ?, note = ?, updated_at = ?
		 WHERE id = ?`,
		bookmark.Title,
		bookmark.URL,
		bookmark.Tags,
		bookmark.Note,
		bookmark.UpdatedAt,
		bookmark.ID,
	)
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("update bookmark: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("read updated bookmark count: %w", err)
	}
	if affected == 0 {
		return model.Bookmark{}, ErrBookmarkNotFound
	}

	return r.GetByID(ctx, bookmark.ID)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(r.queryContext(ctx), "DELETE FROM bookmarks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted bookmark count: %w", err)
	}
	if affected == 0 {
		return ErrBookmarkNotFound
	}
	return nil
}

func (r *Repository) IncrementClickCount(ctx context.Context, id int64) (model.Bookmark, error) {
	result, err := r.db.ExecContext(
		r.queryContext(ctx),
		"UPDATE bookmarks SET click_count = click_count + 1 WHERE id = ?",
		id,
	)
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("increment bookmark click count: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.Bookmark{}, fmt.Errorf("read incremented bookmark count: %w", err)
	}
	if affected == 0 {
		return model.Bookmark{}, ErrBookmarkNotFound
	}

	return r.GetByID(ctx, id)
}
