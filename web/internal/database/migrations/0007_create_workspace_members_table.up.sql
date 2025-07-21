CREATE TABLE IF NOT EXISTS workspace_members (
	id SERIAL PRIMARY KEY,

	workspace_id INT NOT NULL REFERENCES workspaces(id),
	user_id INT NOT NULL REFERENCES users(id),

	bitmask_role INT NOT NULL, -- bitmask of roles
	extra_permissions JSONB, -- extra permissions for the user in the workspace (e.g. invite:users, kick:users, etc.)
	is_active BOOLEAN NOT NULL DEFAULT FALSE, -- whether the user is active in the workspace or not (e.g. invited but not yet accepted)

	invited_at TIMESTAMPTZ, -- when the user was invited to the workspace
	joined_at TIMESTAMPTZ, -- when the user joined the workspace

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ DEFAULT NULL, -- when the user was removed from the workspace

	UNIQUE (workspace_id, user_id)
);

