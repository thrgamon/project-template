-- +goose Up

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

-- Auth0 identities are provisioned explicitly into this application.  There
-- is no password or public registration path in the template.
CREATE TABLE auth_users (
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    local_user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'invited', 'suspended', 'revoked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (issuer, subject),
    UNIQUE (local_user_id)
);

CREATE TABLE auth_sessions (
    token_hash BYTEA PRIMARY KEY,
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    local_user_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);
CREATE INDEX idx_auth_sessions_local_user_id ON auth_sessions(local_user_id);

-- +goose Down

DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS auth_users;
DROP TABLE IF EXISTS users;
