DO $$
DECLARE
    tbl_name TEXT;
BEGIN
    FOR tbl_name IN
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_type = 'BASE TABLE'
          AND table_name IN ('users', 'tags', 'workspaces', 'collections', 'collection_members', 'mfa_accounts', 'workspace_members')
    LOOP
		IF NOT EXISTS (
			SELECT 1
			FROM information_schema.triggers
			WHERE trigger_name = 'set_updated_at'
			AND event_object_table = tbl_name
		) THEN
			EXECUTE format('
				CREATE TRIGGER set_updated_at
				BEFORE UPDATE ON %I
				FOR EACH ROW
				EXECUTE FUNCTION update_timestamp();
			', tbl_name);
		END IF;
    END LOOP;
END;
$$ LANGUAGE 'plpgsql';
