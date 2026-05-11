# Error Monitoring System - Implementation Plan

## Context

Building a self-hosted, Sentry-compatible error monitoring system. The goal is a simple, easily deployable tool that accepts errors from standard Sentry SDKs (`@sentry/browser`, `@sentry/react`) by just changing the DSN URL. Inspired by GlitchTip but with a simpler tech stack: single Go binary + React SPA + PostgreSQL.

Key design principle: **aggregate common data on the issue level, store unique data as raw events**. This keeps the issues list fast while retaining full detail for drill-down.

---

## Tech Stack

- **Backend**: Go (chi router, pgx, golang-jwt)
- **Frontend**: React + TypeScript + Tailwind CSS + Vite + TanStack Query
- **Database**: PostgreSQL 16
- **Deployment**: Docker Compose (backend + frontend/nginx + postgres)

---

## Project Structure

```
error-monitoring/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── database/
│   │   │   ├── db.go
│   │   │   ├── migrate.go
│   │   │   └── migrations/001_initial.up.sql
│   │   ├── middleware/
│   │   │   ├── auth.go          # JWT for dashboard
│   │   │   ├── cors.go
│   │   │   └── logging.go
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── project.go
│   │   │   ├── issue.go
│   │   │   └── event.go
│   │   ├── handlers/
│   │   │   ├── ingest/
│   │   │   │   ├── envelope.go  # POST /api/{project_id}/envelope/
│   │   │   │   ├── store.go    # POST /api/{project_id}/store/
│   │   │   │   └── auth.go     # DSN key validation
│   │   │   ├── dashboard/
│   │   │   │   ├── auth.go
│   │   │   │   ├── projects.go
│   │   │   │   ├── issues.go
│   │   │   │   └── events.go
│   │   │   └── health.go
│   │   ├── processing/
│   │   │   ├── pipeline.go     # Orchestrates: parse -> normalize -> hash -> store
│   │   │   ├── parser.go       # Parse envelope/store payloads
│   │   │   ├── normalizer.go   # Mask UUIDs, IDs, timestamps, etc.
│   │   │   ├── fingerprint.go  # Generate SHA-256 hash for grouping
│   │   │   └── aggregator.go   # Update issue-level JSONB stats
│   │   └── router/router.go
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── api/                 # API client + endpoint functions
│   │   ├── components/
│   │   │   ├── layout/          # Sidebar, Header, Layout
│   │   │   ├── issues/          # IssueList, IssueRow, StackTrace, Breadcrumbs
│   │   │   ├── projects/        # ProjectList, ProjectCreate, DSNDisplay
│   │   │   ├── auth/            # LoginForm, RegisterForm
│   │   │   └── common/          # Badge, Button, Pagination, TimeAgo
│   │   ├── pages/
│   │   ├── context/AuthContext.tsx
│   │   └── hooks/
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── nginx.conf
└── .env.example
```

---

## Database Schema

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    platform    VARCHAR(50) DEFAULT 'javascript',
    public_key  VARCHAR(64) NOT NULL UNIQUE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Issues = aggregated error groups
CREATE TABLE issues (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    fingerprint VARCHAR(64) NOT NULL,           -- SHA-256 hash
    title       VARCHAR(1024) NOT NULL,         -- e.g. "TypeError: Cannot read property 'x'"
    culprit     VARCHAR(1024),                  -- file/function where error occurred
    level       VARCHAR(20) NOT NULL DEFAULT 'error',
    platform    VARCHAR(50),
    status      VARCHAR(20) NOT NULL DEFAULT 'unresolved', -- unresolved/resolved/reappeared/ignored
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_count INTEGER NOT NULL DEFAULT 1,
    -- Aggregate stats (JSONB): {"Chrome 120": 45, "Firefox 121": 12}
    browsers    JSONB DEFAULT '{}',
    os_names    JSONB DEFAULT '{}',
    devices     JSONB DEFAULT '{}',
    urls        JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, fingerprint)
);

-- Events = raw individual occurrences (unique per-event data)
CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    VARCHAR(32) NOT NULL,           -- Sentry event_id
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    timestamp   TIMESTAMPTZ NOT NULL,
    level       VARCHAR(20) NOT NULL DEFAULT 'error',
    platform    VARCHAR(50),
    ip_address  INET,
    user_data   JSONB,                          -- {id, email, username}
    request_data JSONB,                         -- {url, method, headers}
    breadcrumbs JSONB,                          -- [{timestamp, category, message}]
    contexts    JSONB,                          -- {device, os, browser}
    tags        JSONB,
    exception   JSONB,                          -- Full stack trace data
    message     TEXT,
    environment VARCHAR(100),
    release_tag VARCHAR(200),
    server_name VARCHAR(255),
    raw_payload JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## API Endpoints

### Sentry-Compatible Ingestion (auth via DSN public_key)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/{project_id}/envelope/` | Envelope endpoint (primary) |
| POST | `/api/{project_id}/store/` | Legacy store endpoint |

