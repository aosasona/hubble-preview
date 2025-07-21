CREATE TABLE workspace_invites (
	id SERIAL PRIMARY KEY,
-- invite_id is a UUID that is used to generate a unique invite link for the user
	invite_id UUID NOT NULL UNIQUE DEFAULT uuid_generate_v4(),

	workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
-- member_id is the id of the workspace member that was invited to the workspace (not
-- the user that was invited)
	member_id INTEGER NOT NULL REFERENCES workspace_members(id),
-- invited_by is the id of the user that invited the member to the workspace
	invited_by INTEGER NOT NULL REFERENCES users(id),

	invited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
-- these will be used to track the status of the invite
	accepted_at TIMESTAMPTZ DEFAULT NULL,
	declined_at TIMESTAMPTZ DEFAULT NULL,
	deleted_at TIMESTAMPTZ DEFAULT NULL -- soft delete or cancellation
);

