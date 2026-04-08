# Simple Blog API — Design Spec

**Date:** 2026-04-08  
**Status:** Approved  

---

## Problem Statement

Build a production-ready personal blog REST API in Go. The API will serve as the backend for a personal blog that will be deployed and actively used. It must support writing and publishing posts, tagging, moderated comments, S3 image uploads, RBAC-based authorization, and full-text search.

---

## Architecture Approach

**Clean Architecture (Layered)** with strict inward dependency flow:

```
delivery/http  →  usecase  →  repository  →  (database)
                     ↓
                  domain
```

Each layer only imports inward. The `domain` package has zero external dependencies.

---

## Project Structure

```
simple-blog-api/
├── cmd/api/                  # main.go — wires everything together (DI root)
├── internal/
│   ├── domain/               # Pure Go structs + repository/usecase interfaces
│   ├── usecase/              # Business logic implementations
│   ├── repository/           # PostgreSQL implementations of domain repo interfaces
│   ├── delivery/
│   │   └── http/             # Gin route handlers + middleware (JWT, RBAC, CORS, recovery)
│   └── storage/              # S3/S3-compatible upload logic
├── migrations/               # SQL migration files (managed with golang-migrate or goose)
├── config/                   # Config struct loaded from environment variables
└── docs/                     # Design specs, OpenAPI (future)
```

---

## Data Models

### User
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| email | text | Unique |
| password_hash | text | bcrypt |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Role
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| name | text | Unique (admin, editor, reader) |
| description | text | |

### Permission
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| name | text | Unique (e.g. `post:create`, `comment:approve`) |

### RolePermission (join)
`role_id` → `permission_id`

### UserRole (join)
`user_id` → `role_id`

### Post
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| title | text | |
| slug | text | Unique, URL-safe |
| content | text | Markdown |
| excerpt | text | Short summary |
| cover_image_url | text | Nullable |
| status | enum | `draft` \| `published` |
| author_id | UUID | FK → User |
| published_at | timestamptz | Nullable |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| search_vector | tsvector | Auto-updated via trigger for full-text search |

### Tag
| Field | Type |
|---|---|
| id | UUID |
| name | text (unique) |
| slug | text (unique) |

### PostTag (join)
`post_id` → `tag_id`

### Comment
| Field | Type | Notes |
|---|---|---|
| id | UUID | |
| post_id | UUID | FK → Post |
| author_id | UUID | FK → User (login required) |
| body | text | |
| status | enum | `pending` \| `approved` \| `rejected` |
| created_at | timestamptz | |

### Image
| Field | Type | Notes |
|---|---|---|
| id | UUID | |
| filename | text | Original filename |
| s3_key | text | Key in the S3 bucket |
| url | text | Public URL |
| uploaded_by | UUID | FK → User |
| created_at | timestamptz | |

### RefreshToken
| Field | Type | Notes |
|---|---|---|
| id | UUID | |
| user_id | UUID | FK → User |
| token_hash | text | SHA-256 of the raw token |
| expires_at | timestamptz | e.g. 30 days from creation |
| revoked_at | timestamptz | Nullable; set on rotation or logout |
| created_at | timestamptz | |

On refresh: old token is revoked and a new one is issued (rotation). On logout: token is revoked.

### PasswordResetToken
| Field | Type | Notes |
|---|---|---|
| id | UUID | |
| user_id | UUID | FK → User |
| token_hash | text | SHA-256 of the raw token |
| expires_at | timestamptz | 1 hour from creation |
| used_at | timestamptz | Nullable; set on consumption |

---

## RBAC — Default Roles & Permissions

| Permission | admin | editor | reader |
|---|:---:|:---:|:---:|
| `post:create` | ✓ | ✓ | |
| `post:edit` | ✓ | ✓ | |
| `post:delete` | ✓ | ✓ | |
| `post:publish` | ✓ | ✓ | |
| `image:upload` | ✓ | ✓ | |
| `comment:create` | ✓ | ✓ | ✓ |
| `comment:approve` | ✓ | ✓ | |
| `user:manage` | ✓ | | |

JWT tokens encode the user's resolved set of permission names. RBAC middleware checks the token claims — no DB hit per request.

---

## API Endpoints

All routes are prefixed with `/api/v1`.

