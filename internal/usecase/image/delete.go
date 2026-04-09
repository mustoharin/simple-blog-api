package image

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type DeleteImageUsecase struct {
	images   domain.ImageRepository
	uploader S3Uploader
	audit    domain.AuditLogger
}

func NewDeleteImageUsecase(images domain.ImageRepository, uploader S3Uploader, audit domain.AuditLogger) *DeleteImageUsecase {
	return &DeleteImageUsecase{images: images, uploader: uploader, audit: audit}
}

func (uc *DeleteImageUsecase) Execute(ctx context.Context, imageID, actorID, actorEmail string) error {
	img, err := uc.images.GetByID(ctx, imageID)
	if err != nil {
		return err
	}

	if err := uc.uploader.Delete(ctx, img.S3Key); err != nil {
		return err
	}

	if err := uc.images.HardDelete(ctx, imageID); err != nil {
		return err
	}

	_ = uc.audit.Log(ctx, &domain.AuditLog{
		ID:           uuid.NewString(),
		ActorID:      &actorID,
		ActorEmail:   actorEmail,
		Action:       domain.AuditImageDeleted,
		ResourceType: "image",
		ResourceID:   &imageID,
	})

	return nil
}
