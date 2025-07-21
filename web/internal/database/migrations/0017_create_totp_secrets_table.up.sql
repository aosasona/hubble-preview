CREATE TABLE IF NOT EXISTS totp_secrets (
	id SERIAL PRIMARY KEY,
	account_id UUID NOT NULL REFERENCES mfa_accounts(id) ON DELETE CASCADE,
	hash BYTEA NOT NULL,
	version SMALLINT NOT NULL DEFAULT 1,

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ,

	UNIQUE (account_id, version),
	CHECK (version > 0)
);

