-- Remove is_active column from workspace members table
ALTER TABLE workspace_members DROP COLUMN is_active;

-- Use email instead of member_id for invites
ALTER TABLE workspace_invites
DROP COLUMN member_id,
ADD COLUMN email VARCHAR(255) NOT NULL,
ADD COLUMN role INTEGER NOT NULL DEFAULT 0,
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Add unique constraint to workspace invites
ALTER TABLE workspace_invites
ADD CONSTRAINT unique_workspace_invite UNIQUE (workspace_id, email);

-- Add updated_at trigger to workspace invites
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON workspace_invites
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

