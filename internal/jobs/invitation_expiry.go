package jobs

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

// RunInvitationExpiry starts a daily goroutine that finds all users with
// status=pending_invitation whose invitation tokens have all expired,
// sets their status to expired_invitation, and fires an audit event.
func RunInvitationExpiry(pool *pgxpool.Pool, auditLogger domain.AuditLogger) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		runInvitationExpiry(pool, auditLogger)

		for range ticker.C {
			runInvitationExpiry(pool, auditLogger)
		}
	}()
}

func runInvitationExpiry(pool *pgxpool.Pool, auditLogger domain.AuditLogger) {
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT u.id, u.email
		FROM users u
		WHERE u.status = 'pending_invitation'
		  AND u.deleted_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM invitation_tokens it
			  WHERE it.user_id = u.id
				AND it.used_at IS NULL
				AND it.expires_at > NOW()
		  )
	`)
	if err != nil {
		log.Printf("[invitation-expiry] query error: %v", err)
		return
	}
	defer rows.Close()

	type expiredUser struct {
		ID    string
		Email string
	}
	var expiredUsers []expiredUser
	for rows.Next() {
		var u expiredUser
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			log.Printf("[invitation-expiry] scan error: %v", err)
			continue
		}
		expiredUsers = append(expiredUsers, u)
	}
	rows.Close()

	for _, u := range expiredUsers {
		_, err := pool.Exec(ctx,
			`UPDATE users SET status = 'expired_invitation', updated_at = NOW() WHERE id = $1`, u.ID)
		if err != nil {
			log.Printf("[invitation-expiry] update user %s error: %v", u.ID, err)
			continue
		}

		userID := u.ID
		_ = auditLogger.Log(ctx, &domain.AuditLog{
			ID:           uuid.NewString(),
			ActorEmail:   "system",
			Action:       domain.AuditUserInvitationExpired,
			ResourceType: "user",
			ResourceID:   &userID,
		})

		log.Printf("[invitation-expiry] user %s marked as expired_invitation", u.ID)
	}
}
