package user_test

import (
	"context"
	"testing"
	"time"

	"simple-blog-api/internal/domain"
	"simple-blog-api/internal/usecase/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResendInvitation_AlreadyActive_Returns409(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	email := new(mockEmailSender)
	audit := new(mockAuditLogger)

	activeUser := &domain.User{ID: "u1", Email: "alice@example.com", Status: domain.UserStatusActive}
	repo.On("GetByID", mock.Anything, "u1").Return(activeUser, nil)

	uc := user.NewResendInvitationUsecase(repo, invRepo, email, audit, "http://localhost:3000")
	err := uc.Execute(context.Background(), "u1", "admin1", "admin@example.com")
	assert.ErrorIs(t, err, domain.ErrUserAlreadyActive)
}

func TestResendInvitation_PendingUser_Success(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	email := new(mockEmailSender)
	audit := new(mockAuditLogger)

	pendingUser := &domain.User{ID: "u1", Email: "bob@example.com", Status: domain.UserStatusPendingInvitation}
	repo.On("GetByID", mock.Anything, "u1").Return(pendingUser, nil)
	invRepo.On("InvalidatePrevious", mock.Anything, "u1").Return(nil)
	invRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	email.On("Send", "bob@example.com", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewResendInvitationUsecase(repo, invRepo, email, audit, "http://localhost:3000")
	err := uc.Execute(context.Background(), "u1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond) // allow fire-and-forget goroutines
	invRepo.AssertExpectations(t)
	email.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestResendInvitation_ExpiredUser_ResetsStatus(t *testing.T) {
	repo := new(mockUserRepo)
	invRepo := new(mockInvitationTokenRepo)
	email := new(mockEmailSender)
	audit := new(mockAuditLogger)

	expiredUser := &domain.User{ID: "u1", Email: "carol@example.com", Status: domain.UserStatusExpiredInvitation}
	repo.On("GetByID", mock.Anything, "u1").Return(expiredUser, nil)
	repo.On("UpdateStatus", mock.Anything, "u1", domain.UserStatusPendingInvitation).Return(nil)
	invRepo.On("InvalidatePrevious", mock.Anything, "u1").Return(nil)
	invRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	email.On("Send", "carol@example.com", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := user.NewResendInvitationUsecase(repo, invRepo, email, audit, "http://localhost:3000")
	err := uc.Execute(context.Background(), "u1", "admin1", "admin@example.com")
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	repo.AssertExpectations(t)
	invRepo.AssertExpectations(t)
	email.AssertExpectations(t)
	audit.AssertExpectations(t)
}
