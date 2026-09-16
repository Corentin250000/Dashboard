# Database schema

PostgreSQL 16. Source of truth: `db/migrations/001_init_schema.sql`
(apply with `psql -f`). The GORM structs in `backend/internal/models/`
mirror these tables; keep both in sync when the schema evolves.

PKs are UUIDs (`pgcrypto.gen_random_uuid()`). Timestamps are `timestamptz`;
`updated_at` is maintained by trigger. Columns ending in `_enc` hold
**ciphertext encrypted by the Go backend before insert** — Postgres never sees
plaintext and never decrypts (no column-level crypto, `pgcrypto` is only used
for UUID generation).

## Tables

| Table | Role |
|---|---|
| `users` | Platform accounts. `password_hash` is NULL for OAuth-only users. Login is blocked while `is_confirmed` is false; `confirmation_token` is single-use (NULL once consumed) with `confirmation_expires_at`. |
| `services` | Static catalog of available services (weather, github, steam, rss…). `name` is the technical identifier. `auth_type` decides how users subscribe (see below); the `oauth_*` columns hold the app-level OAuth client config, one row per service. |
| `user_services` | One row per (user, service) subscription (`UNIQUE`). Holds per-user credentials: cleartext username + encrypted password for `credentials` services, encrypted tokens + expiry for `oauth2` services. |
| `widgets` | Per-service widget catalog. `UNIQUE (service_id, name)`. |
| `widget_params` | Configurable parameters of a catalog widget. `type` is restricted to `integer`/`string` to match the `/about.json` contract. `UNIQUE (widget_id, name)`. |
| `widget_instances` | Configured widgets placed on a user's dashboard. `config` stores this instance's param values, `cached_data` the last fetch result, `refresh_rate_seconds` (> 0) drives the Timer, `position_x/y` the layout, `last_refreshed_at` feeds the Timer query (indexed). |

## Relations

```
users 1───* user_services *───1 services 1───* widgets 1───* widget_params
  │                                │
  │                                └── (oauth_* app config lives here)
  │
  └───* widget_instances *───1 widgets
```

All FKs are `ON DELETE CASCADE`: deleting a user removes their subscriptions
and dashboard; deleting a service removes its subscriptions, catalog widgets,
params, and (transitively) affected instances. Indexes exist on every FK plus
`widget_instances(last_refreshed_at)`.

## `auth_type` values and their impact on `user_services`

| `services.auth_type` | Meaning | Columns filled in `user_services` |
|---|---|---|
| `none` | No account needed | Only `user_id`, `service_id` (+ `subscribed_at`) |
| `api_key` | User pastes a personal key | Key stored encrypted (same pattern as tokens below; exact column chosen at implementation time) |
| `credentials` | Username + password login | `external_username` (cleartext) + `external_password_enc` (encrypted) |
| `oauth2` | OAuth 2.0 dance via Goth | `oauth_access_token_enc` + `oauth_refresh_token_enc` (encrypted) + `oauth_token_expires_at`; app endpoints come from the service row's `oauth_*` columns |

The OAuth flow must **link to the existing platform user**: the callback attaches
tokens to the logged-in user's `user_services` row (or to the user resolved from
the verified provider email), never creates a detached account.
