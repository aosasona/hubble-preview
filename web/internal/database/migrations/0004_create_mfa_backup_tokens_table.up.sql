-- MFA backup tokens can be used to recover access to an account if the user loses
-- access to their primary MFA device (10 backup tokens are generated when a user
-- enables MFA).
--
-- Only the hashed version of the token is stored in the database. The plaintext token
-- is only shown to the user once when it is generated.
CREATE TABLE IF NOT EXISTS mfa_backup_tokens (
	id SERIAL PRIMARY KEY,
	user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	hashed_token TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	used_at TIMESTAMPTZ DEFAULT NULL
)

