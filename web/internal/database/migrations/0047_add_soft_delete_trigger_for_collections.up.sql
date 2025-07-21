create or replace function cascade_collection_soft_delete()
returns trigger
as
    $$
BEGIN
	-- Soft-delete all entries in the collection
	UPDATE entries
	SET deleted_at = NOW()
	WHERE collection_id = OLD.id AND deleted_at IS NULL;

	-- Soft-delete all entry chunks in the collection
	UPDATE entry_chunks
	SET deleted_at = NOW()
	WHERE entry_id IN (SELECT id FROM entries WHERE collection_id = OLD.id AND deleted_at IS NULL)
	AND deleted_at IS NULL;

	-- Soft-delete all collection members
	UPDATE collection_members
	SET deleted_at = NOW()
	WHERE collection_id = OLD.id AND deleted_at IS NULL;

	RETURN NEW;
END;
$$
language plpgsql
;

-- This trigger will be called after a collection is soft-deleted
CREATE TRIGGER trg_collection_soft_delete
AFTER UPDATE OF deleted_at ON collections
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL)
EXECUTE FUNCTION cascade_collection_soft_delete();

