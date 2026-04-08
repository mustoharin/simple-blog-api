package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

const postSelectCols = `
    p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url,
    p.status, p.author_id, p.published_at, p.created_at, p.updated_at,
    p.view_count, p.comment_count`

func scanPost(row pgx.Row) (*domain.Post, error) {
	p := &domain.Post{}
	err := row.Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &p.Excerpt, &p.CoverImageURL,
		&p.Status, &p.AuthorID, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
		&p.ViewCount, &p.CommentCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO posts (id, title, slug, content, excerpt, cover_image_url, status, author_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		post.ID, post.Title, post.Slug, post.Content,
		post.Excerpt, post.CoverImageURL, post.Status, post.AuthorID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") && strings.Contains(err.Error(), "slug") {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("post create: %w", err)
	}
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+postSelectCols+` FROM posts p
		WHERE p.id = $1 AND p.deleted_at IS NULL`, id)
	return scanPost(row)
}

func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+postSelectCols+` FROM posts p
		WHERE p.slug = $1 AND p.deleted_at IS NULL`, slug)
	return scanPost(row)
}

func (r *PostRepository) List(ctx context.Context, filter domain.PostFilter, publicOnly bool) ([]*domain.Post, int, error) {
	where := "WHERE p.deleted_at IS NULL"
	args := []any{}
	idx := 1

	if publicOnly {
		where += " AND p.status = 'published'"
	}
	if filter.Query != "" {
		where += fmt.Sprintf(" AND p.search_vector @@ plainto_tsquery('english', $%d)", idx)
		args = append(args, filter.Query)
		idx++
	}
	if filter.Tag != "" {
		where += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM post_tags pt JOIN tags t ON pt.tag_id = t.id
			WHERE pt.post_id = p.id AND t.slug = $%d)`, idx)
		args = append(args, filter.Tag)
		idx++
	}
	if filter.AuthorID != "" {
		where += fmt.Sprintf(" AND p.author_id = $%d", idx)
		args = append(args, filter.AuthorID)
		idx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM posts p %s", where)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("post count: %w", err)
	}

	orderBy := "p.created_at DESC"
	if filter.Sort == "views" {
		orderBy = "p.view_count DESC"
	} else if filter.Sort == "comments" {
		orderBy = "p.comment_count DESC"
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	listArgs := append(args, limit, (page-1)*limit)
	query := fmt.Sprintf(`SELECT %s FROM posts p %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		postSelectCols, where, orderBy, idx, idx+1)

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("post list: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}
	return posts, total, nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE posts
		SET title=$1, slug=$2, content=$3, excerpt=$4, cover_image_url=$5, updated_at=NOW()
		WHERE id=$6 AND deleted_at IS NULL`,
		post.Title, post.Slug, post.Content, post.Excerpt, post.CoverImageURL, post.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") && strings.Contains(err.Error(), "slug") {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("post update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE posts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("post soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostRepository) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE posts SET view_count = view_count + 1 WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *PostRepository) SetPublished(ctx context.Context, id string, published bool) error {
	var query string
	if published {
		query = `UPDATE posts SET status='published', published_at=$2, updated_at=NOW()
                 WHERE id=$1 AND deleted_at IS NULL`
		_, err := r.db.Exec(ctx, query, id, time.Now())
		return err
	}
	query = `UPDATE posts SET status='draft', published_at=NULL, updated_at=NOW()
             WHERE id=$1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
