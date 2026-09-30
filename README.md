# Dashboard

> **Status: foundation in place, features under construction.** The Docker setup,
> API contract, local account authentication, Swagger reference and secrets
> management are working. Services and widgets are being implemented next.

## What is this?

Dashboard is a **customizable real-time dashboard**: once registered, a user
subscribes to services (weather, RSS, GitHub, …), places configured **widgets**
on a personal page, and each widget refreshes on its own timer — like a
browser new-tab page turned into a control panel.

**Typical uses:** morning overview (weather + agenda + news), dev monitoring
(GitHub activity, service uptime), market watch (exchange rates, prices).

**Current state:** the running app exposes the API contract, local account
authentication and a frontend authentication shell. Services and widgets are
not usable yet — see *Features* below for what is live vs planned.

## Features, services and widgets

| Area | State |
|---|---|
| `GET /about.json` (service/widget catalog) | ✅ Live |
| `GET /health`, `GET /` | ✅ Live |
| Frontend authentication shell (React, served on port 8081) | ✅ Live |
| User registration / login | ✅ Live, email confirmation required |
| Admin section | 🔲 Planned |
| Services & widgets (weather, RSS, GitHub, …) | 🔲 Planned — declared in `backend/internal/about/about.go` (`Registry()`), which is the single source of truth feeding `/about.json` |
| Per-widget timer refresh | 🔲 Planned |

## Run and use the app

Prerequisites: **Docker** + **Docker Compose** only. No Doppler account needed.

```bash
cp .env.example .env   # then optionally fill secrets (see below)
docker-compose up --build
```

Then open:

- Frontend: http://127.0.0.1:8081
- API catalog: http://127.0.0.1:8080/about.json (should show `client.host`, `server.current_time`, `server.services`)
- API reference: http://127.0.0.1:8080/docs/swagger/

Stop with `Ctrl+C` (or `docker-compose down`). To reset the database as well:
`docker-compose down -v` (⚠️ deletes all data).

**Secrets, in one line:** with a Doppler account, put a `DOPPLER_TOKEN` in
`.env` and secrets are injected automatically; without one, fill the classic
variables in `.env` instead. Same launch command either way. Details:
[`docs/doppler.md`](docs/doppler.md).

### Mailjet email configuration

Create a Mailjet account, validate one sender email through **Account > Sender addresses**
(a custom domain and DNS verification are not required), then get the API key and secret
key from **Account Settings > API Key Management**. Configure `API_KEY_MAILJET`,
`SECRET_KEY_MAILJET` and `MAIL_SENDER_ACCOUNT` in Doppler or `.env`.

## Install and build

```bash
git clone https://github.com/EpitechPGE3-2026/G-WEB-500-MLH-5-1-dashboard-8.git && cd G-WEB-500-MLH-5-1-dashboard-8
cp .env.example .env
docker-compose build
docker-compose up
```

Backend only (without Docker, Go ≥ 1.25):

```bash
cd backend && go vet ./... && go run ./cmd/server
```

### Regenerate API documentation

After changing Swagger annotations, regenerate the OpenAPI files from the backend directory:

```bash
cd backend; \
go install github.com/swaggo/swag/cmd/swag@v1.16.6; \
"$(go env GOPATH)/bin/swag" init -g cmd/server/main.go -o docs; \
cd -
```

The generated files in `backend/docs` are served by the Swagger UI at `/docs/swagger/`.

Frontend only (Node ≥ 22):

```bash
cd frontend && npm ci && npm run dev   # dev server with HMR
```

## Technologies and structure

| Part | Technology | Notes |
|---|---|---|
| Backend | **Go + Fiber** | Graceful shutdown, Swagger UI, CORS for local development |
| Frontend | **React 19 + Vite 8 + TypeScript** | Served in prod via `vite preview`; `VITE_API_URL` baked at build time |
| Database | **PostgreSQL 16** | Reachable only inside the Compose network (`db:5432`) |
| Secrets | **Doppler** | `doppler run` wrappers; see [`docs/doppler.md`](docs/doppler.md) |
| Build/run | **Docker Compose** | Single entrypoint: `docker-compose up` |

```
repo/
├── backend/               Go API (cmd/server, internal/about)
│   ├── Dockerfile         multi-stage build + Doppler CLI, CMD via doppler run
│   ├── docs/               generated Swagger/OpenAPI files
│   └── internal/    API handlers, including authentication
│   └── entrypoint.sh      starts with Doppler if DOPPLER_TOKEN is set, else directly
├── frontend/              React+Vite app (Dockerfile: vite preview on :8081)
├── db/                    Postgres image + Doppler wrapper entrypoint
├── docker-compose.yml     services: server (:8080), db (internal), client (:8081)
├── .env.example           all variables, Doppler (Method A) and classic (Method B)
└── docs/doppler.md        secrets management (English)
```

### Key flows (the two parts worth understanding)

**Startup with Doppler** (same for `server` and `db`):

```
docker-compose up
  └─ container starts → entrypoint.sh
       ├─ DOPPLER_TOKEN set   → doppler run → secrets from Doppler → app
       └─ DOPPLER_TOKEN empty → app starts directly (classic .env / fallbacks)
```

**Service catalog** (the subject's contract):

```
GET /about.json → backend/internal/about/about.go :: Registry()
                     (add future services/widgets here; each widget needs ≥1 param
                      of type "string" or "integer")
```

## Contributing

- Never commit `.env` or `doppler.yaml` (gitignored). Only `.env.example` documents variables.
- Backend: standard library first; run `go vet ./...` before pushing.
- Frontend: call the API via `` `${import.meta.env.VITE_API_URL}/...` `` so the backend URL stays configurable at build time.
- New secret? Add it in Doppler, wire it as `${VAR:-}` in `docker-compose.yml` if needed at Compose level, document it in `.env.example`.
