package user

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type RemoveRoleUsecase struct {
	users domain.UserRepository
	audit domain.AuditLogger
}

func NewRemoveRoleUsecase(users domain.UserRepository, audit domain.AuditLogger) *RemoveRoleUsecase {
	return &RemoveRoleUsecase{users: users, audit: audit}
}

func (uc *RemoveRoleUsecase) Execute(ctx context.Context, userID, roleID, actorID, actorEmail string) error {
	// Guard: only superadmins may remove the superadmin role.
	superadminRole, err := uc.users.GetRoleByName(ctx, "superadmin")
	if err == nil && superadminRole.ID == roleID {
		actorRoles, err := uc.users.GetRoles(ctx, actorID)
		if err != nil {
			return err
		}
		isSuperadmin := false
		for _, r := range actorRoles {
			if r.Name == "superadmin" {
				isSuperadmin = true
				break
			}
		}
		if !isSuperadmin {
			return domain.ErrForbidden
		}
	}

	if err := uc.users.RemoveRole(ctx, userID, roleID); err != nil {
		return err
	}
	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditUserRoleRemoved,
			ResourceType: "user",
			ResourceID:   &userID,
		})
	}()
	return nil
}
