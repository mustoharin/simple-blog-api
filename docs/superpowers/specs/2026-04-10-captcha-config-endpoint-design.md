# Captcha Config Endpoint Design

**Date:** 2026-04-10  
**Status:** Approved

## Problem

The frontend login page needs to render a CAPTCHA widget before calling `POST /auth/login`. To render the widget it needs three things:

1. Whether CAPTCHA is enabled at all (so it can show or hide the widget)
2. Which provider is in use (`hcaptcha` or `recaptcha`) to load the correct JS library
3. The public site key to initialise the widget

Currently none of this is exposed by the API. The server only holds `CAPTCHA_SECRET` (private) and `CAPTCHA_PROVIDER`, and there is no `CAPTCHA_SITE_KEY` config at all.

## Solution

Add `GET /api/v1/auth/captcha-config` — a public, unauthenticated endpoint that returns the captcha configuration the frontend needs to render the widget.

## Config Changes

Add one new environment variable to `config/config.go`:

| Variable | Default | Description |
|---|---|---|
| `CAPTCHA_SITE_KEY` | `""` | Public site key from the CAPTCHA provider dashboard |

`enabled` is derived from `CAPTCHA_SECRET != ""` (consistent with existing noop-verifier logic).

## API

### `GET /api/v1/auth/captcha-config`

**Auth:** None (public)

**Response when enabled (`CAPTCHA_SECRET` is set):**
```json
{
  "enabled": true,
  "provider": "hcaptcha",
  "site_key": "<public-site-key>"
}
```

**Response when disabled (`CAPTCHA_SECRET` is empty):**
```json
{
  "enabled": false
}
```

`provider` is always one of `"hcaptcha"` or `"recaptcha"` (values of `CAPTCHA_PROVIDER`).

## Implementation

### Files Changed

| File | Change |
|---|---|
| `config/config.go` | Add `CaptchaSiteKey string` field, read from `CAPTCHA_SITE_KEY` env var |
| `internal/delivery/http/handler/auth.go` | Add `captchaSiteKey`, `captchaProvider`, `captchaEnabled` fields to `AuthHandler`; add `GetCaptchaConfig` handler method |
| `cmd/api/main.go` | Pass `cfg.CaptchaSiteKey`, `cfg.CaptchaProvider`, and `cfg.CaptchaSecret != ""` to `AuthHandler` constructor |
| `internal/delivery/http/server.go` | Register `GET /auth/captcha-config` route (no middleware) |

### Handler Logic

`GetCaptchaConfig` is a pure config read — no database, no use-case layer:

```
if !captchaEnabled:
    return { "enabled": false }
else:
    return { "enabled": true, "provider": captchaProvider, "site_key": captchaSiteKey }
```

No new use-case or repository is needed.

## Testing

- Unit test on `AuthHandler.GetCaptchaConfig`:
  - When enabled: asserts `enabled=true`, correct provider and site key in response
  - When disabled: asserts `enabled=false`, no `provider` or `site_key` fields in response
- Config test: `TestConfig_CaptchaSiteKey_DefaultsToEmpty`
