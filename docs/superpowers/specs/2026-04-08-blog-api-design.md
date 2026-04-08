# Simple Blog API — Design Spec

**Date:** 2026-04-08  
**Status:** Approved  

---

## Problem Statement

Build a production-ready personal blog REST API in Go. The API will serve as the backend for a personal blog that will be deployed and actively used. It must support writing and publishing posts, tagging, moderated comments, S3 image uploads, RBAC-based authorization (superadmin / admin / editor / reader roles), full-text search, and a tamper-evident audit log of all admin activity.

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
│   ├── storage/              # S3/S3-compatible upload logic
│   └── pkg/
│       ├── sanitize/         # Shared trim + XSS sanitization helpers
│       └── password/         # Password complexity validator + HIBP breach check client
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
| display_name | text | Nullable; user's public name |
| bio | text | Nullable; short biography |
| avatar_url | text | Nullable; URL to profile image |
| last_login_at | timestamptz | Nullable; updated on every successful login |
| email_verified_at | timestamptz | Nullable; set on email verification; NULL = unverified (cannot login) |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | Nullable; set on soft delete; NULL = active |

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
| view_count | bigint | Default 0; incremented atomically on each public fetch |
| comment_count | int | Default 0; maintained by a DB trigger on the comments table |
| deleted_at | timestamptz | Nullable; set on soft delete; NULL = active |

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
| deleted_at | timestamptz | Nullable; set on soft delete; NULL = active |

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

### UserVerificationToken
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | FK → User |
| token_hash | text | SHA-256 of raw token |
| type | enum | `email_verification` \| `invitation` |
| expires_at | timestamptz | 24 hours from creation |
| used_at | timestamptz | Nullable; set on consumption |

One active token per user per type. Issuing a new token for the same user+type invalidates the previous one (`used_at = NOW()` before inserting).

### AuditLog
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| actor_id | UUID | FK → User; nullable (null for unauthenticated events like failed login) |
| actor_email | text | Snapshot of email at time of action (survives user deletion) |
| action | text | Event code e.g. `post.created`, `user.login` |
| resource_type | text | `post`, `comment`, `tag`, `image`, `user`, `session` |
| resource_id | text | UUID of the affected resource (nullable for session events) |
| ip_address | text | Client IP from request |
| user_agent | text | Client user-agent header |
| created_at | timestamptz | Indexed for date-range filtering and retention purge |

Indexes: `(actor_id)`, `(action)`, `(resource_type)`, `(created_at)`.

---

## RBAC — Default Roles & Permissions

| Permission | superadmin | admin | editor | reader |
|---|:---:|:---:|:---:|:---:|
| `post:create` | ✓ | ✓ | ✓ | |
| `post:edit` | ✓ | ✓ | ✓ | |
| `post:delete` | ✓ | ✓ | ✓ | |
| `post:publish` | ✓ | ✓ | ✓ | |
| `image:upload` | ✓ | ✓ | ✓ | |
| `comment:create` | ✓ | ✓ | ✓ | ✓ |
| `comment:approve` | ✓ | ✓ | ✓ | |
| `user:manage` | ✓ | ✓ | | | ← role assignment only |
| `user:create` | ✓ | | | | ← superadmin only |
| `user:update` | ✓ | | | | ← superadmin only |
| `user:delete` | ✓ | | | | ← superadmin only |
| `audit:read` | ✓ | ✓ | | |

**Superadmin protection rule:** Only a superadmin can assign or remove the `superadmin` role. An admin with `user:manage` cannot elevate any user to superadmin.

JWT tokens encode the user's resolved set of permission names. RBAC middleware checks the token claims — no DB hit per request.

---

## API Endpoints

All routes are prefixed with `/api/v1`.

### Auth (public)
| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create a new account (sends verification email) |
| POST | `/auth/login` | Login with email + password + captcha token → JWT + refresh token |
| POST | `/auth/refresh` | Exchange refresh token for new JWT |
| POST | `/auth/forgot-password` | Send password reset email |
| POST | `/auth/reset-password` | Consume reset token + set new password |
| POST | `/auth/logout` | Revoke refresh token (requires valid JWT) |
| POST | `/auth/verify` | Consume verification or invitation token. Body: `{ token, password? }` |
| POST | `/auth/resend-verification` | Resend verification email for unverified self-registered users. Body: `{ email }` |

### Posts
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/posts` | public | List published posts. Query params: `q`, `tag`, `author`, `page`, `limit`, `sort` |
| GET | `/posts/:slug` | public | Get post by slug |
| POST | `/posts` | `post:create` | Create post |
| PUT | `/posts/:id` | `post:edit` | Update post |
| PATCH | `/posts/:id/publish` | `post:publish` | Toggle publish status |
| DELETE | `/posts/:id` | `post:delete` | Soft delete post (sets `deleted_at`); cascades to its comments |

### Tags
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/tags` | public | List all tags |
| POST | `/tags` | `post:create` | Create tag |
| DELETE | `/tags/:id` | `post:delete` | Hard delete tag (no soft delete) |

