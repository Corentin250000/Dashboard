# Dashboard Backend

The backend is the HTTP API for the Dashboard application. It is a Go 1.25
service built with Fiber, PostgreSQL, GORM, and a small set of integrations for
email confirmation and SNCF departure data.

This document describes the backend as it exists today. The database already
contains models for future service subscriptions and widget instances, but
their CRUD API and timer worker are not implemented yet.

## Responsibilities

The backend currently provides:

- A service and widget catalog at `GET /about.json`.
- A health endpoint at `GET /health`.
- Local account registration with bcrypt password hashing.
- Email confirmation and confirmation-token renewal.
- Login with an HTTP-only `session_id` cookie.
- A SNCF departure widget backed by the SNCF Navitia API.
- Automatic PostgreSQL connection retries at startup.
- GORM model migration at startup.
- Swagger UI and generated OpenAPI files.
- CORS, request IDs, request logging, panic recovery, and rate limiting for
  confirmation-token renewal.

The following requirements are represented in the data model but are
not exposed through backend routes yet:

- Service subscription management.
- OAuth account linking.
- Widget-instance CRUD operations.
- Per-widget refresh timers and cached widget data.
- Administration features.

## Technology Stack

| Component | Technology | Purpose |
| --- | --- | --- |
| Language | Go 1.25 | Backend implementation |
| HTTP server | Fiber v3 | Routing, middleware, and JSON responses |
| Validation | `go-playground/validator` | Request validation |
| Database | PostgreSQL 16 | Users, services, widgets, and SNCF stations |
| ORM | GORM | Database access and development migrations |
| Authentication | bcrypt + secure random tokens | Password hashing and email confirmation |
| Email | Mailjet | Registration confirmation messages |
| External data | SNCF Navitia API | Station catalog and train departures |
| API documentation | Swaggo + Fiber Swagger | OpenAPI specification and Swagger UI |
| Runtime | Docker / Docker Compose | Reproducible local and deployment environment |
| Secrets | Doppler or environment variables | Runtime configuration |

## Architecture

```text
                         +----------------------+
                         | React frontend       |
                         | :8081                |
                         +----------+-----------+
                                    |
                                    | HTTP / JSON
                                    v
+------------------+       +--------+---------+       +----------------------+
| PostgreSQL       |<----->| Go + Fiber API   |------>| Mailjet              |
| db:5432          |       | :8080            |       | confirmation emails  |
+------------------+       +--------+---------+       +----------------------+
                                     |
                                     | HTTPS API calls
                                     v
                            +----------------------+
                            | SNCF Navitia API     |
                            +----------------------+
```

The server starts only after it has connected to PostgreSQL. It then migrates
the GORM models, checks whether the local SNCF station catalog is populated,
and starts the Fiber HTTP server. If the station catalog is empty, a background
goroutine imports stations from SNCF while the HTTP server continues starting.

## Requirements

Choose one of the following workflows:

- Docker workflow: Docker Engine and Docker Compose.
- Native workflow: Go 1.25 and a reachable PostgreSQL instance.

The Docker workflow is the recommended path because it starts the API and the
database with the correct network names and ports.

## Run With Docker Compose

From the repository root:

```bash
cp .env.example .env
docker-compose build
docker-compose up
```

The backend is available at:

- API: `http://127.0.0.1:8080`
- Health check: `http://127.0.0.1:8080/health`
- Service catalog: `http://127.0.0.1:8080/about.json`
- Swagger UI: `http://127.0.0.1:8080/docs/swagger/`

The Compose service maps host port `8080` to container port `8080`.  
PostgreSQL is intentionally available only inside the Compose network at `db:5432`.

To stop the application without deleting database data:

```bash
docker-compose down
```

To stop the application and delete the PostgreSQL volume:

```bash
docker-compose down -v
```

The second command permanently deletes local database data.

## Run Natively

Start PostgreSQL and create a database, then configure the database variables
described below. From the backend directory:

