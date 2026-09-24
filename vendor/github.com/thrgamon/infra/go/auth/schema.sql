-- Shared namespaced persistence. Applications keep their existing users and
-- ownership tables; local_user_id stores the existing ID as text.
CREATE TABLE auth_users (
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    local_user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','invited','suspended','revoked')),
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX auth_sessions_expires_at ON auth_sessions(expires_at);
CREATE INDEX auth_sessions_local_user_id ON auth_sessions(local_user_id);
