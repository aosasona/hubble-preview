CREATE TABLE IF NOT EXISTS collection_members (
	id SERIAL PRIMARY KEY,

	collection_id INT NOT NULL REFERENCES collections(id),
	user_id INT NOT NULL REFERENCES users(id),

	extra_permissions JSONB, -- extra permissions for the user in the collection (e.g. invite:users, kick:users, etc.) - these are collection-specific

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ DEFAULT NULL, -- when the user was removed from the collection

	UNIQUE (collection_id, user_id)
);