```bash
go mod download
go vet ./...
go test ./...
go run ./cmd/server
```

By default, the server listens on port `8080` and connects to PostgreSQL at
`db:5432`. For a native process, set `DB_HOST=127.0.0.1` unless a `db` host is
available through local DNS or `/etc/hosts`.

## Configuration

Environment variables are read at runtime. `utils.GetEnv` uses the configured
value when it is non-empty and otherwise applies the documented fallback.

### Server Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `PORT` | No | `8080` | HTTP listen port. The Docker/subject configuration must remain `8080`. |
| `FRONTEND_URL` | No | `http://127.0.0.1:8081` | Frontend origin used for CORS and confirmation links. Multiple origins may be comma-separated. |
| `JWT_SECRET` | No | None | Reserved configuration value. The current login implementation uses a random cookie token rather than JWT. |
| `DOPPLER_TOKEN` | No | Empty | Enables Doppler secret injection in the container entrypoint. |

### Database Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DB_HOST` | No | `db` | PostgreSQL hostname. |
| `DB_PORT` | No | `5432` | PostgreSQL port. |
| `POSTGRES_DB` | Yes for a real database | None | Database name. |
| `POSTGRES_USER` | Yes for a real database | None | Database user. |
| `POSTGRES_PASSWORD` | Yes for a real database | None | Database password. |

### External API and Email Variables

| Variable | Required | Description |
| --- | --- | --- |
| `API_KEY_SNCF` | SNCF features only | SNCF Navitia API key. It is sent as the HTTP Basic Auth username with an empty password. |
| `API_KEY_MAILJET` | Email confirmation only | Mailjet API key. |
| `SECRET_KEY_MAILJET` | Email confirmation only | Mailjet secret key. |
| `MAIL_SENDER_ACCOUNT` | Email confirmation only | Validated sender address used by Mailjet. |

The application reads `API_KEY_MAILJET` and `SECRET_KEY_MAILJET`. Keep those
names aligned with the backend even if an external secret-management setup uses
different labels. Never commit real values to `.env`, source code, or logs.

### Doppler

The backend image contains the Doppler CLI. Its entrypoint behaves as follows:

- With a non-empty `DOPPLER_TOKEN`, it runs `doppler run -- /server`.
- With an empty `DOPPLER_TOKEN`, it removes the Doppler wrapper and starts
  `/server` directly.

See [`../docs/doppler.md`](../docs/doppler.md) for the repository-wide secret
management procedure.

## HTTP API

All JSON endpoints use UTF-8 JSON bodies. Error responses generally have this
shape:

```json
{
  "status": "error",
  "message": "Human-readable explanation"
}
```

### `GET /`

Returns a pointer to the API catalog.

```json
{
  "message": "dashboard API - see GET /about.json"
}
```

### `GET /health`

Returns `200 OK` when the HTTP process is responding.

```json
{
  "status": "ok"
}
```

This endpoint does not perform a database or external API health check.

### `GET /about.json`

Returns the catalog required by the project subject. `server.current_time` is
the Unix epoch timestamp in seconds. `client.host` is resolved in this order:

1. First address in `X-Forwarded-For`.
2. `X-Real-IP`.
3. Fiber's remote client address.

Current response example:

