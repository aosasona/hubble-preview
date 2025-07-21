create or replace function cascade_workspace_soft_delete()
returns trigger
as $$
BEGIN
	-- Soft-delete all entries in all collections in the workspace
	UPDATE entries
		SET deleted_at = NOW()
	WHERE collection_id IN (SELECT id FROM collections WHERE workspace_id = OLD.id)
		AND deleted_at IS NULL;

	-- Soft-delete all collection members
	UPDATE collection_members
		SET deleted_at = NOW()
	WHERE collection_id IN (SELECT id FROM collections WHERE workspace_id = OLD.id)
		AND deleted_at IS NULL;

	-- Soft-delete all collections in the workspace
	UPDATE collections
		SET deleted_at = NOW()
	WHERE workspace_id = OLD.id AND deleted_at IS NULL;

	-- Soft-delete all workspace members
	UPDATE workspace_members
		SET deleted_at = NOW()
	WHERE workspace_id = OLD.id AND deleted_at IS NULL;

	-- Revoke all un-accepted and un-declined workspace invites
	UPDATE workspace_invites
		SET deleted_at = NOW()
	WHERE workspace_id = OLD.id
		AND deleted_at IS NULL
		AND (accepted_at IS NULL AND declined_at IS NULL);
  RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER trg_workspace_soft_delete
AFTER UPDATE OF deleted_at ON workspaces
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL)
EXECUTE FUNCTION cascade_workspace_soft_delete();

