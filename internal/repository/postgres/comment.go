package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, c *domain.Comment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO comments (id, post_id, author_id, body, status)
		VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.PostID, c.AuthorID, c.Body, c.Status)
	return err
}

func (r *CommentRepository) List(ctx context.Context, postID string, page, limit int) ([]*domain.Comment, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM comments
		WHERE post_id = $1 AND status = 'approved' AND deleted_at IS NULL`, postID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("comment count: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, `
		SELECT id, post_id, author_id, body, status, created_at
		FROM comments
		WHERE post_id = $1 AND status = 'approved' AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		postID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("comment list: %w", err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Body, &c.Status, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("comment scan: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, total, nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	c := &domain.Comment{}
	err := r.db.QueryRow(ctx, `
		SELECT id, post_id, author_id, body, status, created_at
		FROM comments WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Body, &c.Status, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("comment get: %w", err)
	}
	return c, nil
}

func (r *CommentRepository) UpdateStatus(ctx context.Context, id string, status domain.CommentStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE comments SET status=$1 WHERE id=$2 AND deleted_at IS NULL`, status, id)
	if err != nil {
		return fmt.Errorf("comment update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CommentRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE comments SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("comment soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CommentRepository) SoftDeleteByPostID(ctx context.Context, postID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE comments SET deleted_at=NOW() WHERE post_id=$1 AND deleted_at IS NULL`, postID)
	return err
}
