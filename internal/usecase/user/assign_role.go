package user

import (
	"context"

	"github.com/google/uuid"

	"simple-blog-api/internal/domain"
)

type AssignRoleUsecase struct {
	users domain.UserRepository
	audit domain.AuditLogger
}

func NewAssignRoleUsecase(users domain.UserRepository, audit domain.AuditLogger) *AssignRoleUsecase {
	return &AssignRoleUsecase{users: users, audit: audit}
}

func (uc *AssignRoleUsecase) Execute(ctx context.Context, userID, roleID, actorID, actorEmail string) error {
	if _, err := uc.users.GetByID(ctx, userID); err != nil {
		return err
	}

	// Guard: only superadmins may assign the superadmin role.
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

	if err := uc.users.AssignRole(ctx, userID, roleID); err != nil {
		return err
	}
	go func() {
		_ = uc.audit.Log(context.WithoutCancel(ctx), &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorID:      &actorID,
			ActorEmail:   actorEmail,
			Action:       domain.AuditUserRoleAssigned,
			ResourceType: "user",
			ResourceID:   &userID,
		})
	}()
	return nil
}