### Dashboard API (auth via JWT Bearer token)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/dashboard/auth/register` | Create account |
| POST | `/api/dashboard/auth/login` | Login, returns JWT |
| GET | `/api/dashboard/auth/me` | Current user info |
| GET | `/api/dashboard/projects` | List projects |
| POST | `/api/dashboard/projects` | Create project (generates DSN) |
| GET | `/api/dashboard/projects/{id}` | Project detail + DSN |
| DELETE | `/api/dashboard/projects/{id}` | Delete project |
| GET | `/api/dashboard/projects/{pid}/issues` | List issues (filter/sort/paginate) |
| GET | `/api/dashboard/issues/{id}` | Issue detail with aggregates |
| PUT | `/api/dashboard/issues/{id}/status` | Update status (resolve/ignore/unresolve) |
| DELETE | `/api/dashboard/issues/{id}` | Delete issue + events |
| GET | `/api/dashboard/issues/{id}/events` | List events for issue |
| GET | `/api/dashboard/events/{id}` | Full event detail |

---

## Error Processing Pipeline

```
Request -> Decompress (gzip) -> Authenticate (DSN key) -> Parse Envelope -> Normalize -> Fingerprint -> Upsert Issue -> Store Event
```

### Normalization (before hashing)

Mask dynamic values so the same logical error always produces the same hash:
- UUIDs (`8-4-4-4-12` hex) -> `<uuid>`
- Numeric IDs (5+ digits) -> `<id>`
- Timestamps (ISO 8601) -> `<timestamp>`
- Hex addresses (`0x...`) -> `<addr>`
- Email addresses -> `<email>`
- URLs -> `<url>`
- Webpack chunk hashes in filenames -> stripped

### Fingerprint Priority

1. Custom fingerprint (if SDK sets one, not `{{ default }}`)
2. Stack trace: concat normalized `filename:function:lineno` for in-app frames
3. Exception type + normalized value
4. Normalized message

All fed through SHA-256 to produce a 64-char hex fingerprint.

### Issue Upsert (single atomic query)

```sql
INSERT INTO issues (project_id, fingerprint, title, culprit, level, platform, first_seen, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
ON CONFLICT (project_id, fingerprint) DO UPDATE SET
    last_seen = EXCLUDED.last_seen,
    event_count = issues.event_count + 1,
    status = CASE
        WHEN issues.status = 'resolved' THEN 'reappeared'
        ELSE issues.status
    END,
    updated_at = NOW()
RETURNING id, (xmax = 0) AS is_new;
```

---

## Authentication

- **Dashboard**: JWT (7-day expiry), stored in localStorage, sent as `Authorization: Bearer <token>`
- **SDK ingestion**: DSN public_key sent via `X-Sentry-Auth` header or `?sentry_key=` query param. Validated against `projects.public_key` column.

DSN format shown to users: `http(s)://<public_key>@<host>/<project_id>`

---

## Frontend Pages

1. **Login/Register** - email + password forms
2. **Projects List** - cards showing project name + issue count, create button
3. **Project Settings** - DSN display with copy button, setup instructions
4. **Issues List** - filterable table (status tabs: unresolved/resolved/reappeared/ignored), sortable by last_seen/count/first_seen
5. **Issue Detail** - title, status actions, stack trace, breadcrumbs, event timeline, browser/OS/device aggregates

---

## Implementation Order

### Phase 1: Backend Foundation
1. Init Go module, install deps (chi, pgx, golang-jwt, bcrypt)
2. Config loading from env vars
3. Database connection + migration runner
4. Write migration SQL
5. Health check endpoint
6. Basic router wiring

### Phase 2: Ingestion Pipeline
1. Envelope parser (newline-delimited format)
2. Store (legacy JSON) parser
3. DSN key auth extraction + validation
4. Normalizer (regex-based masking)
5. Fingerprint generator (priority chain + SHA-256)
6. Aggregator (JSONB stats update)
7. Pipeline orchestrator (parse -> normalize -> hash -> upsert issue -> store event)
8. Wire up `/api/{project_id}/envelope/` and `/store/` handlers

### Phase 3: Dashboard API
1. User model + bcrypt password handling
2. JWT middleware
3. Auth handlers (register, login, me)
4. Project CRUD handlers (with public_key generation)
5. Issues list/detail/status handlers
6. Events list/detail handlers
7. CORS middleware

### Phase 4: Frontend
1. Vite + React + Tailwind + Router setup
2. Auth context + API client with JWT interceptor
3. Login/Register pages
4. Layout (sidebar + header)
5. Projects page + create + DSN display
6. Issues list page with filters/sort/pagination
7. Issue detail page with stack trace + breadcrumbs + event timeline + context panel

### Phase 5: Integration & Deployment
1. Test with real `@sentry/react` SDK
2. Docker setup (Dockerfiles + docker-compose + nginx)
3. Verify full deployment from scratch

---

## Verification Plan

1. **Unit test the normalizer**: feed known strings with UUIDs/IDs, verify masked output
2. **Unit test the fingerprint**: verify same logical error -> same hash, different errors -> different hashes
3. **Integration test ingestion**: curl a real Sentry envelope payload, verify issue + event created
4. **Test deduplication**: send same error twice, verify event_count=2 and only 1 issue
5. **Test reappearance**: resolve an issue, send same error again, verify status=reappeared
6. **End-to-end with Sentry SDK**: create a React app with `@sentry/react`, point DSN to backend, trigger errors, verify they appear in dashboard
7. **Docker compose up from empty state**: verify migrations run, backend starts, frontend serves, full flow works