### Comments
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/posts/:id/comments` | public | List approved, non-deleted comments for a post |
| POST | `/posts/:id/comments` | `comment:create` | Submit comment (status = pending) |
| PATCH | `/comments/:id/status` | `comment:approve` | Approve or reject a comment |
| DELETE | `/comments/:id` | `comment:approve` | Soft delete comment (sets `deleted_at`) |

### Images
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/images` | `image:upload` | Upload image → S3, returns URL |
| DELETE | `/images/:id` | `image:upload` | Hard delete — removes from S3 and DB |

### Users
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/users` | `user:manage` | List active (non-deleted) users |
| POST | `/users` | `user:create` | Create user account (no password); sends invitation email |
| PUT | `/users/:id` | `user:update` | Update a user's profile/email (superadmin) |
| DELETE | `/users/:id` | `user:delete` | Soft delete a user account (sets `deleted_at`); revokes all refresh tokens |
| POST | `/users/:id/roles` | `user:manage` | Assign role to user |
| DELETE | `/users/:id/roles/:roleId` | `user:manage` | Remove role from user |
| POST | `/users/:id/resend-invitation` | `user:create` | Resend invitation email. Returns `409` if already verified |

### Profile (self)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/me` | any authenticated | Get own profile: `id`, `email`, `display_name`, `bio`, `avatar_url`, `last_login_at`, `roles`, `created_at` |
| PATCH | `/me` | any authenticated | Update own `display_name`, `bio`, `avatar_url` |
| POST | `/me/change-password` | any authenticated | Change password (requires `current_password` + `new_password`; revokes all refresh tokens) |

### Audit Logs
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/audit-logs` | `audit:read` | List audit log entries. Query params: `actor_id`, `action`, `resource_type`, `resource_id`, `from`, `to`, `page`, `limit` |
| GET | `/audit-logs/:id` | `audit:read` | Get a single audit log entry |

---

## Key Implementation Details

### Authentication
- Passwords hashed with **bcrypt** (cost 12)
- **JWT** (access token, short-lived e.g. 15 min) + **refresh token** (long-lived, stored in DB, rotated on use)
- Login requires a valid **CAPTCHA token** (hCaptcha or Google reCAPTCHA v2/v3), verified server-side
- Password reset tokens are single-use, hashed before storage, expire after 1 hour
- `last_login_at` on the user record is updated on every successful login
- **Login guard:** If `email_verified_at IS NULL` → return `403` with code `EMAIL_NOT_VERIFIED`

### Email Verification & Invitation

Two flows share the same `UserVerificationToken` table and `POST /auth/verify` endpoint, differentiated by `type`.

**Flow A — Superadmin creates a user (`POST /users`):**
1. User record created: `password_hash = NULL`, `email_verified_at = NULL`
2. `invitation` token generated (raw token → SHA-256 hash stored)
3. Invitation email sent: link to `{FRONTEND_URL}/accept-invitation?token={raw}`
4. On `POST /auth/verify` with `{ token, password }`:
   - Token validated (not expired, not used)
   - `password` field **required** — validated against full complexity rules
   - `password_hash` set, `email_verified_at = NOW()`, token marked `used_at`
   - Fires `user.invitation_accepted` audit event

**Flow B — Self-registration (`POST /auth/register`):**
1. User record created: `password_hash` set, `email_verified_at = NULL`
2. `email_verification` token generated
3. Verification email sent: link to `{FRONTEND_URL}/verify-email?token={raw}`
4. On `POST /auth/verify` with `{ token }`:
   - Token validated
   - `password` field ignored
   - `email_verified_at = NOW()`, token marked `used_at`
   - Fires `user.email_verified` audit event

**Resend flows:**
- `POST /auth/resend-verification` (public): for self-registered unverified users — invalidates previous token, issues and emails a new one
- `POST /users/:id/resend-invitation` (`user:create`): for superadmin-invited unverified users — same invalidation + re-issue; returns `409` if already verified

**Token error codes:**

| Code | HTTP | Meaning |
|---|---|---|
| `TOKEN_INVALID` | 400 | Token not found or already used |
| `TOKEN_EXPIRED` | 400 | Token older than 24 hours |
| `PASSWORD_REQUIRED` | 400 | Invitation token consumed without a `password` field |
| `EMAIL_NOT_VERIFIED` | 403 | Login attempted on unverified account |
| `USER_ALREADY_VERIFIED` | 409 | Resend attempted on already-verified user |

**Frontend URL** for email links is configured via `FRONTEND_URL` env var.

### Password Complexity

**Standard:** NIST SP 800-63B §5.1.1 combined with ISO/IEC 27002:2022 password guidance.

**Rules (applied in order — fail-fast):**

| # | Rule | Requirement | Standard |
|---|---|---|---|
| 1 | Trim | Strip leading/trailing whitespace before any check | NIST 800-63B |
| 2 | Minimum length | ≥ **12 characters** (after trim) | NIST 800-63B + ISO/IEC 27002:2022 |
| 3 | No maximum length | Accept passwords of any length | NIST 800-63B §5.1.1 |
| 4 | Uppercase letter | ≥ 1 character `[A-Z]` | ISO/IEC 27002:2022 |
| 5 | Lowercase letter | ≥ 1 character `[a-z]` | ISO/IEC 27002:2022 |
| 6 | Digit | ≥ 1 character `[0-9]` | ISO/IEC 27002:2022 |
| 7 | Special character | ≥ 1 from `!@#$%^&*()_+-=[]{}';:"\\|,.<>/?` | ISO/IEC 27002:2022 |
| 8 | Unicode allowed | Accept all printable Unicode code points | NIST 800-63B §5.1.1 |
| 9 | No sequential patterns | Reject common sequences: `123456`, `abcdef`, `qwerty`, `password`, etc. | NIST 800-63B §5.1.1.2 |
| 10 | No breach match | Must not appear in HIBP Pwned Passwords database | NIST 800-63B §5.1.1.2 |
| — | No forced rotation | Users are NOT required to change passwords periodically | NIST 800-63B §5.1.1 |