```json
{
  "client": {
    "host": "127.0.0.1"
  },
  "server": {
    "current_time": 1790754537,
    "services": [
      {
        "name": "sncf",
        "widgets": [
          {
            "name": "departures",
            "description": "Display train departures for a station",
            "params": [
              {
                "name": "station",
                "type": "string"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

Widget parameter types are limited by the contract to `string` and `integer`.
The catalog is defined in `internal/about/about.go`; add new services and
widgets there so the endpoint remains the single source of truth.

### `POST /auth/register`

Creates a local account and starts email confirmation delivery.

Request:

```json
{
  "email": "alice@example.com",
  "password": "correct horse battery staple"
}
```

Validation rules:

- `email` is required and must have a valid email format.
- `password` is required and must contain at least 8 characters.

Success: `201 Created`

```json
{
  "message": "Account created, please check your email"
}
```

Important behavior:

- Passwords are stored as bcrypt hashes, never as plaintext.
- New users are unconfirmed for 24 hours.
- The confirmation email is sent asynchronously after the database insert.
- If Mailjet is not configured, the account is still created and the reason is
  logged by the backend.

Common errors:

| Status | Meaning |
| --- | --- |
| `400` | Invalid JSON or invalid fields |
| `409` | Email already exists |
| `500` | Account or password-hash creation failed |
| `503` | Database has not been configured |

### `POST /auth/login`

Authenticates a confirmed local account.

Request:

```json
{
  "email": "alice@example.com",
  "password": "correct horse battery staple"
}
```

Success: `200 OK`

```json
{
  "message": "Logged in"
}
```

The response sets an HTTP-only `session_id` cookie with a seven-day expiry.
`Secure` is enabled when the request uses HTTPS and `SameSite` is set to `Lax`.

Common errors:

| Status | Meaning |
| --- | --- |
| `400` | Invalid JSON or invalid fields |
| `401` | Invalid credentials or OAuth-only account |
| `403` | Account email has not been confirmed |
| `500` | Session token generation failed |
| `503` | Database has not been configured |

### `GET /confirm?token=<token>`

Consumes an email confirmation token and marks the account as confirmed.

Possible responses:

| Status | JSON status | Meaning |
| --- | --- | --- |
| `200` | `confirmed` | Account successfully confirmed |
| `200` | `already_confirmed` | Account was already confirmed |
| `400` | `error` | Query parameter is missing |
| `404` | `invalid` | Token is unknown or already consumed |
| `410` | `expired` | Token is older than 24 hours |
| `500` | `error` | Database error |
| `503` | `error` | Database has not been configured |

Example successful response:

```json
{
  "status": "confirmed"
}
```

### `POST /confirm/resend`

Rotates an unconfirmed account's token and sends a new confirmation email.
This route is protected by a Fiber limiter: at most three requests are accepted
per 15-minute window for the limiter's configured client key.

Request:

```json
{
  "token": "current-confirmation-token"
}
```

Possible responses include:

- `200` with `{ "status": "ok" }`.
- `200` with `{ "status": "already_confirmed" }`.
- `400` when the token is missing.
- `404` when the token is unknown.
- `500` when token generation or persistence fails.
- `503` when the database is unavailable.

### `POST /api/sncf/departures`

Returns the next departures for a station known by the local SNCF station
catalog.

Request:

```json
{
  "text": "Paris Gare du Nord"
}
```

The `text` value is trimmed and must be non-empty. The backend resolves the
station name to a Navitia `stop_area` ID, then calls:

```text
https://api.sncf.com/v1/coverage/sncf/stop_areas/{station-id}/departures
```

The upstream request asks for up to ten departures. The response is normalized
to the following shape:

```json
{
  "departures": [
    {
      "scheduled_departure": "14:35",
      "departure": "14:40",
      "destination": "Lille Flandres",
      "delay_minutes": 5,
      "cancelled": false
    }
  ]
}
```

Departure times are returned as `HH:MM`. Delay and cancellation information is
derived from SNCF disruption data when it is available.

Possible errors:

| Status | Meaning |
| --- | --- |
| `400` | Invalid body or empty `text` |
| `404` | Station is not present in `sncf_stations` |
| `500` | Local request creation or database failure |
| `502` | SNCF could not be reached, returned an error, or returned invalid JSON |
| `503` | Database or `API_KEY_SNCF` is not configured |

The handler uses a 30-second timeout for the upstream SNCF request and passes
the incoming request context to it, so cancelled client requests can cancel the
upstream call.

### Swagger UI

Open `http://127.0.0.1:8080/docs/swagger/` after starting the server. The
generated specification is stored in `backend/docs` and is loaded by the
Swagger middleware from `/docs/swagger/doc.json`.

