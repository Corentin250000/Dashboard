# Database schema

PostgreSQL 16. The SQL files in `db/migrations/` describe the reviewed schema
and seed history. The current server startup path applies GORM `AutoMigrate`
and then seeds widget sizes and the application catalog; it does not execute
the SQL files through a separate migration runner. The GORM structs in
`backend/internal/models/` must remain synchronized with the SQL schema.

PKs are UUIDs (`pgcrypto.gen_random_uuid()`). Timestamps are `timestamptz`;
`updated_at` is maintained by trigger. Columns ending in `_enc` hold
**ciphertext encrypted by the Go backend before insert** — Postgres never sees
plaintext and never decrypts (no column-level crypto, `pgcrypto` is only used
for UUID generation).

## Tables

| Table | Role |
|---|---|
| `users` | Platform accounts. `password_hash` is NULL for OAuth-only users. Login is blocked while `is_confirmed` is false; `confirmation_token` is single-use (NULL once consumed) with `confirmation_expires_at`. |
| `sessions` | Persistent HTTP-only browser sessions. The UUID is stored in the `session_id` cookie and links to one user until `expires_at`. |
| `services` | Static catalog of available services (SNCF, Spotify, Google Calendar, Steam, and Chess.com). `name` is the technical identifier. `auth_type` decides how users subscribe (see below); the `oauth_*` columns hold the app-level OAuth client config, one row per service. |
| `user_services` | One row per (user, service) subscription (`UNIQUE`). Holds per-user credentials: cleartext username + encrypted password for `credentials` services, encrypted tokens + expiry for `oauth2` services. |
| `widgets` | Per-service widget catalog. `UNIQUE (service_id, name)`. |
| `widget_sizes` | Fixed dashboard sizes: `small` (1x1), `wide` (2x1), `tall` (1x2), and `big` (2x2) grid cells. |
| `widget_allowed_sizes` | Association table defining which fixed sizes each catalog widget supports. |
| `widget_params` | Configurable parameters of a catalog widget. `type` is restricted to `integer`/`string` to match the `/about.json` contract. `UNIQUE (widget_id, name)`. |
| `widget_instances` | Configured widgets placed on a user's desktop dashboard. `config` stores this instance's param values, `size_code` stores the selected allowed size, `cached_data` the last fetch result, `refresh_rate_seconds` (> 0) drives the Timer, `position_x/y` are zero-based coordinates on the 6x6 grid, and `last_refreshed_at` feeds the Timer query (indexed). |
| `steam_friends` | Cached public Steam friend metadata synchronized for a connected Steam account. |

## Relations

```
users 1───* user_services *───1 services 1───* widgets 1───* widget_params
  │                                │
  │                                └── (oauth_* app config lives here)
  │
  └───* widget_instances *───1 widgets
                                      │
                                      └──* widget_allowed_sizes *───1 widget_sizes
```

All FKs are `ON DELETE CASCADE`: deleting a user removes their subscriptions
and dashboard; deleting a service removes its subscriptions, catalog widgets,
params, and (transitively) affected instances. Indexes exist on every FK plus
`widget_instances(last_refreshed_at)`.

## Current catalog

The server currently seeds these services and widgets:

| Service | Widgets |
|---|---|
| `sncf` | `departures`, `journeys` |
| `spotify` | `currently_playing`, `top_items` |
| `calendar` | `upcoming_events` |
| `steam` | `friends` |
| `chess` | `stats`, `monthly_stats` |

The public `/about.json` catalog is defined separately in
`backend/internal/about/about.go`. The database catalog seed is maintained in
`backend/internal/widgets/widgets_instances.go`; changes to either catalog must
be reviewed together.

## Migration and initialization history

The repository contains the following SQL history:

| Migration | Purpose |
|---|---|
| `001_init_schema.sql` | Core users, sessions, services, widgets, and instances. |
| `002_create_sncf_stations.sql` | SNCF station catalog. |
| `003_create_sessions.sql` | Persistent browser sessions. |
| `004_add_widget_sizes.sql` | Widget size definitions and relations. |
| `005_validate_desktop_widget_positions.sql` | Desktop layout constraints. |
| `006_seed_oauth_services.sql` | OAuth service catalog seed. |
| `007_create_steam_friends.sql` | Steam friend synchronization table. |
| `008_seed_google_service.sql` | Google/Calendar service seed. |
| `009_seed_spotify_widgets.sql` | Spotify widget seed. |
| `010_seed_steam_friends_widget.sql` | Steam friends widget seed. |
| `011_add_steam_online_widget.sql` | Steam online-widget updates. |

For a normal application start, use the server initialization described in
`SETUP.md`. Recreating the Compose volume with `docker-compose down -v` removes
all data and causes initialization to run again.

## Widget layout rules

The desktop dashboard is a 6x6 grid. Widget positions are zero-based. The
backend rejects positions outside the grid, overlapping instances, unsupported
sizes, and refresh intervals outside 60–21,600 seconds. Responsive positions
are not stored separately in this database.

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
