# Quick Start Guide

This guide will help you get the Simple Blog API up and running in minutes.

## Prerequisites

- Docker Desktop installed and running
- (Optional) Make utility installed

## Steps

### 1. Start the Services

Using Make (recommended):
```bash
cd /path/to/simple-blog-api/.worktrees/development
make setup
```

Or using Docker Compose directly:
```bash
cd /path/to/simple-blog-api/.worktrees/development
docker compose up -d
sleep 10
docker compose exec -T minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker compose exec -T minio mc mb local/blog-images
docker compose exec -T minio mc anonymous set public local/blog-images
```

### 2. Verify Services are Running

```bash
docker compose ps
```

You should see:
- `blog-api` - Running on port 8080
- `blog-api-db` - PostgreSQL on port 5432 (healthy)
- `blog-api-s3` - MinIO on ports 9000-9001 (healthy)

### 3. Test the API

```bash
# List posts (should return empty list initially)
curl http://localhost:8080/api/v1/posts

# Expected output:
# {"limit":20,"page":1,"posts":null,"total":0}
```

### 4. Access MinIO Console

- URL: http://localhost:9001
- Username: `minioadmin`
- Password: `minioadmin`

## Services URLs

- **API**: http://localhost:8080
- **PostgreSQL**: localhost:5432
  - Database: `blogapi`
  - User: `bloguser`
  - Password: `blogpass`
- **MinIO Console**: http://localhost:9001
- **MinIO API**: http://localhost:9000

## Next Steps

1. **Register a User**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@example.com",
       "password": "SecurePass123!",
       "display_name": "Admin User"
     }'
   ```

2. **Login**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@example.com",
       "password": "SecurePass123!"
     }'
   ```
   
   Save the returned `access_token` for authenticated requests.

3. **Create a Post** (requires authentication):
   ```bash
   curl -X POST http://localhost:8080/api/v1/posts \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     -d '{
       "title": "My First Post",
       "slug": "my-first-post",
       "excerpt": "This is my first blog post",
       "content": "Full content here...",
       "published": true
     }'
   ```

## Common Commands

```bash
# View API logs
docker compose logs -f api

# View all logs
docker compose logs -f

# Restart API
docker compose restart api

# Stop all services
docker compose down

# Stop and remove all data
docker compose down -v
```

## Troubleshooting

### API won't start
```bash
# Check logs
docker compose logs api

# Rebuild
docker compose build api
docker compose up -d --force-recreate api
```

### Database connection issues
```bash
# Check PostgreSQL is running
docker compose ps postgres

# Check database logs
docker compose logs postgres
```

### Can't upload images
1. Verify MinIO is running: `docker compose ps minio`
2. Ensure bucket exists: `docker compose exec minio mc ls local/`
3. Recreate bucket: `make create-bucket`

## Environment Configuration

To customize configuration:
1. Copy `.env.example` to `.env`
2. Edit values as needed
3. Restart services: `docker compose down && docker compose up -d`

## Development Mode

Run locally without Docker (requires PostgreSQL and MinIO running):
```bash
# Start only database services
docker compose up -d postgres minio

# Copy environment file
cp .env.example .env

# Edit DATABASE_URL and S3_ENDPOINT in .env
# Then run:
go run cmd/api/main.go
```

## Need Help?

See the main [README.md](README.md) for detailed documentation.
