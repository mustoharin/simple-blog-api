package jobs

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunRetentionPurge starts a daily goroutine that:
// 1. Deletes audit logs older than auditRetentionDays.
// 2. Hard-deletes records that were soft-deleted more than softDeleteRetentionDays ago.
func RunRetentionPurge(pool *pgxpool.Pool, auditRetentionDays, softDeleteRetentionDays int) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		runPurge(pool, auditRetentionDays, softDeleteRetentionDays)

		for range ticker.C {
			runPurge(pool, auditRetentionDays, softDeleteRetentionDays)
		}
	}()
}

func runPurge(pool *pgxpool.Pool, auditDays, softDays int) {
	ctx := context.Background()

	tag, err := pool.Exec(ctx,
		`DELETE FROM audit_logs WHERE created_at < NOW() - make_interval(days => $1)`, auditDays)
	if err != nil {
		log.Printf("[retention] audit log purge error: %v", err)
	} else {
		log.Printf("[retention] purged %d audit log rows (older than %d days)", tag.RowsAffected(), auditDays)
	}

	tag, err = pool.Exec(ctx,
		`DELETE FROM posts WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - make_interval(days => $1)`, softDays)
	if err != nil {
		log.Printf("[retention] posts purge error: %v", err)
	} else if tag.RowsAffected() > 0 {
		log.Printf("[retention] hard-deleted %d posts", tag.RowsAffected())
	}

	tag, err = pool.Exec(ctx,
		`DELETE FROM comments WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - make_interval(days => $1)`, softDays)
	if err != nil {
		log.Printf("[retention] comments purge error: %v", err)
	} else if tag.RowsAffected() > 0 {
		log.Printf("[retention] hard-deleted %d comments", tag.RowsAffected())
	}

	tag, err = pool.Exec(ctx,
		`DELETE FROM users WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - make_interval(days => $1)`, softDays)
	if err != nil {
		log.Printf("[retention] users purge error: %v", err)
	} else if tag.RowsAffected() > 0 {
		log.Printf("[retention] hard-deleted %d users", tag.RowsAffected())
	}
}
