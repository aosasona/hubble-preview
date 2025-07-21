CREATE TABLE IF NOT EXISTS collections (
	id SERIAL PRIMARY KEY,
	public_id UUID NOT NULL UNIQUE DEFAULT uuid_generate_v4(),

	name VARCHAR(255) NOT NULL,
	workspace_id INT NOT NULL REFERENCES workspaces(id),
	slug VARCHAR(255) NOT NULL UNIQUE, -- the unique identifier for the collection (used in the URL)
	description TEXT, -- a brief description of the collection
	avatar_id VARCHAR(255), -- the name (AKA ID) of the avatar file as stored in Minio

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ DEFAULT NULL,

	UNIQUE (workspace_id, name)
);

