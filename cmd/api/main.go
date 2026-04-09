// @title           Simple Blog API
// @version         1.0
// @description     A production-ready personal blog REST API
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Enter: Bearer {token}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"simple-blog-api/config"
	deliveryhttp "simple-blog-api/internal/delivery/http"
	"simple-blog-api/internal/delivery/http/handler"
	"simple-blog-api/internal/jobs"
	"simple-blog-api/internal/pkg/email"
	"simple-blog-api/internal/pkg/password"
	"simple-blog-api/internal/repository/postgres"
	"simple-blog-api/internal/storage/s3"
	audituc "simple-blog-api/internal/usecase/audit"
	authuc "simple-blog-api/internal/usecase/auth"
	commentuc "simple-blog-api/internal/usecase/comment"
	dashboarduc "simple-blog-api/internal/usecase/dashboard"
	imageuc "simple-blog-api/internal/usecase/image"
	postuc "simple-blog-api/internal/usecase/post"
	profileuc "simple-blog-api/internal/usecase/profile"
	taguc "simple-blog-api/internal/usecase/tag"
	useruc "simple-blog-api/internal/usecase/user"
)

// authEmailAdapter adapts *email.Sender to authuc.EmailSender.
type authEmailAdapter struct{ s *email.Sender }

func (a *authEmailAdapter) Send(msg authuc.EmailMessage) error {
	return a.s.Send(email.Message{
		To:      msg.To,
		Subject: msg.Subject,
		Body:    msg.Body,
	})
}

// userEmailAdapter adapts *email.Sender to useruc.EmailSender.
type userEmailAdapter struct{ s *email.Sender }