**Enforcement points:** Rules 1–10 are enforced at all three password-setting flows:
- `POST /auth/register`
- `POST /auth/reset-password`
- `POST /me/change-password`

**HIBP Breach Check (k-Anonymity model):**
1. SHA-1 hash the candidate password
2. Send only the **first 5 hex characters** to `https://api.pwnedpasswords.com/range/{first5}`
3. API returns matching hash suffixes + breach counts
4. Check locally whether the full hash suffix appears in the response
5. If found → reject with `PASSWORD_BREACHED` error
6. **Fail open:** if the HIBP API is unavailable (timeout 3s, 5xx, rate-limited) → log a warning and allow the password (availability over over-rejection per NIST guidance)

**Architecture:** All rules live in `internal/pkg/password/`:
```
internal/pkg/password/
├── validator.go       # PasswordValidator struct; Validate(ctx, password) error
├── hibp.go            # HIBP k-anonymity HTTP client
└── validator_test.go  # Table-driven tests covering all rules; HIBP mocked
```
`PasswordValidator` is constructed once in `main.go` and injected into the three usecases that need it.

**Validation error codes** (returned in `400` response body):

| Code | User-facing message |
|---|---|
| `PASSWORD_TOO_SHORT` | "Password must be at least 12 characters" |
| `PASSWORD_NO_UPPERCASE` | "Password must contain at least one uppercase letter" |
| `PASSWORD_NO_LOWERCASE` | "Password must contain at least one lowercase letter" |
| `PASSWORD_NO_DIGIT` | "Password must contain at least one number" |
| `PASSWORD_NO_SPECIAL` | "Password must contain at least one special character" |
| `PASSWORD_SEQUENTIAL` | "Password contains a common pattern and cannot be used" |
| `PASSWORD_BREACHED` | "This password has appeared in a data breach. Please choose a different one" |

### Soft Delete

`deleted_at TIMESTAMPTZ DEFAULT NULL` is added to `posts`, `comments`, and `users`. Tags and Images use hard delete.

**Repository rule:** Every `SELECT` on these three tables appends `AND deleted_at IS NULL`. No deleted records are ever returned to the application layer.

**On delete:** Repositories execute `UPDATE ... SET deleted_at = NOW()` — never `DELETE FROM`.

**Cascades (application-level, same transaction):**
- Soft-deleting a **post** also soft-deletes all its non-deleted comments
- Soft-deleting a **user** revokes all their active refresh tokens; their posts and comments are **not** soft-deleted (content survives)

**Retention purge:** The daily background job permanently removes records where `deleted_at < NOW() - interval(SOFT_DELETE_RETENTION_DAYS days)`. Default: 90 days.

**Partial DB indexes** on `(deleted_at)` for each affected table ensure queries remain fast.

### User Profile & Change Password

**Profile fields** (`display_name`, `bio`, `avatar_url`, `last_login_at`) live on the `users` table.

**`GET /me`** returns: `id`, `email`, `display_name`, `bio`, `avatar_url`, `last_login_at`, `roles[]`, `created_at`. Sensitive fields (`password_hash`) are never serialized.

