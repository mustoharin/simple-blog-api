package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simple-blog-api/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, bio, avatar_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.PasswordHash,
		user.DisplayName, user.Bio, user.AvatarURL, user.Status,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") && strings.Contains(err.Error(), "email") {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("user create: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, bio, avatar_url,
		       last_login_at, status, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Bio, &u.AvatarURL,
		&u.LastLoginAt, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user get by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, bio, avatar_url,
		       last_login_at, status, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Bio, &u.AvatarURL,
		&u.LastLoginAt, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user get by email: %w", err)
	}
	return u, nil
}

func (r *UserRepository) List(ctx context.Context, page, limit int) ([]*domain.User, int, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user count: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, `
		SELECT id, email, display_name, bio, avatar_url, last_login_at, status, created_at, updated_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user list: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Bio, &u.AvatarURL,
			&u.LastLoginAt, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("user scan: %w", err)
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET display_name=$1, bio=$2, avatar_url=$3, updated_at=NOW()
		WHERE id=$4 AND deleted_at IS NULL`,
		user.DisplayName, user.Bio, user.AvatarURL, user.ID)
	if err != nil {
		return fmt.Errorf("user update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("user soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
		status, id)
	if err != nil {
		return fmt.Errorf("user update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleID)
	return err
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2`, userID, roleID)
	if err != nil {
		return fmt.Errorf("remove role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.description
		FROM roles r JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}
	defer rows.Close()
	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *UserRepository) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}
	defer rows.Close()
	var perms []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("scan perm: %w", err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *UserRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description FROM roles WHERE name = $1`, name,
	).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get role by name: %w", err)
	}
	return role, nil
}
