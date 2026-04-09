# Simple Blog API

A RESTful API for a blog platform built with Go, PostgreSQL, and S3-compatible storage.

## Features

- User authentication and authorization with JWT
- Role-based access control
- Blog post management with tags
- Comment system
- Image upload to S3-compatible storage
- Email notifications
- Audit logging
- Soft delete with retention policies
- CAPTCHA support

## Prerequisites

- Docker and Docker Compose
- (Optional) Make for convenient commands

## Quick Start

### Using Make (Recommended)

```bash
# View all available commands
make help

# Complete setup and start all services
make setup

# Just start services
make up

# View logs
make logs

# Stop services
make down
```

### Using Docker Compose Directly

```bash
# Build and start all services
docker-compose up -d

# Create MinIO bucket
docker-compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker-compose exec minio mc mb local/blog-images
docker-compose exec minio mc anonymous set public local/blog-images

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Services

After running `make setup` or `docker-compose up`, the following services will be available:

- **API**: http://localhost:8080
- **PostgreSQL**: localhost:5432
  - Database: `blogapi`
  - User: `bloguser`
  - Password: `blogpass`
- **MinIO (S3)**: http://localhost:9000
  - Console: http://localhost:9001
  - Access Key: `minioadmin`
  - Secret Key: `minioadmin`

## Configuration

Copy `.env.example` to `.env` and adjust the values as needed:

```bash
cp .env.example .env
```

Key configuration options:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing (change in production!)
- `S3_*`: S3/MinIO configuration
- `SMTP_*`: Email server configuration (optional)
- `CAPTCHA_*`: CAPTCHA configuration (optional)

## Development

### Run without Docker

```bash
# Make sure PostgreSQL and MinIO are running
make up postgres minio

# Run the API locally
make dev
```

### Run Tests

```bash
make test
```

### Database Migrations

Migrations are run automatically when the API container starts. To run them manually:

```bash
# Run all migrations
make migrate-up

# Rollback last migration
make migrate-down
```

Migration files are located in the `migrations/` directory.

## API Documentation

The API follows RESTful conventions. Main endpoints:

- `/api/v1/auth/*` - Authentication (register, login, forgot password, etc.)
- `/api/v1/users/*` - User management
- `/api/v1/profile/*` - Current user profile (also `/api/v1/me/*`)
- `/api/v1/posts` - List posts, create post
- `/api/v1/posts/:id/*` - Post operations by ID (update, delete, publish)
- `/api/v1/posts/slug/:slug` - Get post by slug
- `/api/v1/posts/:id/comments` - Comments for a specific post
- `/api/v1/tags/*` - Tags
- `/api/v1/comments/*` - Comment management
- `/api/v1/images/*` - Image upload
- `/api/v1/audit-logs/*` - Audit logs
- `/api/v1/dashboard` - Dashboard statistics

### Example API Calls

```bash
# List all posts
curl http://localhost:8080/api/v1/posts

# Get a post by slug
curl http://localhost:8080/api/v1/posts/slug/my-blog-post

# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'
```

## Project Structure

```
.
├── cmd/
│   └── api/          # Application entry point
├── config/           # Configuration loading
├── internal/
│   ├── delivery/     # HTTP handlers and routing
│   ├── usecase/      # Business logic
│   ├── repository/   # Data access layer
│   ├── storage/      # S3 storage
│   ├── pkg/          # Shared packages
│   └── jobs/         # Background jobs
├── migrations/       # Database migrations
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Troubleshooting

### Container won't start

```bash
# Check logs
make logs

# Rebuild images
make clean
make build
make up
```

### Database connection issues

```bash
# Check PostgreSQL is healthy
docker-compose ps

# Check database logs
make logs-db
```

### MinIO bucket issues

```bash
# Recreate bucket
make create-bucket
```

## Production Deployment

For production:

1. Change `JWT_SECRET` to a strong random value
2. Update `DATABASE_URL` to point to your production database
3. Configure proper S3 credentials (AWS S3, DigitalOcean Spaces, etc.)
4. Set up SMTP for email notifications
5. Configure CAPTCHA provider
6. Set `ENV=production`
7. Update `ALLOWED_ORIGINS` to your frontend domain
8. Use proper SSL/TLS certificates
9. Set up monitoring and logging

## License

[Your License Here]
