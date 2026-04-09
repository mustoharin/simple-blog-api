package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type PostTagRepository struct {
	db *pgxpool.Pool
}

func NewPostTagRepository(db *pgxpool.Pool) *PostTagRepository {
	return &PostTagRepository{db: db}
}

func (r *PostTagRepository) SetPostTags(ctx context.Context, postID string, tagIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM post_tags WHERE post_id = $1`, postID); err != nil {
		return fmt.Errorf("clear post tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			postID, tagID); err != nil {
			return fmt.Errorf("insert post tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostTagRepository) GetTagsForPost(ctx context.Context, postID string) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.name, t.slug
		FROM tags t JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
		ORDER BY t.name`, postID)
	if err != nil {
		return nil, fmt.Errorf("get tags for post: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, fmt.Errorf("scan post tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get tags for post scan: %w", err)
	}
	return tags, nil
}
