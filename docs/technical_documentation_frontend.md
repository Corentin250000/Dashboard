# Dashboard Frontend

The frontend is a React 19.2.8 single-page application written in TypeScript
6.0.2 and built with Vite 8.3.0. It uses Tailwind CSS 4.3.3,
`react-grid-layout` 2.2.4, Headless UI 2.2.10, and `lucide-react` 1.52.0. It
lets users register, connect services, and build a personal dashboard from
configurable widgets.

This document describes the frontend as it exists today. It implements
authentication screens, a responsive dashboard grid, a services page, eight
widgets, and automatic and manual widget refresh. An admin section is not
implemented.

## Responsibilities

- Provide registration, email confirmation, and login screens.
- Protect dashboard and services routes by checking the current session.
- Display a responsive dashboard grid of configured widgets.
- Add, configure, move, and delete widgets.
- Start service connection flows and disconnect services.
- Fetch widget data and refresh it automatically or on request.
- Show loading, error, and notification messages for user actions and data.

## Technology Stack

| Component | Technology | Purpose |
| --- | --- | --- |
| UI | React 19.2.8 | Single-page application and component rendering |
| Language | TypeScript 6.0.2 | Frontend types and build-time checking |
| Build and development | Vite 8.3.0 | Development server and production bundle |
| Styling | Tailwind CSS 4.3.3 | Utility styles and theme tokens |
| Dashboard layout | `react-grid-layout` 2.2.4 | Responsive widget grid and desktop dragging |
| Accessible components | Headless UI 2.2.10 | Dialogs, menus, and form controls |
| Icons | `lucide-react` 1.52.0 | Interface icons |
| Authentication | HttpOnly cookie | Browser sends the session cookie with API requests |

Versions are the dependency ranges declared in `frontend/package.json`.

## Architecture

```text
+---------+     +----------------------------------------------------+
| Browser | --> | React app                                          |
+---------+     | App.tsx routes -> ProtectedRoute                   |
                | pages: auth / dashboard / subscription             |
                | dashboard: DashboardGrid -> WidgetCard             |
                | widgets registry -> widget components              |
                | useTimer + refresh context                         |
                | shared/api.ts                                      |
                +----------------------+-----------------------------+
                                       |
                                       | HTTP / JSON, credentials included
                                       v
                +----------------------+-----------------------------+
                | Go API :8080                                       |
                +----+-------+------------+------------+---------+---+
                     |       |            |            |         |
                   SNCF   Spotify  Google Calendar   Steam   Chess.com
```

```text
registry.ts -> WidgetCard (refresh context) -> widget component
                                                   |
                                                useTimer
                                                   |
                                                api.ts -> backend
```

The first diagram shows how the browser enters the React routes and how the
shared API client sends requests to the Go API. The API supplies widget data
from the external services shown. `ProtectedRoute` gates dashboard and service
management pages using the current session.

The second diagram shows widget rendering and refresh. The registry maps a
widget type to its component; `WidgetCard` supplies a refresh context, and the
component uses `useTimer` to fetch through `api.ts` on a schedule or after the
manual refresh action.

## Run

### Docker

From the repository root:

```bash
docker-compose up --build
```

The frontend is served at `http://127.0.0.1:8081`. Compose builds the client
image and maps host port `8081` to container port `8081`.

### Local development

From the repository root:

```bash
cd frontend
npm ci
npm run dev
```

The declared scripts are `dev` (Vite development server), `build`
(`tsc -b && vite build`), `lint` (`oxlint`), and `preview` (`vite preview`).

`VITE_API_URL` selects the API base URL. The shared API client defaults to
`http://127.0.0.1:8080`. In Docker, the Dockerfile accepts `VITE_API_URL` as a
build argument, and Compose passes the value from the environment (with the
same localhost URL as its default). Vite embeds this value during the build; it
is not a runtime browser setting.

## Project Layout

```text
frontend/src/
├── App.tsx                         # Routes and protected page grouping
├── main.tsx                        # React entry point and global CSS imports
├── index.css                       # Tailwind import, theme tokens, base styles
├── assets/                         # Local service and OAuth icon assets
└── features/
    ├── auth/                       # Registration, confirmation, login, session guard
    │   ├── components/             # Shared auth card and protected-route logic
    │   └── pages/                  # Authentication and email confirmation pages
    ├── dashboard/                  # Dashboard page, grid, widgets, and editing tools
    │   ├── components/             # Grid cards, toolbar, notifications, widget forms
    │   │   └── addWidget/           # Widget picker, configuration, and parameter fields
    │   ├── pages/                  # Dashboard route page
    │   ├── tools/                  # Layout, free-space, sizing, timer, refresh helpers
    │   └── widgets/                 # Widget registry and service-specific components
    │       ├── chess/               # Chess.com statistics widgets
    │       ├── google/              # Google Calendar events widget
    │       ├── sncf/                # SNCF departure and journey widgets
    │       ├── spotify/             # Spotify playback and top-items widgets
    │       └── steam/               # Steam friends widget
    ├── shared/                      # API client, API types, service icon map
    └── subscription/                # Connected-service management page
        └── pages/                   # Service connection and disconnection UI
```

## Authentication Flow

Registration sends an email and password to the API, then navigates to the
pending-verification screen. Email confirmation consumes the token from the
`/confirm?token=...` URL. After confirmation, the page navigates to the
dashboard after three seconds; the protected route still checks the session,
so a user without a valid session is sent to login. Login sends credentials
and navigates to the dashboard on success. The login page also links to the
Google and GitHub OAuth entry points.