### Auth (public)
| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create a new account |
| POST | `/auth/login` | Login with email + password + captcha token → JWT + refresh token |
| POST | `/auth/refresh` | Exchange refresh token for new JWT |
| POST | `/auth/forgot-password` | Send password reset email |
| POST | `/auth/reset-password` | Consume reset token + set new password |
| POST | `/auth/logout` | Revoke refresh token (requires valid JWT) |

### Posts
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/posts` | public | List published posts. Query params: `q`, `tag`, `author`, `page`, `limit`, `sort` |
| GET | `/posts/:slug` | public | Get post by slug |
| POST | `/posts` | `post:create` | Create post |
| PUT | `/posts/:id` | `post:edit` | Update post |
| PATCH | `/posts/:id/publish` | `post:publish` | Toggle publish status |
| DELETE | `/posts/:id` | `post:delete` | Delete post |

### Tags
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/tags` | public | List all tags |
| POST | `/tags` | `post:create` | Create tag |
| DELETE | `/tags/:id` | `post:delete` | Delete tag |

### Comments
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/posts/:id/comments` | public | List approved comments for a post |
| POST | `/posts/:id/comments` | `comment:create` | Submit comment (status = pending) |
| PATCH | `/comments/:id/status` | `comment:approve` | Approve or reject a comment |
| DELETE | `/comments/:id` | `comment:approve` | Delete comment |

### Images
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/images` | `image:upload` | Upload image → S3, returns URL |
| DELETE | `/images/:id` | `image:upload` | Delete image from S3 + DB |

### Users (admin)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/users` | `user:manage` | List users |
| POST | `/users/:id/roles` | `user:manage` | Assign role to user |
| DELETE | `/users/:id/roles/:roleId` | `user:manage` | Remove role from user |

---

## Key Implementation Details

### Authentication
- Passwords hashed with **bcrypt** (cost 12)
- **JWT** (access token, short-lived e.g. 15 min) + **refresh token** (long-lived, stored in DB, rotated on use)
- Login requires a valid **CAPTCHA token** (hCaptcha or Google reCAPTCHA v2/v3), verified server-side
- Password reset tokens are single-use, hashed before storage, expire after 1 hour

### Full-Text Search
- PostgreSQL `tsvector` column on `posts`, populated via a DB trigger
- `GIN` index on `search_vector` for fast queries
- `GET /posts?q=keyword` uses `to_tsquery` + `ts_rank` for relevance ordering

### Image Uploads
- Storage backend: **S3-compatible** (Cloudflare R2, MinIO, AWS S3)
- Files are validated (MIME type, max size e.g. 10 MB) before upload
- The DB record stores `s3_key` and `url`; deletion removes both

### Email
- Sent via **SMTP** (provider-agnostic: SendGrid, Resend, Mailgun, self-hosted)
- Transactional emails: password reset link
- HTML + plain text templates

### CORS
- Configurable allowed origins via env var
- Gin CORS middleware applied globally

---

## Configuration (Environment Variables)

```
PORT                    # Default: 8080
ENV                     # development | production

DATABASE_URL            # postgres://user:pass@host:5432/dbname

JWT_SECRET
JWT_ACCESS_EXPIRY       # e.g. 15m
JWT_REFRESH_EXPIRY      # e.g. 720h (30 days)

S3_ENDPOINT
S3_BUCKET
S3_REGION
S3_ACCESS_KEY
S3_SECRET_KEY

SMTP_HOST
SMTP_PORT
SMTP_USER
SMTP_PASS
EMAIL_FROM

CAPTCHA_PROVIDER        # recaptcha | hcaptcha
CAPTCHA_SECRET

ALLOWED_ORIGINS         # comma-separated list
```

---

## Error Handling

All errors return a consistent JSON envelope:

```json
{
  "error": "Human-readable message",
  "code": "MACHINE_READABLE_CODE"
}
```

- `400` — validation errors (include field-level details under `"fields"`)
- `401` — unauthenticated
- `403` — authenticated but lacks permission
- `404` — resource not found
- `409` — conflict (e.g. duplicate slug)
- `500` — internal error (generic message, no stack trace in production)

A Gin recovery middleware catches panics and logs them with a request ID.

---

## Testing Strategy

