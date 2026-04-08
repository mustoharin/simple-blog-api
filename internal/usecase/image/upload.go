package image

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// S3Uploader abstracts the S3 upload/delete operations.
type S3Uploader interface {
	Upload(ctx context.Context, key, contentType string, data []byte) (string, error)
	Delete(ctx context.Context, key string) error
}

type UploadInput struct {
	Filename      string
	ContentType   string
	Data          []byte
	UploaderID    string
	UploaderEmail string
}

type UploadImageUsecase struct {
	images   domain.ImageRepository
	uploader S3Uploader
	audit    domain.AuditLogger
}

func NewUploadImageUsecase(images domain.ImageRepository, uploader S3Uploader, audit domain.AuditLogger) *UploadImageUsecase {
	return &UploadImageUsecase{images: images, uploader: uploader, audit: audit}
}

func (uc *UploadImageUsecase) Execute(ctx context.Context, in UploadInput) (*domain.Image, error) {
	if !allowedMimeTypes[in.ContentType] {
		return nil, domain.ErrInvalidMimeType
	}
	if len(in.Data) > maxFileSize {
		return nil, domain.ErrFileTooLarge
	}

	id := uuid.NewString()
	key := fmt.Sprintf("images/%s-%s", id, in.Filename)

	url, err := uc.uploader.Upload(ctx, key, in.ContentType, in.Data)
	if err != nil {
		return nil, err
	}

	img := &domain.Image{
		ID:         id,
		Filename:   in.Filename,
		S3Key:      key,
		URL:        url,
		UploadedBy: in.UploaderID,
		CreatedAt:  time.Now(),
	}

	if err := uc.images.Create(ctx, img); err != nil {
		_ = uc.uploader.Delete(ctx, key) // rollback upload on DB failure
		return nil, err
	}

	actorID := in.UploaderID
	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   in.UploaderEmail,
		Action:       domain.AuditImageUploaded,
		ResourceType: "image",
		ResourceID:   &img.ID,
	})

	return img, nil
}
