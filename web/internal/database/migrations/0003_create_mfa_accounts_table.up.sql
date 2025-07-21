CREATE TABLE IF NOT EXISTS mfa_accounts (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

	account_type mfa_account_type NOT NULL,
	meta JSONB NOT NULL, -- JSON object containing account-specific data (for example, totp secret)
	active BOOLEAN NOT NULL DEFAULT TRUE, -- whether the account is active or not, accounts can be deactivated but not deleted

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_used_at TIMESTAMPTZ DEFAULT NULL -- when the account was last used
)