- **Unit tests** for all usecases — mock repository interfaces using `testify/mock` or `gomock`
- **Integration tests** for HTTP handlers — `net/http/httptest` + real test PostgreSQL (Docker Compose)
- **Table-driven tests** for validation logic
- Test coverage target: ≥ 80% on usecase layer

---

## Coding Guidelines

### SAST — Static Analysis with gosec

**Tool:** [`securego/gosec`](https://github.com/securego/gosec) — Go-native static security analysis.

**CI enforcement:** A dedicated GitHub Actions job runs `gosec ./...` on every push and pull request targeting `main`. The job is required — merges are blocked if any non-suppressed rule fires.

**Enforced rules (must never be suppressed without justification):**

| Rule | Description |
|---|---|
| `G101` | Hardcoded credentials in source code |
| `G104` | Unhandled errors — all errors must be handled or explicitly ignored with `_` + comment |
| `G201/G202` | SQL string formatting — use parameterized queries only; string concatenation in SQL is forbidden |
| `G304` | File path injection — sanitize all user-supplied values used in file/S3 key construction |
| `G401` | Weak crypto hash (MD5, SHA1) — forbidden for any security-sensitive operation |
| `G501` | Use `crypto/rand`, never `math/rand`, for token/secret generation |

**Suppression policy:** `//nolint:gosec` annotations are allowed only with an inline comment explaining why the suppression is safe. Blanket file-level or package-level suppressions are forbidden.

---

### XSS Sanitization

**Library:** [`microcosm-cc/bluemonday`](https://github.com/microcosm-cc/bluemonday)

**Where it runs:** Sanitization is applied in the **usecase layer**, not in HTTP handlers. This ensures it runs regardless of how input enters the system (HTTP, tests, scripts).

**Policy tiers:**

| Field(s) | Policy | Reason |
|---|---|---|
| Post `content` | `bluemonday.UGCPolicy()` | Markdown renders to HTML; safe subset of tags is allowed |
| Post `title`, `excerpt` | `bluemonday.StrictPolicy()` | Plain text only |
| Comment `body` | `bluemonday.StrictPolicy()` | No HTML in comments |
| Tag `name`, `slug` | `bluemonday.StrictPolicy()` | Plain text |
| User `email` | `bluemonday.StrictPolicy()` | Plain text |
| All other text inputs | `bluemonday.StrictPolicy()` | Default: strip all HTML |

`UGCPolicy` allows safe formatting tags (`<b>`, `<i>`, `<em>`, `<a>`, `<p>`, `<ul>`, `<ol>`, `<li>`, `<code>`, `<pre>`, `<blockquote>`) and strips dangerous attributes (e.g., `onclick`, `style`, `javascript:` hrefs).

---

### Input Trim Validation

**Rule:** All user-facing string inputs are trimmed of leading and trailing whitespace using `strings.TrimSpace` before any validation or persistence.

**A shared helper** `sanitize.Trim(s string) string` lives in `internal/pkg/sanitize` and is used consistently throughout the usecase layer.

**Fields where trim is applied:**

| Category | Fields |
|---|---|
| Post | `title`, `slug`, `excerpt`, `cover_image_url` |
| Tag | `name`, `slug` |
| Comment | `body` |
| User | `email` |
| Auth inputs | `email` (all auth endpoints) |
| Query parameters | `q`, `tag`, `author` |

**Exception:** Post `content` — leading/trailing whitespace in a Markdown document is intentional and is not trimmed.

**Order of operations (applied per field in usecases):**
1. **Trim** — `strings.TrimSpace`
2. **Sanitize** — XSS clean via `bluemonday`
3. **Validate** — check required, length, format constraints

---

## Dependencies (planned)

| Package | Purpose |
|---|---|
| `github.com/gin-gonic/gin` | HTTP router |
| `github.com/golang-jwt/jwt/v5` | JWT |
| `github.com/lib/pq` or `pgx/v5` | PostgreSQL driver |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/aws/aws-sdk-go-v2/service/s3` | S3 uploads |
| `github.com/golang-migrate/migrate/v4` | DB migrations |
| `github.com/stretchr/testify` | Test assertions |
| `github.com/google/uuid` | UUID generation |
| `github.com/microcosm-cc/bluemonday` | XSS HTML sanitization |
| `github.com/securego/gosec/v2` | SAST (dev/CI tool) |