The API client sets `credentials: 'include'` on every request, allowing the
browser to send the HttpOnly session cookie. The cookie itself is set by the
backend. `ProtectedRoute` calls `GET /auth/me`; a `401` redirects to login and
other request errors display a retry action.

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/auth/register` | Submit local account registration |
| `GET` | `/confirm?token=...` | Confirm the email token |
| `POST` | `/confirm/resend` | Request a replacement confirmation email |
| `POST` | `/auth/login` | Submit local login credentials |
| `GET` | `/auth/me` | Check the current session for protected routes |
| `GET` | `/auth/google` | Start Google OAuth from the login page |
| `GET` | `/auth/github` | Start GitHub OAuth from the login page |

## Dashboard

The desktop grid has six columns and six rows. At tablet widths (640 px and
above, below 1024 px) it uses four columns; below 640 px it uses two columns.
Dragging is enabled only at the desktop breakpoint. Tablet and mobile layouts
are derived by stacking widgets in desktop position order and are not saved.

Widgets have `small` (1x1), `wide` (2x1), `tall` (1x2), or `big` (2x2) sizes.
The registry restricts the available sizes for each widget. When adding a
widget, the frontend searches the 6x6 desktop grid for the first available
position. Desktop moves are sent to the API and saved there.

Each widget uses `useTimer` with its configured interval. 
The refresh button in `WidgetCard` increments the refresh context nonce,
which starts a fetch immediately and restarts the interval. If a fetch fails
for the same configuration, `useTimer` keeps the last successful data visible
with the error. When a widget is reconfigured, data from the previous
configuration is discarded, so an error is never shown next to outdated data. If a fetch fails,
`useTimer` records the error but keeps the last successful data visible.
Widgets display their own error messages, and dashboard actions such as failed
moves or deletes use notifications.

## Services Page

The page loads the user's services from the API. A service that requires an
account can be connected with a redirect to
`/api/me/services/{service}/connect`; the page also displays a generic error
when it returns with an `error` query parameter. Disconnecting opens a
confirmation dialog and calls the service disconnect endpoint. The dialog says
that the service's widgets will be deleted; that deletion is performed by the
backend.

## Widgets

The following table reflects the frontend registry. Parameter names and
runtime config values come from `registry.ts` and the widget components.

| Type | Service | Parameters | Sizes | Account required |
| --- | --- | --- | --- | --- |
| `sncf.departures` | SNCF | `station` (string) | big | No |
| `sncf.journeys` | SNCF | `from`, `to`, `datetime` (strings) | tall, big | No |
| `spotify.currently_playing` | Spotify | `show_progress` (integer) | small, wide, tall, big | Yes |
| `spotify.top_items` | Spotify | `item_type`, `time_range` (strings) | small, wide, tall, big | Yes |
| `calendar.upcoming_events` | Google Calendar | `maxResults` (integer), `period` (string) | tall, big | Yes |
| `steam.friends` | Steam | `limit`, `inGameOnly` (integers) | tall, big | Yes |
| `chess.stats` | Chess.com | `username`, `time_class` (strings) | wide, tall, big | No |
| `chess.monthly_stats` | Chess.com | `username`, `time_class` (strings), `year`, `month` (integers) | tall, big | No |

### Add a widget

1. Create a component under `frontend/src/features/dashboard/widgets/<service>/`.
   It receives `config` and `refreshSeconds`; use `useTimer` to load and
   refresh its data.
2. Add an entry to `widgets/registry.ts` with its component, service, parameter
   definitions, allowed sizes, account requirement, and default refresh rate.
3. Add response types and an API function in `features/shared/api.ts` for the
   backend request.
4. Add the matching widget name and parameters to the backend catalog in
   `backend/internal/about/about.go`.

For example, `SncfDeparturesWidget` reads `config.station`, calls
`api.getSncfDepartures(station)` through `useTimer`, and the registry identifies
it as `sncf.departures` with a string `station` parameter. The backend catalog
uses the widget name `departures` under service `sncf`, which the frontend
combines into the `service.widget` type.

Configuration form fields are stored as strings while the user edits them.
Integer parameters are converted to numbers when creating or saving the
widget, and values received from the API can also be numbers at runtime. Use
`String()` and `Number()` when reading or converting config values; for example,
the Steam widget uses `Number(config.limit)` and
`String(config.inGameOnly) === '1'`.

## Accessibility

The code includes `aria-label` attributes on icon-only controls and OAuth
links, status and alert roles for loading and error messages, and a progressbar
with accessible values for Spotify playback. `WidgetCard` can be selected
using Enter or Space when it has keyboard focus. The selected card is announced 
with aria-current. Focus-visible outlines use the
accent token. The manual refresh icon disables its spin animation for users
with reduced-motion preferences (`motion-reduce:animate-none`).

`index.css` defines a palette through CSS custom properties: background,
surface, border, input border, primary and secondary text, accent, text on
accent, soft accent, and secondary accent/error. 
The main text and accent pairs meet WCAG AA (4.5:1 minimum): white on
`--accent` 7.12:1, white on `--accent-secondary` 6.57:1, `--text-secondary`
on `--accent-soft` 5.22:1..

Lighthouse accessibility scores (Chrome DevTools, desktop):

| Page | Score |
| --- | --- |
| Login | 97 |
| Dashboard | 100 |
| Subscription | 97 |

## Known Limitations

- Steam shows a friend's current game only when that friend's game details are
  public.
- Tablet and mobile widget positions are derived from desktop ordering and are
  not saved.
