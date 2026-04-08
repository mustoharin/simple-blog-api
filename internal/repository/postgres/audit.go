package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type AuditLogRepository struct {
	db *pgxpool.Pool
}

func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Log implements domain.AuditLogger.
func (r *AuditLogRepository) Log(ctx context.Context, entry *domain.AuditLog) error {
	return r.Create(ctx, entry)
}

func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_id, actor_email, action, resource_type, resource_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.ActorID, log.ActorEmail, log.Action,
		log.ResourceType, log.ResourceID, log.IPAddress, log.UserAgent)
	if err != nil {
		return fmt.Errorf("audit log create: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) List(ctx context.Context, f domain.AuditFilter) ([]*domain.AuditLog, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if f.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", idx)
		args = append(args, f.ActorID)
		idx++
	}
	if f.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", idx)
		args = append(args, f.Action)
		idx++
	}
	if f.ResourceType != "" {
		where += fmt.Sprintf(" AND resource_type = $%d", idx)
		args = append(args, f.ResourceType)
		idx++
	}
	if f.ResourceID != "" {
		where += fmt.Sprintf(" AND resource_id = $%d", idx)
		args = append(args, f.ResourceID)
		idx++
	}
	if f.From != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, f.From)
		idx++
	}
	if f.To != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, f.To)
		idx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", where)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("audit log count: %w", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id, actor_id, actor_email, action, resource_type, resource_id, ip_address, user_agent, created_at
		FROM audit_logs %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1)

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit log list: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		l := &domain.AuditLog{}
		if err := rows.Scan(&l.ID, &l.ActorID, &l.ActorEmail, &l.Action,
			&l.ResourceType, &l.ResourceID, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("audit log scan: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

func (r *AuditLogRepository) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	l := &domain.AuditLog{}
	err := r.db.QueryRow(ctx, `
		SELECT id, actor_id, actor_email, action, resource_type, resource_id, ip_address, user_agent, created_at
		FROM audit_logs WHERE id = $1`, id,
	).Scan(&l.ID, &l.ActorID, &l.ActorEmail, &l.Action,
		&l.ResourceType, &l.ResourceID, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("audit log get by id: %w", err)
	}
	return l, nil
}

func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, days int) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM audit_logs WHERE created_at < NOW() - ($1 || ' days')::INTERVAL`, days)
	if err != nil {
		return fmt.Errorf("audit log delete older than: %w", err)
	}
	return nil
}
