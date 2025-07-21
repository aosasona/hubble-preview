CREATE TABLE IF NOT EXISTS workspaces (
	id SERIAL PRIMARY KEY,
	public_id UUID NOT NULL UNIQUE DEFAULT uuid_generate_v4(),

	namespaced_name VARCHAR(255) NOT NULL UNIQUE, -- workspaces names are not unique by default but to prevent name collisions, we namespace them with the owner's user ID (e.g. "user_id/workspace_name")
	display_name VARCHAR(255) NOT NULL, -- the human readable name of the workspace
	owner_id INT NOT NULL REFERENCES users(id), -- the user who created the workspace (automatically assigned the admin role)
	slug VARCHAR(255) NOT NULL UNIQUE, -- the unique identifier for the workspace (used in the URL)
	description TEXT, -- a brief description of the workspace
	avatar_id VARCHAR(255), -- the name (AKA ID) of the avatar file as stored in Minio

	enable_public_indexing BOOLEAN NOT NULL DEFAULT FALSE, -- whether the workspace is public or not (on hubble, workspaces can be publicly indexed to display results in the global search box)
	invite_only BOOLEAN NOT NULL DEFAULT TRUE, -- whether the workspace is invite only or not (if true, only users with an invite can join the workspace, if false, anyone can REQUEST to join the workspace)

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ DEFAULT NULL
)

