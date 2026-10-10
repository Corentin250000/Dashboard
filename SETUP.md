# DASHBOARD Setup

This guide prepares the DASHBOARD locally. It covers configuration and
running the completed project; it does not explain how to develop it.

## 1. Prerequisites

Install the following:

1. Git.
2. Go 1.25 or newer for a native backend run.
3. Node.js 22 or newer and npm for a native frontend run.
4. PostgreSQL 16 or Docker Desktop/Docker Engine with Docker Compose.
5. The Doppler CLI is optional. The Docker images already install the Doppler
   CLI, so it is not required when using Docker Compose.

Clone the repository and enter it:

```bash
git clone git@github.com:EpitechPGE3-2026/G-WEB-500-MLH-5-1-dashboard-8.git
cd G-WEB-500-MLH-5-1-dashboard-8
```

The normal local URLs are:

- Frontend: `http://127.0.0.1:8081`
- API: `http://127.0.0.1:8080`

Use `127.0.0.1`, not `localhost`, in browser URLs, provider settings, and
environment variables.

## 2. Choose an environment-variable method

The screenshot supplied with this project is the source of truth for the
shared variable names. The application itself reads the names listed below.

### Method A: classic `.env`

1. Create a private local file from the template:

```bash
cp .env.example .env
```

2. Edit `.env` and replace the placeholders. For this method, leave
   `DOPPLER_TOKEN` empty. The database, server, and API credentials must be in
   `.env`.
3. Keep `.env` local. It is ignored by Git and must never be committed,
   uploaded, or pasted into an issue. `.env.example` contains placeholders only.
4. If you use Docker Compose, its PostgreSQL hostname is `db` internally. Do
   not set `DB_HOST=127.0.0.1` for the containerized backend. For a native
   backend connecting to PostgreSQL on the host, set `DB_HOST=127.0.0.1` and
   optionally `DB_PORT=5432`.

### Method B: Doppler

With this method, the root `.env` contains only the Doppler service token. The
other values live in Doppler.

1. Sign in to Doppler and choose **Projects** → **Create Project**. Use a name
   such as `dashboard`.
2. Create or select a configuration, for example `dev`.
3. In the project/configuration, add every required variable from the reference
   table using the exact spelling. Do not add quotes around ordinary values.
4. Open **Project** → select the configuration → **Access** → **Service Tokens**
   → **Generate**. Give the token read access to that configuration and copy it
   once. Provider labels can vary slightly; the goal is a service token scoped
   to the project and configuration.
5. Put the token in the root `.env` under the exact name expected by Compose:

```dotenv
DOPPLER_TOKEN=dp.st.your_service_token_here
```

6. Do not put the other Doppler values in `.env` for this method. The server and
   database images include the Doppler CLI. Their entrypoints run
   `doppler run -- ...` when `DOPPLER_TOKEN` is non-empty, which fetches and
   injects the project/config values before PostgreSQL or the API starts.

The repository's `doppler.yaml` maps the default CLI context to project
`dashboard`, config `dev`; Compose still receives the token from `.env`.
`VITE_API_URL` is special: Vite embeds it during the frontend image build, and
   the Compose build argument comes from the host `.env`. Set it in the host
   environment or `.env` before `docker-compose build`; a value fetched only
   inside the backend container cannot change an already-built frontend.

## 3. Variable reference

Values below are fake examples. Never use them as real credentials.

### Database

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `POSTGRES_DB` | string | Database name | `dashboard` | Choose it when creating PostgreSQL |
| `POSTGRES_USER` | string | Database user | `dashboard` | Choose it when creating PostgreSQL |
| `POSTGRES_PASSWORD` | string | Database password | `change_me_please` | Generate a private password |
| `DB_HOST` | string, optional | Native backend database host; default is `db` | `127.0.0.1` | Use `127.0.0.1` natively, `db` in Compose |
| `DB_PORT` | string, optional | PostgreSQL port; default is `5432` | `5432` | Use the port PostgreSQL exposes |

### Server and URLs

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `PORT` | integer | API listen port | `8080` | Choose `8080` locally |
| `FRONTEND_URL` | URL | CORS origin and email confirmation-link base | `http://127.0.0.1:8081` | Use the actual frontend origin |
| `VITE_API_URL` | URL | API base URL embedded by Vite | `http://127.0.0.1:8080` | Use the actual browser-reachable API URL |