func (a *userEmailAdapter) Send(to, subject, body string) error {
	return a.s.Send(email.Message{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Database
	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer pool.Close()

	// S3 uploader
	s3Uploader, err := s3.NewUploader(ctx, cfg.S3Region, cfg.S3Bucket, cfg.S3Endpoint,
		cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		log.Fatalf("init s3: %v", err)
	}

	// Email sender
	emailSender := email.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.EmailFrom)
	authEmail := &authEmailAdapter{s: emailSender}
	userEmail := &userEmailAdapter{s: emailSender}

	// Repositories
	userRepo         := postgres.NewUserRepository(pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(pool)
	prtRepo          := postgres.NewPasswordResetTokenRepository(pool)
	invTokenRepo     := postgres.NewInvitationTokenRepository(pool)
	auditRepo        := postgres.NewAuditLogRepository(pool)
	postRepo         := postgres.NewPostRepository(pool)
	tagRepo          := postgres.NewTagRepository(pool)
	postTagRepo      := postgres.NewPostTagRepository(pool)
	commentRepo      := postgres.NewCommentRepository(pool)
	imageRepo        := postgres.NewImageRepository(pool)
	dashboardRepo    := postgres.NewDashboardRepository(pool)

	// Password validator (nil = use default HIBP endpoint)
	pwValidator := password.NewValidator(nil)

	// Captcha verifier
	captchaVerifier, err := authuc.NewCaptchaVerifier(cfg.CaptchaProvider, cfg.CaptchaSecret)
	if err != nil {
		log.Fatalf("init captcha: %v", err)
	}

	// Auth usecases
	registerUC  := authuc.NewRegisterUsecase(userRepo, pwValidator, auditRepo)
	loginUC     := authuc.NewLoginUsecase(userRepo, refreshTokenRepo, captchaVerifier, auditRepo,
		cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	refreshUC   := authuc.NewRefreshUsecase(userRepo, refreshTokenRepo, auditRepo,
		cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	logoutUC    := authuc.NewLogoutUsecase(refreshTokenRepo, auditRepo)
	forgotPWUC  := authuc.NewForgotPasswordUsecase(userRepo, prtRepo, authEmail, auditRepo, cfg.FrontendURL)
	resetPWUC   := authuc.NewResetPasswordUsecase(userRepo, prtRepo, refreshTokenRepo, pwValidator, auditRepo)
	acceptInvUC := authuc.NewAcceptInvitationUsecase(userRepo, invTokenRepo, pwValidator, auditRepo)

	// User usecases
	createUserUC := useruc.NewCreateUserUsecase(userRepo, invTokenRepo, userEmail, auditRepo, cfg.FrontendURL)
	listUsersUC  := useruc.NewListUsersUsecase(userRepo)
	updateUserUC := useruc.NewUpdateUserUsecase(userRepo, auditRepo)
	deleteUserUC := useruc.NewDeleteUserUsecase(userRepo, refreshTokenRepo, auditRepo)
	assignRoleUC := useruc.NewAssignRoleUsecase(userRepo, auditRepo)
	removeRoleUC := useruc.NewRemoveRoleUsecase(userRepo, auditRepo)
	resendInvUC  := useruc.NewResendInvitationUsecase(userRepo, invTokenRepo, userEmail, auditRepo, cfg.FrontendURL)

	// Profile usecases
	getMeUC    := profileuc.NewGetMeUsecase(userRepo)
	updateMeUC := profileuc.NewUpdateMeUsecase(userRepo, auditRepo)
	changePWUC := profileuc.NewChangePasswordUsecase(userRepo, refreshTokenRepo, pwValidator, auditRepo)

	// Post usecases
	createPostUC    := postuc.NewCreatePostUsecase(postRepo, postTagRepo, auditRepo)
	listPostsUC     := postuc.NewListPostsUsecase(postRepo, postTagRepo)
	getBySlugUC     := postuc.NewGetBySlugUsecase(postRepo, postTagRepo)
	updatePostUC    := postuc.NewUpdatePostUsecase(postRepo, postTagRepo, auditRepo)
	togglePublishUC := postuc.NewTogglePublishUsecase(postRepo, auditRepo)
	deletePostUC    := postuc.NewDeletePostUsecase(postRepo, commentRepo, auditRepo)

	// Tag usecases
	createTagUC := taguc.NewCreateTagUsecase(tagRepo, auditRepo)
	listTagsUC  := taguc.NewListTagsUsecase(tagRepo)
	deleteTagUC := taguc.NewDeleteTagUsecase(tagRepo, auditRepo)

	// Comment usecases
	createCommentUC       := commentuc.NewCreateCommentUsecase(postRepo, commentRepo, auditRepo)
	listCommentsUC        := commentuc.NewListCommentsUsecase(commentRepo)
	updateCommentStatusUC := commentuc.NewUpdateStatusUsecase(commentRepo, auditRepo)
	deleteCommentUC       := commentuc.NewDeleteCommentUsecase(commentRepo, auditRepo)

	// Image usecases
	uploadImageUC := imageuc.NewUploadImageUsecase(imageRepo, s3Uploader, auditRepo)
	deleteImageUC := imageuc.NewDeleteImageUsecase(imageRepo, s3Uploader, auditRepo)

	// Audit usecases
	listAuditUC := audituc.NewListAuditLogsUsecase(auditRepo)
	getAuditUC  := audituc.NewGetAuditLogUsecase(auditRepo)

	// Dashboard usecase
	getDashboardUC := dashboarduc.NewGetDashboardUsecase(dashboardRepo)

	// HTTP handlers
	authHandler      := handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, forgotPWUC, resetPWUC, acceptInvUC)
	userHandler      := handler.NewUserHandler(createUserUC, listUsersUC, updateUserUC, deleteUserUC, assignRoleUC, removeRoleUC, resendInvUC)
	profileHandler   := handler.NewProfileHandler(getMeUC, updateMeUC, changePWUC)
	postHandler      := handler.NewPostHandler(createPostUC, listPostsUC, getBySlugUC, updatePostUC, togglePublishUC, deletePostUC)
	tagHandler       := handler.NewTagHandler(createTagUC, listTagsUC, deleteTagUC)
	commentHandler   := handler.NewCommentHandler(createCommentUC, listCommentsUC, updateCommentStatusUC, deleteCommentUC)
	imageHandler     := handler.NewImageHandler(uploadImageUC, deleteImageUC)
	auditHandler     := handler.NewAuditHandler(listAuditUC, getAuditUC)
	dashboardHandler := handler.NewDashboardHandler(getDashboardUC)

	// Run database migrations before starting jobs or serving traffic
	runMigrations(cfg.DatabaseURL)
	seedSuperAdmin(context.Background(), pool, cfg) // seed default superadmin

	// Background jobs
	jobs.RunRetentionPurge(pool, cfg.AuditLogRetentionDays, cfg.SoftDeleteRetentionDays)
	jobs.RunInvitationExpiry(pool, auditRepo)

	// Setup router
	router := deliveryhttp.SetupRouter(deliveryhttp.RouterDeps{
		JWTSecret:      cfg.JWTSecret,
		AllowedOrigins: cfg.AllowedOrigins,
		Auth:           authHandler,
		User:           userHandler,
		Profile:        profileHandler,
		Post:           postHandler,
		Tag:            tagHandler,
		Comment:        commentHandler,
		Image:          imageHandler,
		Audit:          auditHandler,
		Dashboard:      dashboardHandler,
	})

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("simple-blog-api listening on %s (env=%s)", addr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown: %v", err)
	}
	log.Println("server stopped")
}

func seedSuperAdmin(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) {
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		log.Println("[seed] SEED_ADMIN_EMAIL or SEED_ADMIN_PASSWORD not set, skipping superadmin seed")
		return
	}

	// Check if user already exists — skip bcrypt if so (avoids ~100ms hash on every restart)
	var userID string
	existErr := db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`,
		cfg.SeedAdminEmail).Scan(&userID)
	if existErr != nil {
		// User doesn't exist — hash and insert
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.SeedAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[seed] failed to hash password: %v", err)
			return
		}

		// Insert new superadmin user; skip if already exists
		err = db.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, display_name, status)
			VALUES ($1, $2, 'Super Admin', 'active')
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		`, cfg.SeedAdminEmail, string(hash)).Scan(&userID)

		if err != nil {
			// pgx returns pgx.ErrNoRows when ON CONFLICT DO NOTHING fires (no row returned)
			// In that case look up the existing user ID
			if !errors.Is(err, pgx.ErrNoRows) {
				log.Printf("[seed] failed to upsert superadmin user: %v", err)
				return
			}
			// User already exists — fetch their ID
			err = db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`,
				cfg.SeedAdminEmail).Scan(&userID)
			if err != nil {
				log.Printf("[seed] failed to look up existing superadmin: %v", err)
				return
			}
		}
	}

	// Resolve superadmin role ID explicitly so we detect if it's missing
	var roleID string
	roleErr := db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'superadmin'`).Scan(&roleID)
	if roleErr != nil {
		log.Printf("[seed] superadmin role not found in roles table: %v", roleErr)
		return
	}

	_, assignErr := db.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID)
	if assignErr != nil {
		log.Printf("[seed] failed to assign superadmin role: %v", assignErr)
		return
	}

	log.Printf("[seed] superadmin user ready: %s", cfg.SeedAdminEmail)
}

func runMigrations(databaseURL string) {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Fatalf("migrations init: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrations up: %v", err)
	}
	log.Println("migrations applied")
}
