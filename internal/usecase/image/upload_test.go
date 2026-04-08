package image_test

import (
	"context"
	"testing"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockImageRepo struct{ mock.Mock }

func (m *mockImageRepo) Create(ctx context.Context, img *domain.Image) error {
	return m.Called(ctx, img).Error(0)
}
func (m *mockImageRepo) GetByID(ctx context.Context, id string) (*domain.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}
func (m *mockImageRepo) HardDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockImageRepo) List(ctx context.Context, page, limit int) ([]*domain.Image, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*domain.Image), args.Int(1), args.Error(2)
}

type mockS3Uploader struct{ mock.Mock }

func (m *mockS3Uploader) Upload(ctx context.Context, key, contentType string, data []byte) (string, error) {
	args := m.Called(ctx, key, contentType, data)
	return args.String(0), args.Error(1)
}
func (m *mockS3Uploader) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}

func TestUploadImage_Success(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	uploader.On("Upload", mock.Anything, mock.Anything, "image/jpeg", mock.Anything).
		Return("https://bucket.s3.amazonaws.com/images/abc.jpg", nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := image.NewUploadImageUsecase(repo, uploader, audit)
	img, err := uc.Execute(context.Background(), image.UploadInput{
		Filename:      "photo.jpg",
		ContentType:   "image/jpeg",
		Data:          []byte("fake-jpeg-data"),
		UploaderID:    "u1",
		UploaderEmail: "alice@example.com",
	})
	assert.NoError(t, err)
	assert.Contains(t, img.URL, "https://")
}

func TestUploadImage_InvalidMimeType(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	uc := image.NewUploadImageUsecase(repo, uploader, audit)
	_, err := uc.Execute(context.Background(), image.UploadInput{
		Filename:    "malware.exe",
		ContentType: "application/octet-stream",
		Data:        []byte("bad data"),
		UploaderID:  "u1",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidMimeType)
}

func TestUploadImage_FileTooLarge(t *testing.T) {
	repo := new(mockImageRepo)
	uploader := new(mockS3Uploader)
	audit := new(mockAuditLogger)

	bigData := make([]byte, 11*1024*1024) // 11 MB
	uc := image.NewUploadImageUsecase(repo, uploader, audit)
	_, err := uc.Execute(context.Background(), image.UploadInput{
		Filename:    "huge.jpg",
		ContentType: "image/jpeg",
		Data:        bigData,
		UploaderID:  "u1",
	})
	assert.ErrorIs(t, err, domain.ErrFileTooLarge)
}