### Security

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `JWT_SECRET` | secret string | Reserved security configuration value | `random_secret_value` | Generate locally with OpenSSL |

### Email

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `API_KEY_MAILJET` | secret string | Mailjet API username/key for confirmation email | `mailjet_key` | Mailjet API Key Management |
| `SECRET_KEY_MAILJET` | secret string | Mailjet API secret/password | `mailjet_secret` | Mailjet API Key Management |
| `MAIL_SENDER_ACCOUNT` | email | Validated `From` address | `verified@example.com` | Mailjet Sender addresses |
| `MDP_SENDER_ACCOUNT` | secret string | Not read by this repository | `unused_password` | No value is needed by the current code |

### OAuth

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `CLIENT_ID_GOOGLE` | string | Google account login and Google Calendar service OAuth client ID | `google-client-id` | Google Cloud OAuth credentials |
| `CLIENT_SECRET_GOOGLE` | secret string | Google account login and Google Calendar service OAuth client secret | `google-client-secret` | Google Cloud OAuth credentials |
| `CLIENT_ID_GITHUB` | string | GitHub account login OAuth client ID | `github-client-id` | GitHub OAuth App |
| `CLIENT_SECRET_GITHUB` | secret string | GitHub account login OAuth client secret | `github-client-secret` | GitHub OAuth App |
| `CLIENT_ID_SPOTIFY` | string | Spotify service OAuth client ID | `spotify-client-id` | Spotify Developer Dashboard |
| `CLIENT_SECRET_SPOTIFY` | secret string | Spotify service OAuth client secret | `spotify-client-secret` | Spotify Developer Dashboard |

### External APIs

| Name | Type | Purpose | Fake example | Obtain/create |
|---|---|---|---|---|
| `API_KEY_SNCF` | secret string | SNCF Navitia Basic Auth username; imports stations and serves SNCF departures | `sncf-token` | SNCF API portal |
| `API_KEY_STEAM` | secret string | Steam OpenID and Web API key for the Steam service | `steam-key` | Steam Web API key page |

## 4. Obtain or create provider values

### PostgreSQL

Choose values for `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD`.
The easiest setup uses the PostgreSQL service in this repository's Compose file.
Those values are passed to the `db` container automatically.

For a separate local PostgreSQL installation, create the user and database with
the same values. The exact SQL is:

```sql
CREATE USER dashboard WITH PASSWORD 'your_database_password_here';
CREATE DATABASE dashboard OWNER dashboard;
```

Use your chosen names instead of `dashboard`. Do not commit the password.

### Server URLs and JWT secret

Use these local values:

```dotenv
PORT=8080
FRONTEND_URL=http://127.0.0.1:8081
VITE_API_URL=http://127.0.0.1:8080
```

Generate `JWT_SECRET` with:

```bash
openssl rand -base64 48
```

The current login implementation uses an HTTP-only session cookie; the
`JWT_SECRET` setting is retained as a security configuration value.

### Mailjet

1. Create or sign in to a Mailjet account.
2. Open **Account Settings** → **API Key Management** and copy the API key and
   secret key into `API_KEY_MAILJET` and `SECRET_KEY_MAILJET`.
3. Open **Account** → **Sender addresses** (the wording may be **Senders &
   Domains**), choose **Add a sender**, enter an address you control, and
   complete the verification email. Use that exact address for
   `MAIL_SENDER_ACCOUNT`.
4. The code calls Mailjet with the API key and secret, and uses
   `MAIL_SENDER_ACCOUNT` as the message's `From` address.

### Google OAuth

1. In Google Cloud Console, use **Project selector** → **New Project**.
2. Open **APIs & Services** → **OAuth consent screen**, choose **External**, set
   the app information, and add your test account under **Test users**.
3. Open **APIs & Services** → **Credentials** → **Create Credentials** →
   **OAuth client ID** → **Web application**.
4. Add `http://127.0.0.1:8081` under **Authorized JavaScript origins**.
5. Add a redirect URI only when the application implements a callback.
6. Copy the client ID and secret to `CLIENT_ID_GOOGLE` and
   `CLIENT_SECRET_GOOGLE` only for that future implementation.

### GitHub OAuth

Use GitHub **Settings** → **Developer settings** →
**OAuth Apps** → **New OAuth App**. Set **Homepage URL** to
`http://127.0.0.1:8081`, set the authorization callback URL,
then choose **Register application** → **Generate
a new client secret**. Copy the client ID and secret to the two GitHub
variables.

