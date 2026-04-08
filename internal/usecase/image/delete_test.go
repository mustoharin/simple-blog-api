package image_test

import (
	"context"
	"errors"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteImage_Success(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	img := &domain.Image{ID: "img1", S3Key: "images/foo.jpg"}
	repo.On("GetByID", mock.Anything, "img1").Return(img, nil)
	uploader.On("Delete", mock.Anything, "images/foo.jpg").Return(nil)
	repo.On("HardDelete", mock.Anything, "img1").Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := image.NewDeleteImageUsecase(repo, uploader, audit)
	err := uc.Execute(context.Background(), "img1", "u1", "alice@example.com")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	uploader.AssertExpectations(t)
}

func TestDeleteImage_NotFound(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	repo.On("GetByID", mock.Anything, "missing").Return(nil, domain.ErrNotFound)

	uc := image.NewDeleteImageUsecase(repo, uploader, audit)
	err := uc.Execute(context.Background(), "missing", "u1", "alice@example.com")

	assert.True(t, errors.Is(err, domain.ErrNotFound))
	uploader.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDeleteImage_S3DeleteError(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	img := &domain.Image{ID: "img2", S3Key: "images/bar.jpg"}
	repo.On("GetByID", mock.Anything, "img2").Return(img, nil)
	uploader.On("Delete", mock.Anything, "images/bar.jpg").Return(errors.New("s3 error"))

	uc := image.NewDeleteImageUsecase(repo, uploader, audit)
	err := uc.Execute(context.Background(), "img2", "u1", "alice@example.com")

	assert.Error(t, err)
	repo.AssertNotCalled(t, "HardDelete", mock.Anything, mock.Anything)
}