When changing Swagger annotations, regenerate the files from `backend/`:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
"$(go env GOPATH)/bin/swag" init -g cmd/server/main.go -o docs
```

## Database

### Startup behavior

The server performs these database steps before serving requests:

1. Connect to PostgreSQL, retrying up to ten times with a two-second delay.
2. Run `models.AutoMigrate` for the application models.
3. Count rows in `sncf_stations`.
4. Start a background station import when the table is empty.

The SQL files in `../db/migrations` are the reviewed schema source of truth.
GORM `AutoMigrate` is a development convenience and should not replace
versioned SQL migrations for production schema changes.

### Main tables

| Table | Purpose |
| --- | --- |
| `users` | Local accounts, password hashes, and confirmation state |
| `services` | Static service catalog and authentication metadata |
| `user_services` | User subscriptions to services |
| `widgets` | Widget catalog entries belonging to services |
| `widget_params` | Widget configuration parameters; only `integer` and `string` are allowed |
| `widget_instances` | User-configured widget placements, refresh rates, and cached data |
| `sncf_stations` | SNCF/Navitia station ID and display-name mapping |

Sensitive service credentials and OAuth tokens are represented by `*_enc`
columns in `user_services`. The intended design is application-side encryption
before insertion; PostgreSQL must not receive plaintext credentials.

### SNCF station import

On first startup with an empty station table, the backend imports the Navitia
`stop_areas` catalog page by page, using a page size of 1,000. It stops when the
upstream response is empty or the reported result count has been reached.

The import requires `API_KEY_SNCF`. If the key is missing, the import logs an
error and the server still starts, but the SNCF departure endpoint cannot resolve
stations until the catalog is populated.

## Middleware and Runtime Behavior

The Fiber application installs middleware in this order:

1. Request ID generation.
2. Structured request logging containing time, request ID, status, method, path,
   and latency.
3. Panic recovery.
4. CORS with credentials enabled.

Allowed origins include the configured `FRONTEND_URL` plus the local development
origins on ports `5173`, `8080`, and `8081`. Wildcard origins are deliberately
not used because credentialed browser requests cannot use `*`.

The server uses these timeouts:

- Read timeout: 5 seconds.
- Write timeout: 10 seconds.
- Idle timeout: 60 seconds.
- Graceful shutdown timeout: 10 seconds.

Only `GET` and `HEAD` are registered for `/`, `/health`, and `/about.json`. An
unmatched route returns `404` with:

```json
{
  "error": "not found"
}
```

## Project Layout

```text
repo/
├── backend/                    # Go API
│   ├── cmd/server/main.go      # Application bootstrap and route registration
│   ├── internal/               # API handlers and backend internals
│   │   ├── about/              # /about.json response and service registry
│   │   ├── auth/               # Registration, login, confirmation, email flow
│   │   ├── mail/               # Confirmation email template rendering
│   │   ├── models/             # GORM models and AutoMigrate helper
│   │   ├── theme/              # Email theme colors
│   │   ├── utils/              # Environment and header helpers
│   │   └── widgets/            # Widgets endpoints
│   ├── docs/                   # Generated Swagger/OpenAPI files
│   ├── Dockerfile              # Multi-stage image with Doppler CLI
│   ├── entrypoint.sh           # Doppler/no-Doppler startup wrapper
│   ├── go.mod                  # Go module and direct dependencies
│   └── go.sum                  # Dependency checksums
├── frontend/                  # React + Vite application
│   └── Dockerfile              # Vite preview server on port 8081
├── db/                        # PostgreSQL image and Doppler wrapper
│   ├── migrations/             # SQL schema reference files
│   ├── Dockerfile              # PostgreSQL 16 image
│   └── entrypoint.sh           # Doppler/no-Doppler database wrapper
├── docker-compose.yml          # server (:8080), db (internal), client (:8081)
├── .env.example                # Doppler and classic environment templates
└── docs/doppler.md             # Doppler secrets-management documentation
```

## Development Guidelines

### Add a service or widget to `/about.json`

Edit `internal/about/about.go` and add a `Service` entry to `Registry()`.
Every widget must have at least one configurable parameter to satisfy the
project subject. Parameter types must be exactly `string` or `integer`.

Example:

```go
{
    Name: "weather",
    Widgets: []Widget{
        {
            Name:        "city_temperature",
            Description: "Display the temperature for a city",
            Params: []Param{
                {Name: "city", Type: "string"},
            },
        },
    },
}
```

### Add an endpoint

1. Define request and response types close to the handler.
2. Validate request input before database or external API calls.
3. Return consistent JSON errors with an appropriate HTTP status.
4. Add Swagger annotations to the handler.
5. Register the route in `cmd/server/main.go`.
6. Regenerate `backend/docs`.
7. Run formatting, static checks, and tests.

### Recommended verification commands

From `backend/`:

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
```

