CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

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

CREATE INDEX idx_projects_public_key ON projects(public_key);
CREATE INDEX idx_projects_user_id ON projects(user_id);

CREATE TABLE issues (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    fingerprint VARCHAR(64) NOT NULL,
    title       VARCHAR(1024) NOT NULL,
    culprit     VARCHAR(1024),
    level       VARCHAR(20) NOT NULL DEFAULT 'error',
    platform    VARCHAR(50),
    status      VARCHAR(20) NOT NULL DEFAULT 'unresolved',
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_count INTEGER NOT NULL DEFAULT 1,
    browsers    JSONB DEFAULT '{}',
    os_names    JSONB DEFAULT '{}',
    devices     JSONB DEFAULT '{}',
    urls        JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, fingerprint)
);

CREATE INDEX idx_issues_project_status ON issues(project_id, status);
CREATE INDEX idx_issues_project_last_seen ON issues(project_id, last_seen DESC);
CREATE INDEX idx_issues_project_event_count ON issues(project_id, event_count DESC);

CREATE TABLE events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     VARCHAR(32) NOT NULL,
    issue_id     UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    timestamp    TIMESTAMPTZ NOT NULL,
    level        VARCHAR(20) NOT NULL DEFAULT 'error',
    platform     VARCHAR(50),
    ip_address   INET,
    user_data    JSONB,
    request_data JSONB,
    breadcrumbs  JSONB,
    contexts     JSONB,
    tags         JSONB,
    exception    JSONB,
    message      TEXT,
    environment  VARCHAR(100),
    release_tag  VARCHAR(200),
    server_name  VARCHAR(255),
    raw_payload  JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_issue_id ON events(issue_id, timestamp DESC);
CREATE INDEX idx_events_project_id ON events(project_id, timestamp DESC);
CREATE INDEX idx_events_event_id ON events(event_id);
