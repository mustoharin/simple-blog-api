package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type ImageRepository struct {
	db *pgxpool.Pool
}

func NewImageRepository(db *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, img *domain.Image) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO images (id, filename, s3_key, url, uploaded_by)
		VALUES ($1, $2, $3, $4, $5)`,
		img.ID, img.Filename, img.S3Key, img.URL, img.UploadedBy)
	if err != nil {
		return fmt.Errorf("image create: %w", err)
	}
	return nil
}

func (r *ImageRepository) GetByID(ctx context.Context, id string) (*domain.Image, error) {
	img := &domain.Image{}
	err := r.db.QueryRow(ctx, `
		SELECT id, filename, s3_key, url, uploaded_by, created_at
		FROM images WHERE id = $1`, id,
	).Scan(&img.ID, &img.Filename, &img.S3Key, &img.URL, &img.UploadedBy, &img.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("image get: %w", err)
	}
	return img, nil
}

func (r *ImageRepository) HardDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM images WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("image delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ImageRepository) List(ctx context.Context, page, limit int) ([]*domain.Image, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM images`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("image count: %w", err)
	}
	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, `
		SELECT id, filename, s3_key, url, uploaded_by, created_at
		FROM images ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("image list: %w", err)
	}
	defer rows.Close()
	var imgs []*domain.Image
	for rows.Next() {
		img := &domain.Image{}
		if err := rows.Scan(&img.ID, &img.Filename, &img.S3Key, &img.URL, &img.UploadedBy, &img.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("image scan: %w", err)
		}
		imgs = append(imgs, img)
	}
	return imgs, total, nil
}