### SNCF API

1. Open the SNCF/Navitia API portal and create an account or sign in.
2. Create an application/token and copy its API key to `API_KEY_SNCF`.

The backend sends the SNCF key as the Basic Auth username with an empty
password. At first startup it imports SNCF stations when the table is empty.

### Steam API

Steam keys can be obtained from the Steam Web API key page at
`https://steamcommunity.com/dev/apikey` after signing in.

## 5. Database setup and migrations

### Docker Compose

Start PostgreSQL as part of the project stack:

```bash
docker-compose up --build
```

The database image initializes PostgreSQL with `POSTGRES_DB`,
`POSTGRES_USER`, and `POSTGRES_PASSWORD`. On startup the Go server connects to
`db:5432` and calls GORM `AutoMigrate` for the models in
`backend/internal/models`. The server also imports SNCF stations if the table
is empty and `API_KEY_SNCF` is available.

The SQL files in `db/migrations/` document the intended schema, but they are
not executed by a separate migration command in the current startup path.
`AutoMigrate` is what applies the schema when the server starts.

To reset the database and rerun initialization, note that this deletes all
local data:

```bash
docker-compose down -v
```

### Native PostgreSQL

Start PostgreSQL using your operating system service, create the database and
user with the values above, then set the native host:

```dotenv
DB_HOST=127.0.0.1
DB_PORT=5432
```

Starting the backend with `go run` applies the same GORM `AutoMigrate` calls.

## 6. Run the project

### Method A: classic `.env`

Fill `.env` using Method A, then run both containers:

```bash
docker-compose up --build
```

For a native run, start PostgreSQL first, set `DB_HOST=127.0.0.1`, then use
two terminals. In the backend terminal:

```bash
cd backend
go mod download
set -a
. ../.env
set +a
export DB_HOST=127.0.0.1 DB_PORT=5432
go run ./cmd/server
```

In the frontend terminal:

```bash
cd frontend
npm ci
set -a
. ../.env
set +a
npm run dev -- --host 127.0.0.1 --port 8081
```

### Method B: Doppler

Put only `DOPPLER_TOKEN` in `.env`, add the other values to the selected
Doppler configuration, and build/run Compose:

```bash
docker-compose up --build
```

## 7. Verification checklist

1. Open `http://127.0.0.1:8080/health`; it should return `{"status":"ok"}`.
2. Open `http://127.0.0.1:8080/about.json`; it should list the SNCF service.
3. Open `http://127.0.0.1:8081` and register an account.
4. Confirm the account using the Mailjet email and then log in.
5. Open the subscription page and request an SNCF departures widget.
6. Google and GitHub login require their provider credentials and exact callback
   URLs. Spotify, Google Calendar, and Steam service connections require their
   corresponding service credentials. Provider callback URLs are documented in
   `docs/technical_documentation_backend.md` and must be registered exactly.

## 8. Troubleshooting

### `redirect_uri_mismatch`

This error means the provider's callback does not exactly match the callback
sent by the application.

### CORS errors

Set `FRONTEND_URL` to `http://127.0.0.1:8081` and set
`VITE_API_URL` to `http://127.0.0.1:8080`. Restart/rebuild after changes. The
backend allows these local origins and credentials are enabled.

### Mailjet says the sender is unverified

Verify the address under Mailjet **Account** → **Sender addresses**, then use
that exact address in `MAIL_SENDER_ACCOUNT`. The API key and secret are
`API_KEY_MAILJET` and `SECRET_KEY_MAILJET`.

### Wrong Doppler token or config

Generate a new read-capable token from the intended project/configuration at
**Access** → **Service Tokens**. Check `DOPPLER_TOKEN` spelling and remove
placeholder text. A bad non-empty token makes the container fail; an empty token
selects the classic `.env` path.

### Database connection refused

For Compose, use `DB_HOST=db` or omit it and wait for the healthy `db` service.
For a native backend, use `DB_HOST=127.0.0.1`, confirm PostgreSQL is running,
and check port `5432`, database name, user, and password. If credentials changed
after a Compose volume was initialized, reset the volume with
`docker-compose down -v` and start again.

### Mixing `localhost` and `127.0.0.1`

Use `127.0.0.1` consistently in the browser, `FRONTEND_URL`,
`VITE_API_URL`, and provider origins. Browsers treat the two hostnames as
different origins.

### Steam data is empty

Steam profile data may be private.