From the repository root:

```bash
docker-compose build
docker-compose up -d
curl -fsS http://127.0.0.1:8080/health
curl -fsS http://127.0.0.1:8080/about.json
docker-compose down
```

## Security Notes

- Do not commit `.env`, service tokens, API keys, database passwords, or OAuth
  secrets.
- Passwords are hashed with bcrypt before storage.
- Confirmation tokens are generated with 256 bits of cryptographic randomness,
  expire after 24 hours, and are cleared after successful confirmation.
- The login cookie is HTTP-only and uses `SameSite=Lax`.
- CORS uses explicit origins and allows credentials only for those origins.
- External credentials should be encrypted in the application before being
  stored in `*_enc` database columns.
- Do not log passwords or API keys. The current email helper logs confirmation
  tokens when credentials are missing; this should be removed or replaced with
  a safer operational signal before production use.

### Current authentication limitation

`createSession` currently generates a secure random token but does not persist a
session record or validate the cookie on protected routes. It is therefore a
temporary authentication shell, not a complete server-side session system.
Before exposing authenticated dashboard data, implement a session store with
expiry, revocation, user lookup, and authentication middleware.

## Subject Compliance

The project subject requires:

- A root `docker-compose.yml`.
- A server listening on port `8080`.
- `GET /about.json` with `client.host`, `server.current_time`, and a service
  catalog.
- Widget parameters whose types are `integer` or `string`.
- Configurable widgets rather than parameterless widgets.

The current backend satisfies the HTTP catalog contract with the `sncf`
service and its configurable `departures` widget. The broader dashboard
workflow described by the subject still requires service-subscription,
widget-instance, timer, and authenticated data routes to be completed.

## Troubleshooting

### The server cannot connect to PostgreSQL

- With Docker Compose, use `DB_HOST=db`, not `127.0.0.1`.
- Confirm that the `db` container is healthy.
- Verify `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD`.
- If credentials changed after the first startup, recreate the volume with
  `docker-compose down -v` and start again.

### `/api/sncf/departures` returns `503`

Check that `API_KEY_SNCF` is available inside the server container and that the
station import completed successfully. The key must be present when the station
catalog is initially loaded and when departures are requested.

### Registration succeeds but no email arrives

Check `API_KEY_MAILJET`, `SECRET_KEY_MAILJET`, and `MAIL_SENDER_ACCOUNT`. The
sender address must be validated in Mailjet. Email delivery runs in a goroutine,
so the HTTP response does not wait for Mailjet.

### Browser requests fail with a CORS error

Set `FRONTEND_URL` to the exact frontend origin, including scheme and port, for
example `http://127.0.0.1:8081`. Do not add a trailing slash; it is normalized by
the backend, but matching the browser origin exactly makes configuration easier
to inspect.
