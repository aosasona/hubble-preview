CREATE TABLE IF NOT EXISTS auth_sessions (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token UUID NOT NULL DEFAULT uuid_generate_v4(),
	meta JSONB NOT NULL DEFAULT '{}', -- Store additional metadata in a JSONB column, for example, user agent, IP address, etc.
	issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days' -- by default, sessions expire after 7 days
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth_sessions(user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_sessions_token ON auth_sessions(token);