**`PATCH /me`** allows self-editing of `display_name`, `bio`, `avatar_url`. Email is not self-editable (superadmin only via `PUT /users/:id`). Trim + `StrictPolicy` XSS rules apply.

**`POST /me/change-password`** flow:
1. Verify `current_password` against `password_hash` — `403` on mismatch
2. Validate `new_password` against the **Password Complexity Rules** (see §Password Complexity) — `400` on failure
3. Hash and store `new_password`
4. Revoke all active refresh tokens for the user
5. Return `200` — caller must re-authenticate
6. Fires `user.password_changed` audit event

### Full-Text Search
- PostgreSQL `tsvector` column on `posts`, populated via a DB trigger
- `GIN` index on `search_vector` for fast queries
- `GET /posts?q=keyword` uses `to_tsquery` + `ts_rank` for relevance ordering

### Post Counters

**`comment_count`** — maintained by a PostgreSQL trigger on the `comments` table:
- Increments by 1 when a comment is inserted with `status = 'approved'`
- Increments by 1 when a comment's status transitions **to** `'approved'`
- Decrements by 1 when a comment's status transitions **from** `'approved'` to `'pending'` or `'rejected'`
- Decrements by 1 when an `approved` comment is soft-deleted (`deleted_at` set)

**`view_count`** — incremented atomically in the `GetPostBySlug` usecase after a successful fetch of a `published` post:
```sql
UPDATE posts SET view_count = view_count + 1 WHERE id = $1
```
Draft post previews (admin access) must explicitly skip this increment. View counts are not audited.

Both fields are returned in all post response objects — both the list (`GET /posts`) and detail (`GET /posts/:slug`) endpoints. On the detail endpoint the response reflects the already-incremented count.

### Image Uploads
- Storage backend: **S3-compatible** (Cloudflare R2, MinIO, AWS S3)
- Files are validated (MIME type, max size e.g. 10 MB) before upload
- The DB record stores `s3_key` and `url`; deletion removes both

### Email
- Sent via **SMTP** (provider-agnostic: SendGrid, Resend, Mailgun, self-hosted)
- Transactional emails: password reset link, email verification link, invitation link
- HTML + plain text templates

### CORS
- Configurable allowed origins via env var
- Gin CORS middleware applied globally

### Audit Logging

**Architecture:** An `AuditLogger` interface is defined in `domain`. Usecases receive it via constructor injection and call it after each successful audited operation. The PostgreSQL implementation writes the row synchronously. A nil/no-op implementation is used in unit tests.

```go
// domain interface
type AuditLogger interface {
    Log(ctx context.Context, entry AuditEntry) error
}
```

**Audited events:**

| Event Code | Trigger |
|---|---|
| `user.registered` | New account created via register endpoint (email verification pending) |
| `user.email_verified` | Self-registered user verified their email |
| `user.invitation_accepted` | Superadmin-invited user accepted invitation and set password |
| `user.login` | Successful login |
| `user.login_failed` | Failed login (wrong password or captcha rejection) |
| `user.logout` | Refresh token revoked via logout |
| `user.password_reset_requested` | Forgot-password email sent |
| `user.password_reset_completed` | Password successfully reset |
| `user.password_changed` | Authenticated user changed their own password (all sessions revoked) |
| `user.created` | Superadmin creates a user account directly |
| `user.updated` | Superadmin updates a user profile |
| `user.deleted` | Superadmin deletes a user |
| `user.role_assigned` | Role assigned to a user |
| `user.role_removed` | Role removed from a user |
| `post.created` | Post created |
| `post.updated` | Post updated |
| `post.published` | Post published |
| `post.unpublished` | Post set back to draft |
| `post.deleted` | Post deleted |
| `comment.created` | Comment submitted |
| `comment.approved` | Comment approved |
| `comment.rejected` | Comment rejected |
| `comment.deleted` | Comment deleted |
| `tag.created` | Tag created |
| `tag.deleted` | Tag deleted |
| `image.uploaded` | Image uploaded to S3 |
| `image.deleted` | Image deleted from S3 + DB |

**Special case — `user.login_failed`:** `actor_id` is `null`; `actor_email` holds the attempted email; `ip_address` is always captured.

**Retention:** A background goroutine starts at server boot and runs once daily. It deletes rows where `created_at < NOW() - interval(AUDIT_LOG_RETENTION_DAYS days)`. Default: 365 days.

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
FRONTEND_URL            # Base URL for email verification/invitation links

AUDIT_LOG_RETENTION_DAYS  # Default: 365
SOFT_DELETE_RETENTION_DAYS  # Default: 90
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
| `github.com/securego/gosec/v2` | SAST — CI tool only, not imported in application code |
