# Secrets management with Doppler

All secrets are managed with **Doppler**. No real secret is ever committed to git:
the only secret file is the local, gitignored `.env`, and in the Doppler setup it
contains a single line: the service token.

## How it works

The Doppler CLI is baked into the service images (`backend/Dockerfile` and
`db/Dockerfile`, installed with the official script, no package manager):

- `server` starts with `doppler run -- /server` (see `backend/entrypoint.sh`)
- `db` starts with `doppler run -- docker-entrypoint.sh postgres` (see `db/entrypoint.sh`)

At container startup, `doppler run` fetches the secrets from Doppler and injects
them into the application process. Doppler values take precedence over the
variables coming from `docker-compose.yml` / `.env`.

The launch command never changes:

```bash
docker compose up --build
```

## The service token (`.env`)

`docker-compose.yml` reads `DOPPLER_TOKEN` from the `.env` file at the repository
root (native Compose interpolation: `DOPPLER_TOKEN: ${DOPPLER_TOKEN:-}`) and
passes it into the `server` and `db` containers. That token is the **only thing**
Doppler needs inside the containers — every other secret is fetched from the
Doppler dashboard, never written to disk.

Get a token:

1. Open the Doppler dashboard → project `dashboard` → config `dev` (or `prod`).
2. Go to `Access` → `Service Tokens` → `Generate` (read-only is enough).
3. Copy `.env.example` to `.env`.
4. Fill in: `DOPPLER_TOKEN=dp.st.<...>` — then `docker compose up`.

Token behavior at startup:

| `DOPPLER_TOKEN` | Result |
|---|---|
| Valid service token | Secrets are fetched from Doppler, app starts with them |
| Invalid / placeholder value | Startup fails (`Invalid Auth token`) — replace it with a real token |
| Empty / unset | `doppler run` is skipped, the app starts with classic variables / fallbacks (this is the no-Doppler path, e.g. grading) |

Because of the fallback row, `docker compose up` keeps working for anyone without
a Doppler account: leave `DOPPLER_TOKEN` empty and fill the classic variables in
`.env` instead (see `.env.example`, "Method B"). For that reason the
`DOPPLER_TOKEN` line in `.env.example` is intentionally **commented out** by
default, so a plain `cp .env.example .env && docker compose up` starts without
Doppler.

## Secrets reference

Recreate these names in the Doppler dashboard (`dev` and `prod` configs):

| Variable | Service | Usage |
|---|---|---|
| `DOPPLER_TOKEN` | server, db | Doppler service token (from `.env`, never committed) |
| `PORT` | server | Backend listen port (`8080`, required by the subject) |
| `JWT_SECRET` | server | Secret used to sign JWTs (generate a strong random value) |
| `FRONTEND_URL` | server | Frontend URL (CORS / OAuth redirects) |
| `API_KEY_SNCF` | server | External API key (injected by `doppler run` even without a Compose entry) |
| `API_KEY_MAILJET` / `SECRET_KEY_MAILJET` / `MAIL_SENDER_ACCOUNT` | server | Mailjet API credentials for confirmation emails + the sender email address |
| `POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` | db | Postgres credentials. Must stay in sync with the initialized `pgdata` volume: if you change them, recreate the volume with `docker compose down -v`, otherwise the database keeps the old credentials silently |
| `VITE_API_URL` | — | Informational only: the frontend build arg is resolved on the host at build time, so the Doppler value is not picked up by the client build (it uses the Compose default `http://localhost:8080`) |

## CI / production

1. Generate a service token for the `prod` config in the Doppler dashboard.
2. Store it as a CI secret (e.g. `DOPPLER_TOKEN`), never in git.
3. In the pipeline:

```bash
DOPPLER_TOKEN="$DOPPLER_TOKEN" docker compose up --build -d
curl -sf http://localhost:8080/about.json
```

## Security rules

- Never commit `.env` or `doppler.yaml` (both gitignored).
- `.env.example` holds key names and placeholders only.
- A new secret means: add it in Doppler, and if a service needs it explicitly at
  Compose level, wire it as `${VAR:-}` in `docker-compose.yml` plus document it
  in `.env.example`.
- If a token leaks (logs, screenshots, shared machine), revoke it in the
  Doppler dashboard and generate a new one.
