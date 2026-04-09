package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tags (id, name, slug) VALUES ($1, $2, $3)`,
		tag.ID, tag.Name, tag.Slug)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return domain.ErrTagNameAlreadyExists
		}
		return fmt.Errorf("tag create: %w", err)
	}
	return nil
}

func (r *TagRepository) List(ctx context.Context) ([]*domain.Tag, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, slug FROM tags ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("tag list: %w", err)
	}
	defer rows.Close()
	var tags []*domain.Tag
	for rows.Next() {
		t := &domain.Tag{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, fmt.Errorf("tag scan: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tag list scan: %w", err)
	}
	return tags, nil
}

func (r *TagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	t := &domain.Tag{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug FROM tags WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("tag get: %w", err)
	}
	return t, nil
}

func (r *TagRepository) HardDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("tag delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TagRepository) GetOrCreateByName(ctx context.Context, name string) (*domain.Tag, error) {
	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
	t := &domain.Tag{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug FROM tags WHERE name = $1`, name,
	).Scan(&t.ID, &t.Name, &t.Slug)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("tag get-or-create lookup: %w", err)
	}

	t.ID = uuid.NewString()
	t.Name = name
	t.Slug = slug
	if err := r.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
