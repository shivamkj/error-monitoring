# Error Monitoring

A self-hosted, Sentry-compatible error monitoring system. Drop-in replacement for Sentry — just change the DSN URL in any standard Sentry SDK (`@sentry/browser`, `@sentry/react`, etc.) to start capturing errors.

## Tech Stack

- **Backend**: Go (chi, pgx, golang-jwt)
- **Frontend**: React 19 + TypeScript + Tailwind CSS v4 + Vite + TanStack Query
- **Database**: PostgreSQL 16
- **Deployment**: Docker Compose (backend + frontend/nginx + postgres)

## Features

- Sentry-compatible ingestion (envelope and legacy store endpoints)
- Automatic error grouping via stack trace fingerprinting
- Issue aggregation with browser, OS, device, and URL statistics
- Dashboard with filterable/sortable issue list
- Full event detail with stack traces, breadcrumbs, and context
- Project management with auto-generated DSN keys
- JWT-based dashboard authentication
- Status management (unresolved/resolved/reappeared/ignored)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- (For local development) Go 1.26+, Node.js 20+

## Quick Start

1. Clone the repository:

   ```bash
   git clone https://github.com/qnify/error-monitoring.git
   cd error-monitoring
   ```

2. Copy the environment file and configure it:

   ```bash
   cp .env.example .env
   ```

3. Start all services:

   ```bash
   docker compose up -d
   ```

4. Open `http://localhost` in your browser and create an account.

5. Create a project to get your DSN, then configure your Sentry SDK:

   ```javascript
   import * as Sentry from "@sentry/react";

   Sentry.init({
     dsn: "http://<public_key>@localhost/<project_id>",
   });
   ```

## Local Development

### Backend

```bash
cd backend
go run ./cmd/server
```

The server starts on port `8080` by default. Requires a running PostgreSQL instance (configure via `DATABASE_URL` environment variable).

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server starts on port `5173` with hot module replacement.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | — |
| `JWT_SECRET` | Secret key for signing JWT tokens (min 32 chars) | — |
| `BASE_URL` | Public URL of the service | `http://localhost:8080` |
| `ALLOWED_ORIGINS` | Comma-separated CORS origins | `http://localhost:3000,http://localhost:5173` |

## API Overview

### Sentry-Compatible Ingestion

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/{project_id}/envelope/` | DSN public key |
| POST | `/api/{project_id}/store/` | DSN public key |

### Dashboard API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/dashboard/auth/register` | Create account |
| POST | `/api/dashboard/auth/login` | Login (returns JWT) |
| GET | `/api/dashboard/projects` | List projects |
| POST | `/api/dashboard/projects` | Create project |
| GET | `/api/dashboard/projects/{pid}/issues` | List issues |
| GET | `/api/dashboard/issues/{id}` | Issue detail |
| PUT | `/api/dashboard/issues/{id}/status` | Update issue status |
| GET | `/api/dashboard/issues/{id}/events` | List events for issue |
| GET | `/api/dashboard/events/{id}` | Full event detail |

## Project Structure

```
error-monitoring/
├── backend/
│   ├── cmd/server/          # Application entrypoint
│   └── internal/
│       ├── config/          # Environment-based configuration
│       ├── database/        # Connection pool + migrations
│       ├── handlers/        # HTTP handlers (ingest + dashboard)
│       ├── middleware/      # Auth, CORS, logging
│       ├── models/          # Domain types
│       ├── processing/      # Error pipeline (parse, normalize, fingerprint, aggregate)
│       └── router/          # Route definitions
├── frontend/
│   └── src/                 # React SPA
├── nginx.conf               # Reverse proxy config
├── docker-compose.yml
└── .env.example
```

## How It Works

Incoming errors flow through a processing pipeline:

1. **Decompress** — handle gzip-encoded payloads
2. **Authenticate** — validate DSN public key against the project
3. **Parse** — extract error data from Sentry envelope format
4. **Normalize** — mask dynamic values (UUIDs, timestamps, IDs) for consistent grouping
5. **Fingerprint** — generate a SHA-256 hash from stack trace or error message
6. **Upsert Issue** — create or update the aggregated issue record
7. **Store Event** — persist the full raw event for detailed inspection

## License

AGPL 3.0
